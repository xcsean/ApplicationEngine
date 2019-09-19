package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

type srvSession net.Conn

type cliSession struct {
	cliConn   net.Conn
	forwardTo string
}

var (
	selfID int64
	seedID uint16
	srvMst string
	srvMap map[string]srvSession
	cliMap map[uint64]*cliSession
)

func start(cfg *gconndConfig, id int64) {
	// save cfg & id
	config = cfg
	selfID = id

	// setup the main logger
	log.SetupMainLogger(cfg.Log.Dir, cfg.Log.FileName, cfg.Log.LogLevel)
	log.Info("------------------------------------>")
	log.Info("start with division=%s", cfg.Division)
	log.Info("getcd service addr=%s", cfg.GetcdAddr)
	log.Debug("server queue size=%d", cfg.SrvQueueSize)
	log.Debug("client queue size=%d", cfg.CliQueueSize)
	log.Debug("packet max bodyLen=%d", conn.LengthOfMaxBody)

	// try to query the service & global config
	//  if failed, just print fatal log and exit
	etc.SetGetcdAddr(cfg.GetcdAddr)
	if err := etc.QueryService(); err != nil {
		log.Fatal("query service from %s failed: %s", cfg.GetcdAddr, err.Error())
	}
	log.Info("query service ok")
	if err := etc.QueryGlobalConfig(cfg.Categories); err != nil {
		log.Fatal("query global config from %s failed: %s", cfg.GetcdAddr, err.Error())
	}
	log.Info("query global config ok")

	// validate the host which could provide the service specified by division
	ok, err := etc.CanProvideService(cfg.Division)
	if err != nil {
		log.Fatal("server provides service %s failed: %s", cfg.Division,  err.Error())
	}
	if !ok {
		log.Fatal("server can't provide service %s", cfg.Division)
	}

	// validate the division can be selected
	nodeIP, cliPort, srvPort, rpcPort, err := etc.SelectNode(cfg.Division)
	if err != nil {
		log.Fatal("server select node %s failed: %s", cfg.Division, err.Error())
	}

	// start query service & global config periodically
	etc.StartQueryServiceLoop(cfg.RefreshTime)
	etc.StartQueryGlobalConfigLoop(cfg.Categories, cfg.RefreshTime)
	//etc.StartReportWithAddr(config.Division, fmt.Sprintf("%s:%s", config.Mon.Ep.Ip, config.Mon.Ep.Port), config.Mon.ReportInterval)

	// create the channels for communication between server and client
	srvChannel := make(chan *innerCmd, cfg.SrvQueueSize)
	cliChannel := make(chan *innerCmd, cfg.CliQueueSize)
	rpcChannel := make(chan *innerCmd, 10)

	// create the maps for server and client connections
	srvMst = ""
	srvMap = make(map[string]srvSession)
	cliMap = make(map[uint64]*cliSession)

	// start the acceptors for server/client/rpc by using channels
	//  cli acceptor will use node:service_port in registry
	//  srv acceptor will use 127.0.0.1:admin_port in registry
	//  rpc acceptor will use node:rpc_port in registry
	// TODO: rpc change from http to rpc
	cliAddr := fmt.Sprintf("%s:%d", nodeIP, cliPort)
	srvAddr := fmt.Sprintf("%s:%d", "127.0.0.1", srvPort)
	rpcAddr := fmt.Sprintf("%s:%d", nodeIP, rpcPort)
	go acceptCli(cliAddr, cliChannel)
	go acceptSrv(srvAddr, srvChannel)
	go acceptRPC(rpcAddr, rpcChannel)

	// start a profiler timer, print the performance information periodically
	tick := time.NewTicker(time.Duration(cfg.ProfilerTime) * time.Second)
	for {
		select {
		case c := <-cliChannel:
			exit := dispatchCliCmd(c, cliChannel)
			if exit {
				break
			}
		case c := <-srvChannel:
			exit := dispatchSrvCmd(c, srvChannel)
			if exit {
				break
			}
		case c := <-rpcChannel:
			exit := dispatchRPCCmd(c, rpcChannel)
			if exit {
				break
			}
		case <-tick.C:
			if config.IsProfilerEnabled() {
				dispatchProfiler(cliChannel, srvChannel, rpcChannel)
			}
		}
	}

	// server exit
	log.Info("server exit")
}

