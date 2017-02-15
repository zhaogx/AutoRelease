package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
	. "utils"
)

type CdncmsWeb struct {
	addr  string         "listen addr"
	s     *WebServer     "web server"
	conf  *WebServerConf "config"
	gconf *GlobalConf    "config"
	db    *sql.DB        "sql connection"
}

func CdncmsWebInit(conf *ServerConfig, ret chan bool) {
	var db *sql.DB
	var err error
	var dburl string
	var web *CdncmsWeb
	var httpSeekHandler RouterHandleFunc

	if conf == nil {
		ret <- false
		return
	}
	web = new(CdncmsWeb)
	web.conf = &conf.WebServer
	web.gconf = &conf.Global

	dburl = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?allowOldPasswords=1",
		web.gconf.LocalSqlServer.User,
		web.gconf.LocalSqlServer.Password,
		web.gconf.LocalSqlServer.Host,
		web.gconf.LocalSqlServer.Port,
		web.gconf.LocalSqlServer.Name)

	db, err = sql.Open("mysql", dburl)
	if err != nil {
		VLOG(VLOG_ERROR, "sql open failed:%s", dburl)
		goto exit
	}
	web.db = db

	web.addr = web.conf.Host + ":" + web.conf.Port
	web.s = HttpInit()

	httpSeekHandler = web.getHandler("/seek")
	web.s.HttpSetRouter("/seek", httpSeekHandler)
	err = web.s.HttpStart(web.addr)
	if err == nil {
		goto exit
	}
	ret <- true
	return
exit:
	if web.db != nil {
		web.db.Close()
	}
	ret <- false
	return
}

func (this *CdncmsWeb) getHandler(pattern string) RouterHandleFunc {

	switch pattern {
	case "/seek":
		return func(w http.ResponseWriter, req *http.Request) {
			var fid string = ""
			var format int = 7

			req.ParseForm()

			for k, v := range req.Form {
				if true == strings.EqualFold(k, "fid") {
					fid = strings.Join(v, "")
				}
			}
			if fid == "" {
				VLOG_LINE(VLOG_ERROR, "Get fid failed.", req.Form)
				return
			}

			MachineDate := time.Now().String()
			tx, _ := this.db.Begin()
			tx.Exec("Insert into content_notify_seek (fid, format, createtime) values (?, ?, ?)",
				strings.ToLower(fid), format, MachineDate)
			tx.Commit()
			VLOG(VLOG_MSG, "Insert db succeed. [%s][%s][%s]", req.URL.RequestURI(), fid, MachineDate)
			return
		}
	default:
		return func(w http.ResponseWriter, req *http.Request) {
			return
		}
	}
}
