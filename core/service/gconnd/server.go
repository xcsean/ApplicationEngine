package main

import (
	"fmt"
	"io"
	"net"
	"path"
	"strconv"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/packet"
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

func start(c *gconndConfig, id int64) {
	// save cfg & id
	config = c
	selfID = id

	// setup the main logger
	log.SetupMainLogger(path.Join(c.Log.Dir, c.Division), c.Log.FileName, c.Log.LogLevel)
	log.Info("------------------------------------>")
	log.Info("start with division=%s", c.Division)
	log.Info("getcd service addr=%s", c.GetcdAddr)
	log.Debug("server queue size=%d", c.SrvQueueSize)
	log.Debug("client queue size=%d", c.CliQueueSize)
	log.Debug("packet max bodyLen=%d", conn.LengthOfMaxBody)

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
	nodeIP, cliPort, srvPort, rpcPort, err := etc.SelectNode(c.Division)
	if err != nil {
		log.Fatal("server select node %s failed: %s", c.Division, err.Error())
	}

	// start query service & global config periodically
	etc.StartQueryServiceLoop(c.RefreshTime)
	etc.StartQueryGlobalConfigLoop(c.Categories, c.RefreshTime)
	//etc.StartReportWithAddr(config.Division, fmt.Sprintf("%s:%s", config.Mon.Ep.Ip, config.Mon.Ep.Port), config.Mon.ReportInterval)

	// create the channels for communication between server and client
	srvChannel := make(chan *innerCmd, c.SrvQueueSize)
	cliChannel := make(chan *innerCmd, c.CliQueueSize)
	rpcChannel := make(chan *reqRPC, c.SrvQueueSize)

	// create the maps for server and client connections
	srvMst = ""
	srvMap = make(map[string]srvSession)
	cliMap = make(map[uint64]*cliSession)

	// start the acceptors for server/client/rpc by using channels
	//  cli acceptor will use node:service_port in registry
	//  srv acceptor will use node:admin_port in registry
	//  rpc acceptor will use node:rpc_port in registry
	cliAddr := fmt.Sprintf("%s:%d", nodeIP, cliPort)
	srvAddr := fmt.Sprintf("%s:%d", nodeIP, srvPort)
	rpcAddr := fmt.Sprintf("%s:%d", nodeIP, rpcPort)
	ls, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		log.Fatal("RPC service listen faild: %s", err.Error())
	}
	ls2, err := net.Listen("tcp", cliAddr)
	if err != nil {
		log.Fatal("Cli listen failed: %s", err.Error())
	}
	ls3, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatal("Srv listen failed: %s", err.Error())
	}

	go startRPCLoop(ls, rpcChannel)
	go startCliLoop(ls2, cliChannel)
	go startSrvLoop(ls3, srvChannel)

	// start a profiler timer, print the performance information periodically
	tick := time.NewTicker(time.Duration(c.ProfilerTime) * time.Second)
	for {
		exit := false
		select {
		case cmd := <-cliChannel:
			exit = dispatchCliCmd(cmd, cliChannel)
		case cmd := <-srvChannel:
			exit = dispatchSrvCmd(cmd, srvChannel)
		case cmd := <-rpcChannel:
			exit = dispatchRPCCmd(cmd, rpcChannel)
		case <-tick.C:
			if config.isProfilerEnabled() {
				dispatchProfiler(cliChannel, srvChannel, rpcChannel)
			}
		}
		if exit {
			break
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
			pkt := packet.MakeNotifyClient(map[string]string{"result": fmt.Sprintf("%d", errno.CONNMASTEROFFLINE)})
			cliConn.Write(pkt)
			cliConn.Close()
		} else {
			count := len(cliMap)
			funGE := func(a, b int64) bool { return a >= b }
			if etc.CompareInt64WithConfig("global", "maxClientConnections", int64(count), int64(config.CliMaxConns), funGE) {
				log.Debug("client connections full, client num=%d", count)
				pkt := packet.MakeNotifyClient(map[string]string{"result": fmt.Sprintf("%d", errno.CONNMAXCONNECTIONS)})
				cliConn.Write(pkt)
				cliConn.Close()
			} else {
				sessionID := conn.MakeSessionID(uint16(selfID), seedID)
				seedID++

				// the forwardTo field set to master by default
				cs := &cliSession{cliConn: cliConn, forwardTo: srvMst}
				cliMap[sessionID] = cs
				log.Debug("client map size=%d", count+1)

				// send a session enter to master, with "client address" in body
				pkt, _ := packet.MakeSessionEnter([]uint64{sessionID}, cliAddr)
				_, err := srvConn.Write(pkt)
				if err != nil {
					log.Error("send session=%d enter failed: %s", sessionID, err.Error())
					cliConn.Close()
				} else {
					go recvCliLoop(cliConn, sessionID, srvMst, cliChannel)
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
			pkt, _ := packet.MakeSessionLeave([]uint64{sessionID})
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

		go recvSrvLoop(srvConn, srvChannel)
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
			pkt, _ := packet.MakeMasterNot()
			srvConn.Write(pkt)
		} else {
			srvMst = srvAddr
			log.Info("master=%s apply", srvMst)
			pkt, _ := packet.MakeMasterYou()
			srvConn.Write(pkt)
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

func dispatchRPCCmd(c *reqRPC, _ chan<- *reqRPC) bool {
	defer dbg.Stacktrace()

	exit := false
	switch c.Type {
	case innerCmdRPCAllocSessionID:
		n, _ := strconv.ParseInt(c.StrParam, 10, 64)
		if n <= 0 {
			n = 1
		}
		if n > 20 {
			n = 20
		}
		var ids []uint64
		for i := int64(0); i < n; i++ {
			sessionID := conn.MakeSessionID(uint16(selfID), seedID)
			seedID++
			ids = append(ids, sessionID)
		}
		s := ""
		l := len(ids)
		for i := 0; i < l; i++ {
			if i == 0 {
				s += fmt.Sprintf("%d", ids[i])
			} else {
				s += fmt.Sprintf(",%d", ids[i])
			}
		}
		c.Rsp <- &rspRPC{Result: errno.OK, StrParam: s}
	case innerCmdRPCIsSessionAlive:
		sessionID, _ := strconv.ParseUint(c.StrParam, 10, 64)
		_, ok := cliMap[sessionID]
		if ok {
			c.Rsp <- &rspRPC{Result: errno.OK, StrParam: "1"}
		} else {
			c.Rsp <- &rspRPC{Result: errno.OK, StrParam: "0"}
		}
	}
	return exit
}

func dispatchProfiler(cliChannel, srvChannel chan<- *innerCmd, rpcChannel chan<- *reqRPC) {
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
		if config.isTrafficEnabled() {
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
			if config.isTrafficEnabled() {
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
		if config.isTrafficEnabled() {
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
	cmdID := protocol.PacketType(header.CmdID)
	if cmdID == protocol.Packet_PRIVATE_SESSION_KICK {
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
	cmdID := protocol.PacketType(header.CmdID)
	if cmdID == protocol.Packet_PRIVATE_SESSION_ROUTE {
		sessionIDs, newForwardTo := packet.ParseSessionRouteBody(body)
		sessionNum := len(sessionIDs)
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
