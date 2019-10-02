package main

import (
	"fmt"
	"net"
	"path"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"github.com/xcsean/ApplicationEngine/core/shared/id"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func start(c *ghostConfig, selfID int64) bool {
	// save cfg
	config = c

	// setup the main logger
	log.SetupMainLogger(path.Join(c.Log.Dir, c.Division), c.Log.FileName, c.Log.LogLevel)
	log.Info("------------------------------------>")
	log.Info("start with division=%s", c.Division)
	log.Debug("log dir=%s, log name=%s, log level=%s", c.Log.Dir, c.Log.FileName, c.Log.LogLevel)
	log.Debug("mysql addr=%s, database=%s, username=%s, password=%s", c.Mysql.Addr(), c.Mysql.Database, c.Mysql.Username, c.Mysql.Password)
	log.Debug("getcd addr=%s", c.GetcdAddr)
	log.Debug("gconnd division=%s", c.Gconnd)

	// try to query the service & global config
	//  if failed, just print fatal log and exit
	etc.SetGetcdAddr(c.GetcdAddr)
	if err := etc.QueryService(); err != nil {
		log.Fatal("query service from %s failed: %s", c.GetcdAddr, err.Error())
	}
	log.Info("query service ok")
	if err := etc.QueryGlobalConfig(c.Categories); err != nil {
		log.Fatal("query global config from %s failed: %s", c.GetcdAddr, err.Error())
	}
	log.Info("query global config ok")

	// validate the host which could provide the service specified by division
	ok, err := etc.CanProvideService(c.Division)
	if err != nil {
		log.Fatal("server provides service %s failed: %s", c.Division, err.Error())
	}
	if !ok {
		log.Fatal("server can't provide service %s", c.Division)
	}

	// validate the division can be selected
	nodeIP, _, _, rpcPort, err := etc.SelectNode(c.Division)
	if err != nil {
		log.Fatal("server select node %s failed: %s", c.Division, err.Error())
	}

	// validate the gconnd specified is exist or not
	connIP, _, connPort, _, err := etc.SelectNode(c.Gconnd)
	if err != nil {
		log.Fatal("server select node %s failed: %s", c.Gconnd, err.Error())
	}

	// create the channels for communication between gconnd and vm(s)
	connChannel := make(chan *connCmd, 3000)
	rpcChannel := make(chan *innerCmd, 1000)
	vmmChannel := make(chan *innerCmd, 1000)
	tmmChannel := tmmGetChannel()

	// start the acceptor and client
	//  rpcAddr is the rpc address we should bind and provide service
	//  connAddr is the gconnd address which we should connect as a client
	rpcAddr := fmt.Sprintf("%s:%d", nodeIP, rpcPort)
	ls, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		log.Fatal("RPC service listen failed: %s", err.Error())
	}
	go startRPC(ls, rpcChannel)

	connAddr := fmt.Sprintf("%s:%d", connIP, connPort)
	csk, err := net.Dial("tcp", connAddr)
	if err != nil {
		log.Fatal("server can't connect %s", connAddr)
	}
	go startConn(csk, connChannel)

	// create the vm-manager, and add a global timer for the vmm
	settings := id.Settings{
		StartTime:      time.Now(),
		MachineID:      func() (uint16, error) { return uint16(selfID), nil },
		CheckMachineID: func(uint16) bool { return true },
	}
	newVMMgr(id.NewSnowflake(settings), vmmChannel)
	for {
		exit := false
		select {
		case cmd := <-connChannel:
			connExit := dispatchConn(cmd)
			if connExit {
				// TODO change the vmm and sm state
			}
		case cmd := <-rpcChannel:
			exit = dispatchRPC(cmd)
		case cmd := <-vmmChannel:
			exit = dispatchVMM(cmd)
		case cmd := <-tmmChannel:
			exit = dispatchTMM(cmd)
		}
		if exit {
			break
		}
	}

	// server exit
	return true
}

func startRPC(ls net.Listener, rpcChannel chan *innerCmd) {
	defer ls.Close()

	reqChannel = rpcChannel

	srv := grpc.NewServer()
	ghost.RegisterGhostServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)

	log.Info("RPC service exit")
}

func startConn(csk net.Conn, connChannel chan *connCmd) {
	defer csk.Close()

	// try to request master
	conn.SendMasterSet(csk)

	isMaster := false
	err := conn.HandleStream(csk, func(_ net.Conn, hdr, body []byte) {
		h := conn.ParseHeader(hdr)
		if isMaster {
			// common packet deal, push to connChannel
			dupHdr := make([]byte, len(hdr))
			dupBody := make([]byte, len(body))
			copy(dupHdr, hdr)
			copy(dupBody, body)
			select {
			case connChannel <- newConnCmd(innerCmdConnSessionUp, dupHdr, dupBody):
			default:
				// just discard the packet
			}
		} else {
			// wait the CmdMasterYou or CmdMasterNot
			switch h.CmdID {
			case conn.CmdMasterYou:
				log.Info("I'm master, that's ok")
				// init the conn-manager
				isMaster = true
				initConnMgr(csk)
			case conn.CmdMasterNot:
				log.Fatal("I can't be master, so exit")
			}
		}
	})

	if err != nil {
		log.Error("client for gconnd exit, reason: %s", err.Error())
	} else {
		log.Info("client for gconnd exit")
	}
	connChannel <- newConnCmd(innerCmdConnExit, nil, nil)
}
