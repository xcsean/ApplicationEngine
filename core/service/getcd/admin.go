package main

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

type adminWebRsp struct {
	Result  uint   `json:"result"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func adminWebHandler(w http.ResponseWriter, r *http.Request) {
	defer dbg.Stacktrace()

	var result adminWebRsp
	w.Header().Add("Access-Control-Allow-Origin", "*")
	rip := r.Header.Get("X-Forwarded-For")
	if rip == "" {
		rip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	/*
	adminAddrMap := getAdminFromMap()
	if state, ok := adminAddrMap[rip]; !ok || state != 0 {
		result.Result = 1
		result.Error = "access denied"
		rsp, _ := json.Marshal(result)
		log.Info("access denied, remote ip: %s", rip)
		w.Write(rsp)
		return
	}*/

	r.ParseForm()
	log.Debug("http request form: %v", r.Form)
	op := r.FormValue("op")
	switch op {
	case "query":
		log.Info("query the registry")
		info1 := getRegistryServerString()
		info2 := getRegistryServiceString()
		info1 += " | "
		info1 += info2
		result.Result = 0
		result.Message = info1
	case "log":
		newLevel := log.ChangeLevel(r.FormValue("level"))
		log.Info("log level set to %s", newLevel)
		result.Result = 0
		result.Message = "log level set ok"
	case "":
	default:
		result.Result = 1
		result.Error = "invalid operation"
	}

	rsp, _ := json.Marshal(result)
	w.Write(rsp)
}