func dispatchCliCmd(c *innerCmd, cliChannel chan<- *innerCmd) bool {
	defer dbg.Stacktrace()

	exit := false
	cmdID := c.getCmdID()
	switch cmdID {
	case innerCmdClientListenStart:
		_, _, _, err := c.getNotifyCmd()
		if err != nil {
			log.Error("start listen client failed: %s", err.Error())
			exit = true
		} else {
			log.Info("start listen client ok")
		}
	case innerCmdClientListenStop:
		exit = true
	case innerCmdClientIncoming:
		cliConn, _, _, _ := c.getNotifyCmd()
		cliAddr := cliConn.RemoteAddr().String()
		log.Debug("client incoming: conn=%v, remote=%s, master=%s", cliConn, cliAddr, srvMst)
		srvConn, ok := srvMap[srvMst]
		if !ok {
			// master isn't ready, so kick the client
			log.Debug("master is nil, so close the connection from client=%s", cliAddr)
			conn.SendNotifyToClient(cliConn, errno.CONNMASTEROFFLINE)
			cliConn.Close()
		} else {
			count := len(cliMap)
			funGE := func(a, b int64) bool {return a >= b}
			if etc.CompareInt64WithConfig("global", "maxClientConnections", int64(count), int64(config.CliMaxConns), funGE) {
				log.Debug("client connections full, client num=%d", count)
				conn.SendNotifyToClient(cliConn, errno.CONNMAXCONNECTIONS)
				cliConn.Close()
			} else {
				sessionID := conn.MakeSessionID(uint16(selfID), seedID)
				seedID++

				// the forwardTo field set to master by default
				cs := &cliSession{cliConn: cliConn, forwardTo: srvMst, }
				cliMap[sessionID] = cs
				log.Debug("client map size=%d", count + 1)

				// send a session enter to master, with "client address" in body
				sessionIDs := make([]uint64, 1)
				sessionIDs[0] = sessionID

				pb := &conn.PrivateBody{StrParam: cliAddr, }
				body, _ := json.Marshal(pb)

				pkt, _ := conn.MakeSessionPkt(sessionIDs, conn.CmdSessionEnter, 0, 0, body)
				_, err := srvConn.Write(pkt)
				if err != nil {
					log.Error("send session=%d enter failed: %s", sessionID, err.Error())
					cliConn.Close()
				} else {
					go handleClientConn(cliConn, sessionID, srvMst, cliChannel)
				}
			}
		}
	case innerCmdClientLeave:
		_, cliAddr, sessionID, err := c.getNotifyCmd()
		log.Debug("client leave: remote=%s, session=%d", cliAddr, sessionID)
		if err != nil {
			log.Error("client leave: reason=%s", err.Error())
		}

		// send a session leave to master, with "client address" in body
		srvConn, ok := srvMap[srvMst]
		if ok {
			sessionIDs := make([]uint64, 1)
			sessionIDs[0] = sessionID

			pb := &conn.PrivateBody{StrParam: cliAddr, }
			body, _ := json.Marshal(pb)

			pkt, _ := conn.MakeSessionPkt(sessionIDs, conn.CmdSessionLeave, 0, 0, body)
			_, err := srvConn.Write(pkt)
			if err != nil {
				log.Error("send session=%d leave failed: %s", sessionID, err.Error())
			}
		}

		// delete the client
		delete(cliMap, sessionID)
		log.Debug("client map size=%d", len(cliMap))
	case innerCmdClientUp:
		sessionID, hdr, body := c.getClientCmd()
		cs, ok := cliMap[sessionID]
		if ok {
			log.Debug("client up: server=%s", cs.forwardTo)
			srvConn, ok := srvMap[cs.forwardTo]
			if ok {
				log.Debug("server=%s found, conn=%v", cs.forwardTo, srvConn)
				forwardToServer(srvConn, sessionID, hdr, body)
			} else {
				log.Debug("server=%s not found, discard client up", cs.forwardTo)
			}
		} else {
			log.Debug("client up, but session=%d not found", sessionID)
		}
	}
	return exit
}

