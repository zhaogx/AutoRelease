package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Shen Zhen Sheng Ying ts upload notify
type Notify struct {
	fid               string
	furl              string
	md5               string
	cpid              string
	src               string
	name              string
	ftype             string
	oldid             string
	format            int
	level             string
	errcbhost_3rd     string
	platinterface_3rd string
}

func seekhandler(w http.ResponseWriter, req *http.Request) {

	notify := new(Notify)
	req.ParseForm()
	for k, v := range req.Form {
		if k == "fid" {
			notify.fid = strings.Join(v, "")
		}
	}
	/*Insert */
	if notify.fid == "" {
		w.Write([]byte("{\"status\": 0, \"msg\": \"Need fid\"}"))
		return
	}

	MachineDate := time.Now().String()
	u, err := url.Parse(notify.furl)
	if err == nil {
		notify.name = path.Base(u.Path)
		if notify.name == "." {
			notify.name = ""
		}

		fext := filepath.Ext(notify.name)

		if fext == ".zip" || fext == ".apk" {
			notify.format = 0
		}

		if fext == ".ts" {
			notify.format = 6
		}

		if fext == ".mp4" {
			notify.format = 7
		}

		if fext == ".m3u8" {
			notify.format = 5
		}
	}

	stmt, err := mydb.Prepare("Insert into content_notify_seek (fid, format, createtime) values (?, ?, ?)")
	if err != nil {
		fmt.Println("Insert error:", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(strings.ToLower(notify.fid), notify.format, MachineDate)
	if err != nil {
		fmt.Println("Insert error:", err, res)
		w.Write([]byte("{\"status\": 0, \"msg\": \"Insert into Database Faild\"}"))
		return
	}
	log.Println(req.URL.RequestURI())
	w.Write([]byte("{\"status\":1, \"msg\":\"Save Successful\"}"))
}

func notifyhandler(w http.ResponseWriter, req *http.Request) {
	//http://localhost:8001?cpid=10008001&fid=1072000001&md5=0a30e8937013d83b298dec490e5e6d51&url=ftp://hometv:HoMeTv@112.124.118.174/h264_640/1072000001.mkv&oldid=123
	//notify := Notify{}
	notify := new(Notify)
	req.ParseForm()
	for k, v := range req.Form {
		if k == "fid" {
			notify.fid = strings.Join(v, "")
		}
		if k == "url" {
			notify.furl = strings.Join(v, "")
		}
		if k == "md5" {
			notify.md5 = strings.Join(v, "")
		}
		if k == "cpid" {
			notify.cpid = strings.Join(v, "")
		}
		if k == "ftype" {
			notify.ftype = strings.Join(v, "")
		}
		if k == "oldid" {
			notify.oldid = strings.Join(v, "")
		}
		if k == "level" {
			notify.level = strings.Join(v, "")
		}
		if k == "errcbhost_3rd" {
			notify.errcbhost_3rd = strings.Join(v, "")
		}
		if k == "platinterface_3rd" {
			notify.platinterface_3rd = strings.Join(v, "")
		}
	}

	/*Insert */
	if notify.oldid == "" || notify.furl == "" || notify.cpid == "" || notify.md5 == "" || notify.errcbhost_3rd == "" || notify.platinterface_3rd == "" {
		w.Write([]byte("{\"status\": 0, \"msg\": \"Need furl, cpid, md5, oldid,errcbhost_3rd,platinterface_3rd\"}"))
		return
	}

	if notify.level == "" {
		notify.level = "100"
	}
	MachineDate := time.Now().String()
	u, err := url.Parse(notify.furl)
	if err == nil {
		notify.name = path.Base(u.Path)
		if notify.name == "." {
			notify.name = ""
		}

		fext := filepath.Ext(notify.name)

		if fext == ".zip" || fext == ".apk" {
			notify.format = 0
		}

		if fext == ".ts" {
			notify.format = 6
		}

		if fext == ".mp4" {
			notify.format = 7
		}

		if fext == ".m3u8" {
			notify.format = 5
		}
	}

	stmt, err := mydb.Prepare("Insert into content_notify_hg (md5, furl, name, cpid, src, format, oldid,level, createtime,errcbhost_3rd,platinterface_3rd) values (?, ?, ?, ?, ?, ?, ?,?, ?,?,?)")
	if err != nil {
		fmt.Println("Insert content_notify_hg error:", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(strings.ToLower(notify.md5), notify.furl, notify.name, notify.cpid, req.URL.RawQuery, notify.format, notify.oldid, notify.level, MachineDate, notify.errcbhost_3rd, notify.platinterface_3rd)
	if err != nil {
		fmt.Println("Insert content_notify_hg error:", err, res)
		w.Write([]byte("{\"status\": 0, \"msg\": \"Insert into Database content_notify_hg Faild\"}"))

		stmt, err := mydb.Prepare("update content_notify_hg set status=0,retry=0,md5=?, furl=?, name=?, cpid=?, src=?, format=?, oldid=?,level=?, createtime=? errcbhost_3rd=?,platinterface_3rd=? where md5 = ?")
		if err != nil {
			fmt.Println("Update content_notify_hg error:", err)
		}
		defer stmt.Close()

		res, err := stmt.Exec(strings.ToLower(notify.md5), notify.furl, notify.name, notify.cpid, req.URL.RawQuery, notify.format, notify.oldid, notify.level, MachineDate, notify.errcbhost_3rd, notify.platinterface_3rd, strings.ToLower(notify.md5))
		if err != nil {
			fmt.Println("Update content_notify_hg error:", err, res)
			w.Write([]byte("{\"status\": 0, \"msg\": \"Update Database content_notify_hg Faild\"}"))
			return
		}
		//return
	}
	log.Println(req.URL.RequestURI())
	w.Write([]byte("{\"status\":1, \"msg\":\"Save Successful\"}"))
}

type ScanMsg struct {
	fid      string
	cpid     string
	filesize string
	duration string
	filename string
	ftype    string
	format   int
}

//B2B voole content Scan and insert into mysql , need callback palatform
func ScanMsgHandler(w http.ResponseWriter, req *http.Request) {
	msg := new(ScanMsg)
	req.ParseForm()

	if req.Method != "GET" {
		return
	}
	fmt.Println(req.Form)
	for k, v := range req.Form {
		if k == "fid" {
			msg.fid = strings.Join(v, "")
		}
		if k == "cpid" {
			msg.cpid = strings.Join(v, "")
		}
		if k == "size" {
			msg.filesize = strings.Join(v, "")
		}
		if k == "duration" {
			msg.duration = strings.Join(v, "")
		}
		if k == "name" {
			msg.filename = strings.Join(v, "")
		}
		if k == "ftype" {
			msg.ftype = strings.Join(v, "")
		}
	}
	fmt.Println(msg)

	MachineDate := time.Now().String()
	fmt.Println("fid:", msg.fid, "size:", msg.filesize, "name:", msg.filename, "ftype:", msg.ftype, "cpid:", msg.cpid, "time:", MachineDate)
	/*Insert */
	if msg.fid == "" || msg.filesize == "" || msg.filename == "" || msg.ftype == "" {
		w.Write([]byte("{\"status\": 0, \"msg\": \"Faild: Need fid, size, cpid, name, ftype\"}"))
		return
	}
	if len(msg.cpid) == 0 {
		msg.cpid = "100010"
	}

	stmt, err := mydb.Prepare("Insert into scan_info (fid, filesize, filename, duration, cpid, ftype, createtime) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		w.Write([]byte("{\"status\": 0, \"msg\": \"Init  SQL Faild\"}"))
		return
	}
	defer stmt.Close()

	fmt.Println(msg.filename, msg.filesize, msg.duration)
	res, err := stmt.Exec(msg.fid, msg.filesize, msg.filename, msg.duration, msg.cpid, msg.ftype, MachineDate)
	if err != nil {
		fmt.Println("Insert error:", err, res)
		w.Write([]byte("{\"status\": 0, \"msg\": \"Insert Into Database Faild\"}"))
		return
	}

	log.Println(req.URL.RequestURI())
	w.Write([]byte("{\"status\":1, \"msg\":\"Save Successful\"}"))
}
