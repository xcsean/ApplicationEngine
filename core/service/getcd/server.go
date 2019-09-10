package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"

	"github.com/xcsean/ApplicationEngine/core/protocol/getcd"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func start(fileName string) bool {
	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return false
	}

	err = xml.Unmarshal(fileData, &config)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return false
	}

	// set the log
	log.SetupMainLogger(path.Join(config.Log.Dir, config.Division), config.Log.FileName, config.Log.LogLevel)

	// print the server config
	log.Info("%s start with config=%s", config.Division, fileName)
	log.Debug("log dir=%s, log name=%s, log level=%s", config.Log.Dir, config.Log.FileName, config.Log.LogLevel)
	log.Debug("rpc addr=%s, admin addr=%s", config.RPC.Addr(), config.Admin.Addr())
	log.Debug("mysql addr=%s, database=%s, username=%s, password=%s", config.Mysql.Addr(), config.Mysql.Database, config.Mysql.Username, config.Mysql.Password)
	log.Debug("mysql registry refresh=%d", config.Refresh)

	// start RPC service
	ch := make(chan string, 1)
	go startRPC(config, ch)
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
	adminAddr := config.Admin.Addr()
	log.Info("admin start binding addr=%s", adminAddr)
	http.HandleFunc("/admin", adminWebHandler)
	admin := &http.Server{Addr: adminAddr, Handler: nil}
	err = admin.ListenAndServe()
	if err != nil {
		log.Fatal("admin listen failed: %s", err.Error())
		log.Info("admin exit")
		return false
	}

	log.Info("admin exit")
	return true
}

func startRPC(config getcdConfig, ch chan<- string) {
	s := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", config.Mysql.Username, config.Mysql.Password, config.Mysql.Addr(), config.Mysql.Database)

	// load registry from mysql immediately
	// if failed, just send error message and exit
	if err := loadRegistryFromMysql(s); err != nil {
		log.Error("load registry from mysql failed: %s", err.Error())
		ch <- "load from mysql failed"
		return
	}

	// start a timer, load registry from mysql periodically
	go loadRegistryPeriodically(s, config.Refresh)

	// listen and start work
	rpcAddr := config.RPC.Addr()
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
	getcd.RegisterGetcdServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)

	log.Info("RPC service exit")
	ch <- "exit"
	return
}
