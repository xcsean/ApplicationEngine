package main

import (
	"fmt"
	"net"
	"net/http"
	"path"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func start(c *getcdConfig, _ int64) bool {
	// save cfg
	config = c

	// setup the main logger
	log.SetupMainLogger(path.Join(c.Log.Dir, c.Division), c.Log.FileName, c.Log.LogLevel)
	log.Info("------------------------------------>")
	log.Info("start with division=%s", c.Division)
	log.Debug("log dir=%s, log name=%s, log level=%s", c.Log.Dir, c.Log.FileName, c.Log.LogLevel)
	log.Debug("rpc addr=%s, admin addr=%s", c.RPC.Addr(), c.Admin.Addr())
	log.Debug("mysql addr=%s, database=%s, username=%s, password=%s", c.Mysql.Addr(), c.Mysql.Database, c.Mysql.Username, c.Mysql.Password)
	log.Debug("mysql registry refresh=%d", c.Refresh)

	// start RPC service
	ch := make(chan string, 1)
	go startRPCLoop(c, ch)
	select {
	case s, _ := <-ch:
		if s == "ready" {
			log.Info("RPC service is ready now")
			break
		} else {
			log.Error("RPC service send %s", s)
			log.Info("should exit by RPC service exited or error occured")
			return false
		}
	}

	// provide admin service by http
	adminAddr := c.Admin.Addr()
	log.Info("admin start binding addr=%s", adminAddr)
	http.HandleFunc("/admin", adminWebHandler)
	admin := &http.Server{Addr: adminAddr, Handler: nil}
	err := admin.ListenAndServe()
	if err != nil {
		log.Fatal("admin listen failed: %s", err.Error())
		log.Info("admin exit")
		return false
	}

	log.Info("admin exit")
	return true
}

func startRPCLoop(c *getcdConfig, ch chan<- string) {
	s := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", c.Mysql.Username, c.Mysql.Password, c.Mysql.Addr(), c.Mysql.Database)

	// load registry from mysql immediately
	// if failed, just send error message and exit
	if err := loadRegistryFromMysql(s); err != nil {
		log.Error("load registry from mysql failed: %s", err.Error())
		ch <- "load from mysql failed"
		return
	}

	// start a timer, load registry from mysql periodically
	go loadRegistryPeriodically(s, c.Refresh)

	// listen and start work
	rpcAddr := c.RPC.Addr()
	ls, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		log.Error("RPC service listen failed: %s", err.Error())
		ch <- "listen port failed"
		return
	}
	defer ls.Close()

	log.Info("RPC service listen addr=%s ok", rpcAddr)
	ch <- "ready"

	srv := grpc.NewServer()
	protocol.RegisterGetcdServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)

	log.Info("RPC service exit")
	ch <- "exit"
	return
}
