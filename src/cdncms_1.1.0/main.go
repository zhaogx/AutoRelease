package main

import (
	_ "fmt"
	"os"
	_ "os/exec"
	. "utils"
)

func init_dir(conf *ServerConfig) bool {
	if conf == nil {
		return false
	}
	gconf := &conf.Global

	var dir string
	var err error
	dir = gconf.WebServerBaseDir + "/download"
	err = os.MkdirAll(dir, 077)
	if err != nil {
		return false
	}

	dir = gconf.ShotServerBaseDir + "/mp4"
	err = os.MkdirAll(dir, 077)
	if err != nil {
		return false
	}

	dir = gconf.ShotServerBaseDir + "/shot"
	err = os.MkdirAll(dir, 077)
	if err != nil {
		return false
	}

	dir = gconf.ShotServerBaseDir + "/back"
	err = os.MkdirAll(dir, 077)
	if err != nil {
		return false
	}
	return true
}

func main() {
	var conf ServerConfig

	lock := make(chan int)
	defer close(lock)

	log_conf_file := "./Log.json"
	server_conf_file := "./server.json"

	Vlog_init(log_conf_file)

	VLOG(VLOG_ALTER, "========================================================================\n")

	err := ReadConfig(server_conf_file, &conf)
	if err != nil {
		VLOG(VLOG_ERROR, "ReadConfig failed!!![%s][%s]", server_conf_file, err.Error())
		return
	}
	VLOG_LINE(VLOG_ALTER, conf)

	f := init_dir(&conf)
	if f == false {
		VLOG(VLOG_ERROR, "init dir failed !!!")
	}
	//start web server
	go CdncmsWebInit(&conf)
	//start download
	go DownloadMgmtInit(&conf)
	//shot server
	go CdncmsShotMgmtInit(&conf)

	<-lock
	return
}
