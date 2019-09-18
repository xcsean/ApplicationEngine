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
		conn, err := ls.Accept()
		if err != nil {
			log.Error("accept client failed: %s", err.Error())
			continue
		}

		// notify a new client is incoming
		cliChannel <- newNotifyCmd(innerCmdClientIncoming, conn, "", 0, nil)
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
		conn, err := ls.Accept()
		if err != nil {
			log.Error("accept server failed: %s", err.Error())
			continue
		}

		// notify a new server is incoming
		srvChannel <- newNotifyCmd(innerCmdServerIncoming, conn, "", 0, nil)
	}

	// notify listen stop
	srvChannel <- newNotifyCmd(innerCmdServerListenStop, nil, "", 0, nil)
	log.Info("acceptor for server exit")
}

func acceptAdm(addr string, admChannel chan<- *innerCmd) {
	log.Info("start acceptor for admin: %s", addr)

	http.HandleFunc("/admin", func(rsp http.ResponseWriter, req *http.Request) {
		handleAdminRequest(rsp, req, admChannel)
	})

	admChannel <- newNotifyCmd(innerCmdAdminListenStart, nil, "", 0, nil)

	admin := &http.Server{Addr: addr, Handler: nil}
	err := admin.ListenAndServe()
	if err != nil {
		log.Error("admin ListenAndServe: %s", err.Error())
	}

	admChannel <- newNotifyCmd(innerCmdAdminListenStop, nil, "", 0, nil)
}
