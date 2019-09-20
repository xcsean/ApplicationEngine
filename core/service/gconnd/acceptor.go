package main

import (
	"net"
	"net/http"

	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func acceptCli(addr string, cliChannel chan<- *innerCmd) {
	defer dbg.Stacktrace()

	log.Info("start acceptor for client: %s", addr)
	cliAddr := addr
	ls, err := net.Listen("tcp", cliAddr)
	if err != nil {
		cliChannel <- newNotifyCmd(innerCmdClientListenStart, nil, "", 0, err)
		return
	}
	defer ls.Close()

	// notify listen start
	cliChannel <- newNotifyCmd(innerCmdClientListenStart, nil, "", 0, nil)

	for {
		c, err := ls.Accept()
		if err != nil {
			log.Error("accept client failed: %s", err.Error())
			continue
		}

		// notify a new client is incoming
		cliChannel <- newNotifyCmd(innerCmdClientIncoming, c, "", 0, nil)
	}

	// notify listen stop
	cliChannel <- newNotifyCmd(innerCmdClientListenStop, nil, "", 0, nil)
	log.Info("acceptor for client exit")
}

func acceptSrv(addr string, srvChannel chan<- *innerCmd) {
	defer dbg.Stacktrace()

	log.Info("start acceptor for server: %s", addr)
	srvAddr := addr
	ls, err := net.Listen("tcp", srvAddr)
	if err != nil {
		srvChannel <- newNotifyCmd(innerCmdServerListenStart, nil, "", 0, err)
		return
	}
	defer ls.Close()

	// notify listen start
	srvChannel <- newNotifyCmd(innerCmdServerListenStart, nil, "", 0, nil)

	for {
		c, err := ls.Accept()
		if err != nil {
			log.Error("accept server failed: %s", err.Error())
			continue
		}

		// notify a new server is incoming
		srvChannel <- newNotifyCmd(innerCmdServerIncoming, c, "", 0, nil)
	}

	// notify listen stop
	srvChannel <- newNotifyCmd(innerCmdServerListenStop, nil, "", 0, nil)
	log.Info("acceptor for server exit")
}

func acceptRPC(addr string, rpcChannel chan<- *innerCmd) {
	log.Info("start acceptor for rpc: %s", addr)

	http.HandleFunc("/rpc", func(rsp http.ResponseWriter, req *http.Request) {
		handleAdminRequest(rsp, req, rpcChannel)
	})

	rpcChannel <- newNotifyCmd(innerCmdAdminListenStart, nil, "", 0, nil)

	rpc := &http.Server{Addr: addr, Handler: nil}
	err := rpc.ListenAndServe()
	if err != nil {
		log.Error("rpc ListenAndServe: %s", err.Error())
	}

	rpcChannel <- newNotifyCmd(innerCmdAdminListenStop, nil, "", 0, nil)
}
