package main

import (
	"net"
	"path"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func start(c *ghostConfig, _ int64) bool {
	// save cfg
	config = c

	// set the main logger
	log.SetupMainLogger(path.Join(c.Log.Dir, c.Division), c.Log.FileName, c.Log.LogLevel)
	log.Info("------------------------------------>")
	log.Info("start with division=%s", c.Division)
	log.Debug("log dir=%s, log name=%s, log level=%s", c.Log.Dir, c.Log.FileName, c.Log.LogLevel)
	log.Debug("mysql addr=%s, database=%s, username=%s, password=%s", c.Mysql.Addr(), c.Mysql.Database, c.Mysql.Username, c.Mysql.Password)

	// start RPC service
	ch := make(chan string, 1)
	startRPC("", ch)

	return true
}

func startRPC(rpcAddr string, ch chan<- string) {
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
	ghost.RegisterGhostServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)

	log.Info("RPC service exit")
	ch <- "exit"
	return
}
