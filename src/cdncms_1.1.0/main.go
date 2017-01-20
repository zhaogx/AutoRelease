package main

import (
	. "utils"
)

func main() {
	var conf ServerConfig

	lock := make(chan int)
	defer close(lock)

	Vlog_init("./Log.json")

	ReadConfig("server.json", &conf)
	VLOG_LINE(VLOG_ERROR, conf)

	//start web server
	go CdncmsWebInit(&conf)
	//start download
	go DownloadMgmtInit(&conf)

	<-lock
	return
}
