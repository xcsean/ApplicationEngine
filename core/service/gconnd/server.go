package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	cnn "github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
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
	log.Info("getcnn.service addr=%s", cfg.GetcdAddr)
	log.Debug("server queue size=%d", cfg.SrvQueueSize)
	log.Debug("client queue size=%d", cfg.CliQueueSize)
	log.Debug("packet max bodyLen=%d", cnn.LengthOfMaxBody)

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
	nodeIP, svcPort, admPort, rpcPort, err := etc.SelectNode(cfg.Division)
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
	admChannel := make(chan *innerCmd, 10)

	// create the maps for server and client connections
	srvMst = ""
	srvMap = make(map[string]srvSession)
	cliMap = make(map[uint64]*cliSession)

	// start the acceptors for server/client/admin by using channels
	//  client acceptor will use node:service_port
	//  server acceptor will use 127.0.0.1:rpc_port
	//  admin  acceptor will use node:admin_port
	cliAddr := fmt.Sprintf("%s:%d", nodeIP, svcPort)
	srvAddr := fmt.Sprintf("%s:%d", "127.0.0.1", rpcPort)
	admAddr := fmt.Sprintf("%s:%d", nodeIP, admPort)
	go acceptCli(cliAddr, cliChannel)
	go acceptSrv(srvAddr, srvChannel)
	go acceptAdm(admAddr, admChannel)

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
		case c := <-admChannel:
			dispatchAdmCmd(c, admChannel)
		case <-tick.C:
			if config.IsProfilerEnabled() {
				dispatchProfiler(cliChannel, srvChannel, admChannel)
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
			// TODO add a rsp to client
			log.Debug("master is nil, so close the connection from client=%s", cliAddr)
			cliConn.Close()
		} else {
			count := len(cliMap)
			funGE := func(a, b int64) bool {return a >= b}
			if etc.CompareInt64WithConfig("global", "maxClientConnections", int64(count), int64(config.CliMaxConns), funGE) {
				// TODO add a rsp to client
				log.Debug("client connections full, client num=%d", count)
				cliConn.Close()
			} else {
				sessionID := cnn.MakeSessionID(uint16(selfID), seedID)
				seedID++

				// the forwardTo field set to master by default
				cs := &cliSession{cliConn: cliConn, forwardTo: srvMst, }
				cliMap[sessionID] = cs
				log.Debug("client map size=%d", count + 1)

				// send a session enter to master
				sessions := make([]uint64, 1)
				sessions[0] = sessionID
				pkt, _ := cnn.MakeSessionPkt(sessions, cnn.CmdSessionEnter, 0, 0, nil)
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

		// send a session leave to master
		srvConn, ok := srvMap[srvMst]
		if ok {
			sessions := make([]uint64, 1)
			sessions[0] = sessionID
			pkt, _ := cnn.MakeSessionPkt(sessions, cnn.CmdSessionLeave, 0, 0, nil)
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
			cnn.SendMasterNot(srvConn)
		} else {
			srvMst = srvAddr
			log.Info("master=%s apply", srvMst)
			cnn.SendMasterYou(srvConn)
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

func dispatchAdmCmd(c *innerCmd, _ chan<- *innerCmd) {
	defer dbg.Stacktrace()

	cmdID := c.getCmdID()
	switch cmdID {
	case innerCmdAdminListenStart:
		log.Debug("start listen admin ok")
	case innerCmdAdminListenStop:
		log.Debug("stop listen admin ok")
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
}

func dispatchProfiler(client, server, admin chan<- *innerCmd) {
	defer dbg.Stacktrace()

	log.Debug("client cmd queue size=%d", len(client))
	log.Debug("server cmd queue size=%d", len(server))
	log.Debug("admin  cmd queue size=%d", len(admin))
}

func forwardToServer(srvConn io.Writer, sessionID uint64, hdr, body []byte) {
	sessions := make([]uint64, 1)
	sessions[0] = sessionID
	pkt, cmdID := cnn.CopySessionPkt(sessions, hdr, body)
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
	sessionNum, sessions, newBody := cnn.ParseSessionBody(body)
	log.Debug("broadcast %d client(s)", sessionNum)
	pkt, cmdID := cnn.CopyCommonPkt(hdr, newBody)
	for i := 0; i < int(sessionNum); i++ {
		sessionID := sessions[i]
		cs, ok := cliMap[sessionID]
		if ok {
			n, err := cs.cliConn.Write(pkt)
			if config.IsTrafficEnabled() {
				log.Info("DN|session=%d|cmd=%d|hdr=%d|body=%d", sessionID, cmdID, len(hdr), len(newBody))
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
			hdr := cnn.ParseHeader(pkt)
			log.Info("DN|session=%d|cmd=%d|hdr=%d|body=%d", sessionID, hdr.CmdID, cnn.LengthOfHeader, len(pkt)-cnn.LengthOfHeader)
			if err != nil {
				log.Error("DNFORWARD|error=%s", err.Error())
			} else {
				log.Info("DNFORWARD|bytes=%d", n)
			}
		}
	}
}

func kickClients(hdr, body []byte) {
	header := cnn.ParseHeader(hdr)
	if header.CmdID == cnn.CmdSessionKick {
		sessionNum, sessions, _ := cnn.ParseSessionBody(body)
		log.Debug("kick %d client(s)", sessionNum)
		for i := 0; i < int(sessionNum); i++ {
			sessionID := sessions[i]
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
	header := cnn.ParseHeader(hdr)
	if header.CmdID == cnn.CmdSessionRoute {
		sessionNum, sessions, newBody := cnn.ParseSessionBody(body)

		var pb cnn.PrivateBody
		err := json.Unmarshal(newBody, &pb)
		if err != nil {
			log.Error("set route unmarshal failed: %s", err.Error())
			return
		}
		log.Debug("private body=%v", pb)

		newTarget := pb.StrParam
		_, ok := srvMap[newTarget]
		if ok {
			for i := 0; i < int(sessionNum); i++ {
				sessionID := sessions[i]
				cs, ok := cliMap[sessionID]
				if ok {
					log.Info("client set route, session=%d from %s to %s", sessionID, cs.forwardTo, newTarget)
					cs.forwardTo = newTarget
				} else {
					log.Debug("client set route, session=%d not found", sessionID)
				}
			}
		}
	}
}