func dispatchSrvCmd(c *innerCmd, srvChannel chan<- *innerCmd) bool {
	defer dbg.Stacktrace()

	exit := false
	cmdID := c.getCmdID()
	switch cmdID {
	case innerCmdServerListenStart:
		_, _, _, err := c.getNotifyCmd()
		if err != nil {
			log.Error("start listen server failed: %s", err.Error())
			exit = true
		} else {
			log.Info("start listen server ok")
		}
	case innerCmdServerListenStop:
		log.Info("stop listen server")
		exit = true
	case innerCmdServerIncoming:
		srvConn, _, _, _ := c.getNotifyCmd()
		srvAddr := srvConn.RemoteAddr().String()
		log.Debug("server incoming: conn=%v, remote=%s", srvConn, srvAddr)

		srvMap[srvAddr] = srvConn
		log.Debug("server map size=%d", len(srvMap))

		go handleServerConn(srvConn, srvChannel)
	case innerCmdServerLeave:
		_, srvAddr, _, err := c.getNotifyCmd()
		log.Debug("server leave: remote=%s", srvAddr)
		if err != nil {
			log.Error("server leave: reason=%s", err.Error())
		}

		kickAll := false
		if srvMst == srvAddr {
			srvMst = ""
			log.Debug("master=%s leave", srvAddr)
			kickAll = true
		}
		log.Info("master=%s", srvMst)

		delete(srvMap, srvAddr)
		log.Debug("server map size=%d", len(srvMap))

		if kickAll {
			log.Debug("master leave, should kick all clients")
			kickAllClients()
		}
	case innerCmdServerMasterSet:
		srvConn, _, _ := c.getServerCmd()
		srvAddr := srvConn.RemoteAddr().String()
		log.Debug("server=%s want to be master", srvAddr)
		_, ok := srvMap[srvMst]
		if ok {
			log.Debug("master=%s exist, refuse the request from %s", srvMst, srvAddr)
			conn.SendMasterNot(srvConn)
		} else {
			srvMst = srvAddr
			log.Info("master=%s apply", srvMst)
			conn.SendMasterYou(srvConn)
		}
	case innerCmdServerBroadcast:
		srvConn, hdr, body := c.getServerCmd()
		srvAddr := srvConn.RemoteAddr().String()
		log.Debug("server=%s broadcast", srvAddr)
		broadcastToClients(hdr, body)
	case innerCmdServerBroadcastAll:
		srvConn, _, body := c.getServerCmd()
		srvAddr := srvConn.RemoteAddr().String()
		log.Info("server=%s broadcast all", srvAddr)
		broadcastAll(body)
	case innerCmdServerKick:
		srvConn, hdr, body := c.getServerCmd()
		srvAddr := srvConn.RemoteAddr().String()
		log.Info("server=%s request kick", srvAddr)
		kickClients(hdr, body)
	case innerCmdServerKickAll:
		srvConn, _, _ := c.getServerCmd()
		srvAddr := srvConn.RemoteAddr().String()
		log.Info("server=%s request kick all clients", srvAddr)
		kickAllClients()
	case innerCmdServerSetRoute:
		_, hdr, body := c.getServerCmd()
		setRouteForClients(hdr, body)
	}
	return exit
}

func dispatchRPCCmd(c *innerCmd, _ chan<- *innerCmd) bool {
	defer dbg.Stacktrace()

	exit := false
	cmdID := c.getCmdID()
	switch cmdID {
	case innerCmdAdminListenStart:
		log.Info("start listen admin ok")
	case innerCmdAdminListenStop:
		log.Info("stop listen admin")
		exit = true
	case innerCmdAdminKick:
		sessionID := c.getAdminCmd()
		log.Debug("admin kick session=%d", sessionID)
		cs, ok := cliMap[sessionID]
		if ok {
			log.Debug("client session=%d kicked by admin", sessionID)
			cs.cliConn.Close()
		} else {
			log.Debug("admin kick session=%d, but not found", sessionID)
		}
	case innerCmdAdminKickAll:
		log.Info("admin kick all")
		kickAllClients()
	}
	return exit
}

func dispatchProfiler(cliChannel, srvChannel, rpcChannel chan<- *innerCmd) {
	defer dbg.Stacktrace()

	log.Debug("cliChannel cmd queue size=%d", len(cliChannel))
	log.Debug("srvChannel cmd queue size=%d", len(srvChannel))
	log.Debug("rpcChannel cmd queue size=%d", len(rpcChannel))
}

