package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	. "utils"
)

type DownloadMgmt struct {
	ch_task chan string    "task channel"
	db      *sql.DB        "sql connection"
	gconf   *GlobalConf    "global config"
	conf    *WebServerConf "config"
}

func DownloadMgmtInit(pconf *ServerConfig) *DownloadMgmt {
	if pconf == nil {
		return nil
	}
	mgmt := new(DownloadMgmt)

	mgmt.gconf = &pconf.Global
	mgmt.conf = &pconf.WebServer
	mgmt.init_conf()

	dburl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?allowOldPasswords=1",
		mgmt.gconf.LocalSqlServer.User,
		mgmt.gconf.LocalSqlServer.Password,
		mgmt.gconf.LocalSqlServer.Host,
		mgmt.gconf.LocalSqlServer.Port,
		mgmt.gconf.LocalSqlServer.Name)

	db, err := sql.Open("mysql", dburl)
	if err != nil {
		VLOG(VLOG_ERROR, "sql open failed:%s", dburl)
		return nil
	}
	mgmt.db = db
	mgmt.ch_task = make(chan string, mgmt.conf.Threadnum)

	mgmt.start()
	return mgmt
}

func (this *DownloadMgmt) init_conf() {
	if this.conf.Threadnum <= 0 {
		this.conf.Threadnum = 1
	}
	if this.conf.CheckInterval <= 0 {
		this.conf.Threadnum = 5
	}
}

func (this *DownloadMgmt) start() int {
	for i := 0; i < int(this.conf.Threadnum); i++ {
		go work(this)
	}
	for {
		query := fmt.Sprintf("select fid from content_notify_seek where status = 0 limit 0,%d", this.conf.Threadnum)
		rows, err := this.db.Query(query)
		if err != nil {
			VLOG(VLOG_ERROR, "%s[FAILED]", query)
			goto next
		}
		VLOG(VLOG_MSG, "%s[SUCCEED]", query)
		for rows.Next() {
			var fid string
			err := rows.Scan(&fid)
			if err != nil {
				VLOG(VLOG_ERROR, "LocalSqlServer scan error [%s] [%s]", fid, err.Error())
				continue
			}
			//update status
			tx, _ := this.db.Begin()
			tx.Exec("update content_notify_seek set status=1 where fid=?", fid)
			tx.Commit()

			this.ch_task <- fid
		}
	next:
		if rows != nil {
			rows.Close()
		}
		time.Sleep(time.Second * time.Duration(this.conf.CheckInterval))
	}
}

func work(mgmt *DownloadMgmt) {
	if mgmt == nil {
		return
	}
	for {
		fid := <-mgmt.ch_task
		VLOG(VLOG_MSG, "Add cen down task. fid:%s", fid)

		cendownload(mgmt, fid)
	}
}

func cendownload(mgmt *DownloadMgmt, fid string) bool {
	if mgmt == nil {
		return false
	}
	savepath := mgmt.gconf.WebServerBaseDir + "/download/"
	listenpath := mgmt.gconf.ShotServerBaseDir + "/mp4/"
	vendown_name := mgmt.conf.VendownName

	VLOG(VLOG_MSG, "Prepare Download [%s][%s]", savepath, fid)
	if len(fid) != 32 {
		VLOG(VLOG_ERROR, "#1.Fid len error[%s]", fid)
		return false
	}
	notify_status(mgmt, 42, fid)

	fname := fmt.Sprintf("%s/%s.mp4", savepath, fid)

	cmd_str := fmt.Sprintf("%s -d -o %s -f %s", vendown_name, fname, fid)
	cmd := exec.Command("/bin/sh", "-c", cmd_str) //调用Command函数
	err := cmd.Run()
	if err != nil {
		VLOG(VLOG_ERROR, "[%s] failed. [%s]", cmd_str, err.Error())
		notify_status(mgmt, 43, fid)
		return false
	}

	md5 := md5sum(fname)
	if false == strings.EqualFold(md5, fid) {
		VLOG(VLOG_ERROR, "File md5 check error.md5:[%s], fid:[%s]", md5, fid)
		os.Remove(fname)
		notify_status(mgmt, 44, fid)
		return false
	}

	notify_status(mgmt, 41, fid)
	filesize := get_filesize(fname)
	VLOG(VLOG_ERROR, "=====Download Successful.[%s][%d]", fname, filesize)

	tx, _ := mgmt.db.Begin()
	tx.Exec("update content_notify_seek set filesize=?, status=?, md5=? where fid=?", filesize, 11, md5, fid)
	tx.Commit()

	//rename
	newname := fmt.Sprintf("%s/%s.mp4", listenpath, fid)
	os.Rename(fname, newname)
	VLOG(VLOG_MSG, "[%s]-->[%s]", fname, newname)
	return true
}

func notify_status(mgmt *DownloadMgmt, statuscode int, fid string) bool {
	if mgmt == nil {
		return false
	}
	errUrl := fmt.Sprintf("http://%s:%s/freedom-PreViewInfo-errorws.action?mp4fid=%s&status=%d",
		mgmt.conf.Errcbhost, mgmt.conf.Errcbport, fid, statuscode)

	VLOG(VLOG_MSG, "Notify status.[%s]", errUrl)

	_, err := http.Get(errUrl)
	if err != nil {
		VLOG(VLOG_MSG, "Get error.[%s]", errUrl)
		return false
	}
	return true
}
