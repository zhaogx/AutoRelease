package main

import (
	"database/sql"
	"fmt"
	"time"
	. "utils"
)

type DownloadMgmt struct {
	ch_task chan string     "task channel"
	conf    *DownLoadConfig "config"
	db      *sql.DB         "sql connection"
}

func DownloadMgmtInit() *DownloadMgmt {
	mgmt := new(DownloadMgmt)

	mgmt.conf = GetDownloadConfig(g_conf_mgmt)
	if mgmt.conf == nil {
		return nil
	}
	fmt.Printf("+v\n", mgmt.conf)

	dburl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?allowOldPasswords=1",
		mgmt.conf.Mysql.User, mgmt.conf.Mysql.Password, mgmt.conf.Mysql.Host, mgmt.conf.Mysql.Port, mgmt.conf.Mysql.Name)
	db, err := sql.Open("mysql", dburl)
	if err != nil {
		VLOG(VLOG_ERROR, "sql open failed:%s", dburl)
		return nil
	}
	mgmt.db = db
	mgmt.ch_task = make(chan string, mgmt.conf.Threadnum)

	return mgmt
}

func (this *DownloadMgmt) DownloadStart() int {
	for i := 0; i < int(this.conf.Threadnum); i++ {
		go work(this)
	}
	for {
		tmp := "select fid from content_notify_seek where status = 0 limit 0, ?"
		stmt, err := this.db.Prepare(tmp)
		if err != nil {
			VLOG(VLOG_ERROR, "prepare failed![%s][%s]", tmp, err)
			time.Sleep(time.Second * 10)
			continue
		}
		res, err := stmt.Query(&this.conf.Threadnum)
		if err != nil {
			VLOG(VLOG_ERROR, "query failed![%s][%s]", tmp, err)
			stmt.Close()
			time.Sleep(time.Second * 10)
			continue
		}
		for res.Next() {
			var fid string
			err := res.Scan(&fid)
			if err == nil {
				//update
				stmt2, err := this.db.Prepare("update content_notify_seek set status=1 where fid=?")
				if err == nil {
					_, err = stmt2.Exec(fid)
					if err == nil {
						this.ch_task <- fid
					}
				}
				stmt2.Close()
			} else {
				fmt.Println("Mysql scan err: ", err, fid)
			}
		}
		stmt.Close()
		time.Sleep(time.Second * 10)
	}
}

func work(mgmt *DownloadMgmt) {
	if mgmt == nil {
		return
	}
	for {
		fid := <-mgmt.ch_task
		fmt.Println("add cen down task. fid:", fid)
		if cendownload(mgmt, fid) == true {
			fmt.Println("found resource, fid:", fid)
		} else {
			fmt.Println("found resource failed, fid:", fid)
		}
	}
}

func cendownload(mgmt *DownloadMgmt, fid string) bool {
	//func cendownload(savepath string, fid string) bool {
	if mgmt == nil {
		return false
	}
	savepath := mgmt.conf.SavePath
	listenpath := mgmt.conf.ListenPath

	fmt.Println("Download: ", savepath, fid)
	if len(fid) != 32 {
		fmt.Println("#1.Fid error!")
		return false
	}
	notify_status(mgmt, 42, fid)

	fname := down(mgmt, fid)
	if fname == "" {
		fmt.Println("down() return false.", savepath, fid) //输出执行结果
		notify_status(mgmt, 43, fid)
		return false
	}
	//	md5 := md5sum2(fname)
	md5 := md5sum(fname)
	if md5 != strings.ToLower(fid) {
		fmt.Println("File md5 error, remove.md5:", md5, "fid:", fid, "fname:", fname)
		os.Remove(fname)
		notify_status(mgmt, 44, fid)
		return false
	}

	filesize := get_filesize(fname)

	fmt.Println("======Download Successful. ", md5, fname, filesize)
	notify_status(mgmt, 41, fid)

	stmt, err := mydb.Prepare("update content_notify_seek set filesize=?, status=11,md5=? where fid=?")
	if err != nil {
		fmt.Println("Update error: ", err, md5)
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(filesize, md5, fid)
	if err != nil {
		fmt.Println("Update error:", err, md5)
		return false
	}
	DownloadCount_cen = DownloadCount_cen + 1

	newname := fmt.Sprintf("%s/%s.mp4", listenpath, fid)
	fmt.Println(fname, newname, "-------------------------")
	os.Rename(fname, newname)

	//	cmd := exec.Command("move", fname, newname) //调用Command函数
	//	err := cmd.Run()  //运行指令 ，做判断

	fmt.Println("OK:", DownloadCount_cen)
	return true
}

func down(mgmt *DownloadMgmt, fid string) string {
	//func down(savepath string, fid string) string {
	savepath := mgmt.conf.SavePath

	vendown_name := mgmt.conf.VendownName

	fname := fmt.Sprintf("%s/%s.mp4", savepath, fid)
	fmt.Println(fname)

	cmd := exec.Command(vendown_name, "-d", "-o", fname, "-f", fid) //调用Command函数

	var out bytes.Buffer //缓冲字节
	cmd.Stdout = &out    //标准输出

	err := cmd.Run() //运行指令 ，做判断
	if err != nil {
		fmt.Println(err) //输出执行结果
		return ""
	}
	fmt.Printf("\n%s", out.String()) //输出执行结果
	return fname
}

func notify_status(mgmt *DownloadMgmt, statuscode int, fid string) bool {

	errUrl := fmt.Sprintf("http://%s:%s/freedom-PreViewInfo-errorws.action?mp4fid=%s&status=%d",
		mgmt.conf.Errcbhost, mgmt.conf.Errcbport, fid, statuscode)
	fmt.Println("%s", errUrl)
	_, err := http.Get(errUrl)
	if err != nil {
		fmt.Println("Get error:", errUrl)
		return false
	}
	return true
}
