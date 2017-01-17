package main

import (
	"net/http"
	. "utils"
)

var g_conf_mgmt *VooleConfigMgmt

func main() {
	var err error

	lock := make(chan int)
	Vlog_init("./Log.json")

	g_conf_mgmt := VooleConfigInit("server.json")
	if g_conf_mgmt == nil {
		VLOG(VLOG_ERROR, "Read config file %s error: ", err)
		goto exit
	}

	//start download
	down_mgmt := DownloadMgmtInit()
	down_mgmt.DownloadStart()

	//start web server
	server := HttpInit()
	server.HttpSetRouter("./seek", HttpSeekHandler)
	addr := conf.Host + ":" + conf.Port
	err = server.HttpStart(addr)
	if err != nil {
		goto exit
	}
	<-lock
exit:
	return
}

func HttpSeekHandler(w http.ResponseWriter, req *http.Request) {
	return
}
