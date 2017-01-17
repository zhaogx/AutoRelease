package main

import (
	"fmt"
	. "utils"
)

type DownLoadConfig struct {
	Host        string
	Port        uint32
	VendownName string
	ListenPath  string
	SavePath    string
	Errcbhost   string
	Errcbport   string
	Threadnum   uint32

	Mysql struct {
		Host     string
		Port     uint32
		User     string
		Password string
		Name     string
	}
}

func GetDownloadConfig(mgmt *VooleConfigMgmt) *DownLoadConfig {
	var ok bool
	if mgmt == nil {
		return nil
	}
	conf := new(DownLoadConfig)
	conf.Host, ok = mgmt.GetString("WebServer", "host")
	conf.Port, ok = mgmt.GetUint32("WebServer", "port")
	conf.VendownName, ok = mgmt.GetString("WebServer", "vendown_name")
	conf.ListenPath, ok = mgmt.GetString("WebServer", "ListenPath")
	conf.SavePath, ok = mgmt.GetString("WebServer", "SavePath")
	conf.Errcbhost, ok = mgmt.GetString("WebServer", "errcbhost")
	conf.Errcbport, ok = mgmt.GetString("WebServer", "errcbport")
	conf.Threadnum, ok = mgmt.GetUint32("WebServer", "Threadnum")

	conf.Mysql.Host, ok = mgmt.GetString("WebServer", "Mysql", "host")
	conf.Mysql.Port, ok = mgmt.GetUint32("WebServer", "Mysql", "port")
	conf.Mysql.User, ok = mgmt.GetString("WebServer", "Mysql", "user")
	conf.Mysql.Password, ok = mgmt.GetString("WebServer", "Mysql", "password")
	conf.Mysql.Name, ok = mgmt.GetString("WebServer", "Mysql", "name")

	fmt.Printf("%+v\n", conf)
	return conf
}