func forwardToServer(srvConn io.Writer, sessionID uint64, hdr, body []byte) {
	sessionIDs := make([]uint64, 1)
	sessionIDs[0] = sessionID
	pkt, cmdID := conn.CopySessionPkt(sessionIDs, hdr, body)
	if pkt != nil {
		n, err := srvConn.Write(pkt)
		if config.IsTrafficEnabled() {
			log.Info("UP|session=%d|cmd=%d|hdr=%d|body=%d", sessionID, cmdID, len(hdr), len(body))
			if err != nil {
				log.Error("UPFORWARD|error=%s", err.Error())
			} else {
				log.Info("UPFORWARD|bytes=%d", n)
			}
		}
	}
}

func broadcastToClients(hdr, body []byte) {
	sessionNum, sessionIDs, innerBody := conn.ParseSessionBody(body)
	log.Debug("broadcast %d client(s)", sessionNum)
	pkt, cmdID := conn.CopyCommonPkt(hdr, innerBody)
	for i := 0; i < int(sessionNum); i++ {
		sessionID := sessionIDs[i]
		cs, ok := cliMap[sessionID]
		if ok {
			n, err := cs.cliConn.Write(pkt)
			if config.IsTrafficEnabled() {
				log.Info("DN|session=%d|cmd=%d|hdr=%d|body=%d", sessionID, cmdID, len(hdr), len(innerBody))
				if err != nil {
					log.Error("DNFORWARD|error=%s", err.Error())
				} else {
					log.Info("DNFORWARD|bytes=%d", n)
				}
			}
		} else {
			log.Debug("client dn, but session=%d not found", sessionID)
		}
	}
}

func broadcastAll(pkt []byte) {
	defer dbg.Stacktrace()

	for sessionID, cs := range cliMap {
		log.Debug("broadcast all, session=%d, pkt=%d", sessionID, len(pkt))
		n, err := cs.cliConn.Write(pkt)
		if config.IsTrafficEnabled() {
			hdr := conn.ParseHeader(pkt)
			log.Info("DN|session=%d|cmd=%d|hdr=%d|body=%d", sessionID, hdr.CmdID, conn.LengthOfHeader, len(pkt)-conn.LengthOfHeader)
			if err != nil {
				log.Error("DNFORWARD|error=%s", err.Error())
			} else {
				log.Info("DNFORWARD|bytes=%d", n)
			}
		}
	}
}

func kickClients(hdr, body []byte) {
	header := conn.ParseHeader(hdr)
	if header.CmdID == conn.CmdSessionKick {
		sessionNum, sessionIDs, _ := conn.ParseSessionBody(body)
		log.Debug("kick %d client(s)", sessionNum)
		for i := 0; i < int(sessionNum); i++ {
			sessionID := sessionIDs[i]
			cs, ok := cliMap[sessionID]
			if ok {
				log.Debug("client kicked, session=%d", sessionID)
				cs.cliConn.Close()
			} else {
				log.Debug("client kick failed, session=%d not found", sessionID)
			}
		}
	}
}

func kickAllClients() {
	for sessionID, cs := range cliMap {
		log.Debug("client kicked, session=%d", sessionID)
		cs.cliConn.Close()
	}
}

func setRouteForClients(hdr, body []byte) {
	header := conn.ParseHeader(hdr)
	if header.CmdID == conn.CmdSessionRoute {
		sessionNum, sessionIDs, innerBody := conn.ParseSessionBody(body)

		var pb conn.PrivateBody
		err := json.Unmarshal(innerBody, &pb)
		if err != nil {
			log.Error("set route unmarshal failed: %s", err.Error())
			return
		}
		log.Debug("private body=%v", pb)

		newForwardTo := pb.StrParam
		_, ok := srvMap[newForwardTo]
		if ok {
			for i := 0; i < int(sessionNum); i++ {
				sessionID := sessionIDs[i]
				cs, ok := cliMap[sessionID]
				if ok {
					log.Info("client set route, session=%d from %s to %s", sessionID, cs.forwardTo, newForwardTo)
					cs.forwardTo = newForwardTo
				} else {
					log.Debug("client set route, session=%d not found", sessionID)
				}
			}
		}
	}
}
