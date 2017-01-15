package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var DownloadCount int

type DownTask struct {
	url           string
	oldid         string
	filesize      string
	id            string
	fid           string
	md5           string
	errcbhost_3rd string
}

var downloading_num int
var scan_num int
var taskdownChan chan *DownTask

func thread_down(i int) {
	for {
		t := <-taskdownChan
		time.Sleep(time.Second * 1)
		fmt.Println("Start download:", i, t.fid)
		if download(t.url, Cf.Path, t.md5, t.oldid, t.errcbhost_3rd) == true {
			fmt.Println("Insert into resource, md5:", t.md5)
			mysql_sync_local(t.md5)
		}
		downloading_num = downloading_num + 1
		fmt.Println("download end: downloading_num=:", downloading_num)
	}
}
func thrd_content_download() {
	taskdownChan = make(chan *DownTask, Cf.Threadnum)
	for i := 0; i < Cf.Threadnum; i++ {
		fmt.Println("Start download:", i)
		go thread_down(i)
	}
	for {
		fmt.Println("Start download......................")
		stmt, err := mydb.Prepare("select fid, md5, oldid, furl,errcbhost_3rd from content_notify_hg where status = 0 order by level limit 0, ?")
		if err != nil {
			fmt.Println("mysql error:", err)
			time.Sleep(time.Second * 30)
			continue
		}
		res, err := stmt.Query(&Cf.Threadnum)
		if err != nil {
			fmt.Println("mysql error:", err)
			stmt.Close()
			time.Sleep(time.Second * 30)
			continue
		}

		for res.Next() {
			var (
				url           string
				fid           string
				md5           string
				oldid         string
				errcbhost_3rd string
			)
			err := res.Scan(&fid, &md5, &oldid, &url, &errcbhost_3rd)
			if err == nil {
				fmt.Println(fid, md5, url, oldid, errcbhost_3rd)
				t := new(DownTask)
				t.oldid = oldid
				t.fid = fid
				t.md5 = md5
				t.url = url
				t.errcbhost_3rd = errcbhost_3rd
				taskdownChan <- t

				scan_num = scan_num + 1
				fmt.Println("scan_num =:", scan_num)

			} else {
				fmt.Println("Mysql scan err: ", err, url)
			}
		}

		stmt.Close()
		time.Sleep(time.Second * 60)
		for downloading_num != scan_num {
			fmt.Println("downloading_num =:", downloading_num)
			time.Sleep(time.Second * 10)
		}
		//fmt.Println("downloading_num == scan_num")
		downloading_num = 0
		scan_num = 0
	}
}

func download(downurl string, savepath string, md5 string, oldid string, errcbhost_3rd string) bool {
	fmt.Println("Download: ", downurl, savepath, md5, oldid)
	if len(md5) != 32 {
		fmt.Println("#1.Md5 error!")
		return false
	}
	notify_status_third(32, oldid, errcbhost_3rd)

	response, err := http.Get(downurl)
	if err != nil {
		fmt.Println("Get error:", downurl)
		notify_status_third(33, oldid, errcbhost_3rd)
		return false
	}
	defer response.Body.Close()

	u, err := url.Parse(downurl)
	if err != nil {
		fmt.Println("URL error", downurl)
		return false
	}
	fname := path.Join(savepath, path.Base(u.Path))
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Download Filename: ", fname, " create error", err)
		return false
	}

	defer file.Close()
	io.Copy(file, response.Body)

	fid := md5sum(fname)

	if fid != strings.ToLower(md5) {
		os.Remove(fname)
		fmt.Println("File md5 error, remove")
		notify_status_third(34, oldid, errcbhost_3rd)

		stmt, err := mydb.Prepare("select retry from content_notify_hg where status = 0 and oldid =?")
		if err != nil {
			fmt.Println("mysql error:", err)
			return false
		}
		defer stmt.Close()

		res, err := stmt.Query(oldid)
		if err != nil {
			fmt.Println("mysql error:", err)
			stmt.Close()
			return false
		}

		for res.Next() {
			var retry int
			err := res.Scan(&retry)
			if err != nil {
				fmt.Println("Mysql scan err: ", err)
			} else {
				fmt.Println("retry:", retry)

				retry = retry + 1
				fmt.Println("retry:", retry)
				stmt, err := mydb.Prepare("update content_notify_hg set retry=? where oldid=?")
				if err != nil {
					fmt.Println("Update error: ", err, oldid)
					return false
				}
				defer stmt.Close()

				_, err = stmt.Exec(retry, oldid)
				if err != nil {
					fmt.Println("Update error:", err, oldid)
					return false
				}

				if retry == 3 {
					fmt.Println("retry == 3")
					stmt, err = mydb.Prepare("update content_notify_hg set retry=3, status=13 where oldid=?")
					if err != nil {
						fmt.Println("Update error: ", err, oldid)
						return false
					}
					notify_status_third(35, oldid, errcbhost_3rd)
					defer stmt.Close()

					_, err = stmt.Exec(oldid)
					if err != nil {
						fmt.Println("Update error:", err, oldid)
						return false
					}

				}
			}

		}
		return false
	}

	filesize := get_filesize(fname)

	fmt.Println("Download Successful", md5, fname, filesize)
	notify_status_third(31, oldid, errcbhost_3rd)

	stmt, err := mydb.Prepare("update content_notify_hg set filesize=?, status=11 where md5=?")
	//stmt, err := mydb.Prepare("update content_notify_hg set filesize=?, status=11, md5=? where oldid=?")
	if err != nil {
		fmt.Println("Update error: ", err, md5)
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(filesize, md5)
	//_, err = stmt.Exec(filesize, fid, oldid)
	if err != nil {
		fmt.Println("Update error:", err, md5)
		return false
	}
	DownloadCount = DownloadCount + 1
	fmt.Println("OK:", DownloadCount)

	return true
}

func get_contentlength(str_url string) int64 {
	req, _ := http.NewRequest("GET", str_url, nil)
	req.Header.Set("Connection", "close")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	return resp.ContentLength
}
func notify_status_third(statuscode int, oldid string, errcbhost_3rd string) bool {

	errUrl := fmt.Sprintf("%s?oldid=%s&status=%d", errcbhost_3rd, oldid, statuscode)
	fmt.Println("%s", errUrl)
	_, err := http.Get(errUrl)
	if err != nil {
		fmt.Println("notify_status_third Get error:", errUrl)
		return false
	}
	return true

}
