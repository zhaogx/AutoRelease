package main

import (
	"os"
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

	dir = gconf.UploadServerBaseDir + "/waitting"
	err = os.MkdirAll(dir, 077)
	if err != nil {
		return false
	}
	dir = gconf.UploadServerBaseDir + "/down"
	err = os.MkdirAll(dir, 077)
	if err != nil {
		return false
	}
	return true
}

func main() {
	var conf ServerConfig

	if false == deamon() {
		return
	}

	//lock := make(chan int)
	//defer close(lock)

	log_conf_file := "./Log.json"
	server_conf_file := "./server.json"

	Vlog_init(log_conf_file)

	VLOG(VLOG_ALTER, "\n=========================== Start =============================================\n")

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

	ret_web := make(chan bool)
	defer close(ret_web)
	ret_down := make(chan bool)
	defer close(ret_down)
	ret_shot := make(chan bool)
	defer close(ret_shot)
	ret_up := make(chan bool)
	defer close(ret_up)

	//start web server
	go CdncmsWebInit(&conf, ret_web)
	//start download
	go DownloadMgmtInit(&conf, ret_down)
	//shot server
	go CdncmsShotMgmtInit(&conf, ret_shot)
	//upload server
	go CdncmsUploadMgmtInit(&conf, ret_up)

	select {
	case b := <-ret_web:
		if b == false {
			VLOG(VLOG_ERROR, "WebInit Failed!!!")
			return
		}
	case b := <-ret_down:
		if b == false {
			VLOG(VLOG_ERROR, "DownloadMgmtInit Failed!!!")
			return
		}
	case b := <-ret_shot:
		if b == false {
			VLOG(VLOG_ERROR, "ShotMgmtInit Failed!!!")
			return
		}
	case b := <-ret_up:
		if b == false {
			VLOG(VLOG_ERROR, "UploadMgmtInit Failed!!!")
			return
		}
	}
	//<-lock
	return
}
