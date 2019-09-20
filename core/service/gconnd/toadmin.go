package main

import (
	"encoding/json"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

type adminInfo struct {
	result  uint   `json:"result"`
	errors  string `json:"error"`
	message string `json:"message"`
}

func handleAdminRequest(rsp http.ResponseWriter, req *http.Request, ch chan<- *innerCmd) {
	defer dbg.Stacktrace()

	var info adminInfo
	rsp.Header().Add("Access-Control-Allow-Origin", "*")

	// get the ip address of remote
	rip := req.Header.Get("X-Forwarded-For")
	if rip == "" {
		rip, _, _ = net.SplitHostPort(req.RemoteAddr)
	}

	// validate the ip address in "ip admin list" or not
	if !etc.InConfig("permission", "ipAdminList", rip) {
		info.result = 1
		info.errors = "invalid admin ip address"
		body, _ := json.Marshal(info)
		_, err := rsp.Write(body)
		if err != nil {
			log.Error("%s", err.Error())
		}
		return
	}

	if req.Method != "POST" {
		log.Debug("error Method=%v", req.Method)

		info.result = 1
		info.errors = "should use POST method"
		body, _ := json.Marshal(info)
		_, err := rsp.Write(body)
		if err != nil {
			log.Error("%s", err.Error())
		}
	} else {
		err := req.ParseForm()
		if err != nil {
			log.Error("%s", err.Error())
			return
		}

		op := req.FormValue("op")
		switch op {
		case "":
			info.result = 1
			info.errors = "should contain field 'op'"
		case "kick":
			s := req.FormValue("session")
			var i int64
			i, err = strconv.ParseInt(s, 10, 64)
			if err != nil || i == 0 {
				info.result = 1
				info.errors = err.Error()
			} else {
				ch <- newAdminCmd(innerCmdAdminKick, uint64(i))
				info.result = 0
				info.errors = ""
				info.message = "ok"
			}
		case "kickall":
			ch <- newAdminCmd(innerCmdAdminKickAll, 0)
			info.result = 0
			info.message = "ok"
		case "log":
			level := req.FormValue("level")
			s := log.ChangeLevel(level)
			log.Info("log level set to %s", s)
			info.result = 0
			info.message = s
		}

		body, _ := json.Marshal(info)
		_, err = rsp.Write(body)
		if err != nil {
			log.Error("%s", err.Error())
		}
	}
}
