package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	. "utils"
)

var conf CdncmsConfig

func main() {
	var err error

	lock := make(chan int)
	Vlog_init("./Log.json")

	err = ReadConfig("server.json", conf)
	if err != nil {
		VLOG(VLOG_ERROR, "Read config file %s error: ", err)
		goto exit
	}

	//start download
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
