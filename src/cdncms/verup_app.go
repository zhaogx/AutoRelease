package main

import (
	"fmt"
	"time"
)

func verup_app_download() {
	for {
		stmt, err := mydb.Prepare("select fid, md5, furl from content_notify where status=0 and format=0 and cpid=100010 limit 0, 10")
		if err != nil {
			fmt.Println("mysql error:", err)
			time.Sleep(time.Second * 30)
			continue
		}
                
		res, err := stmt.Query()
		if err != nil {
			fmt.Println("mysql error:", err)
                        stmt.Close()
			time.Sleep(time.Second * 30)
			continue
		}

		//fmt.Println("new version file:", res, err)

		for res.Next() {
			var (
				url string
				fid string
				md5 string
			)
			if err := res.Scan(&fid, &md5, &url); err == nil {
				fmt.Println(fid, md5, url)
				download(url, "/opt/ftp/upload_ftp/", md5, "","")
			}
		}

                stmt.Close()
		time.Sleep(time.Second * 100)
	}
}
