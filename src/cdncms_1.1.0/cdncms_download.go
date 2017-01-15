package main

import (
	. "utils"
)

type DownloadMgmt struct {
	thread_num uint32      "download thread count"
	ch_task    chan string "task channel"
}

func DownloadMgmtInit(th_count uint32) *DownloadMgmt {
	mgmt := new(DownloadMgmt)
	if th_count <= 0 {
		th_count = 1
	}
	mgmt.thread_num = th_count
	mgmt.ch_task = make(chan string)
	return mgmt
}

func (this *DownloadMgmt) DownloadStart() {
}

var taskdown_queue chan string

func cen_content_download() {
	taskdown_queue = make(chan string, Cf.Threadnum)
	for i := 0; i < Cf.Threadnum; i++ {
		go work(i)
	}
	for {
		fmt.Println("Start download......................")
		stmt, err := mydb.Prepare("select fid from content_notify_seek where status = 0 limit 0, ?")
		if err != nil {
			fmt.Println("mysql error:", err)
			time.Sleep(time.Second * 10)
			continue
		}

		res, err := stmt.Query(&Cf.Threadnum)
		if err != nil {
			fmt.Println("mysql error:", err)
			stmt.Close()
			time.Sleep(time.Second * 10)
			continue
		}

		for res.Next() {
			var (
				fid string
			)
			err := res.Scan(&fid)
			if err == nil {
				//update
				stmt2, err := mydb.Prepare("update content_notify_seek set status=1 where fid=?")
				if err == nil {
					_, err = stmt2.Exec(fid)
					if err == nil {
						taskdown_queue <- fid
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

func work(i int) {
	for {
		fid := <-taskdown_queue
		fmt.Println("add cen down task. fid:", fid)
		if cendownload(Cf.Path, fid) == true {
			fmt.Println("found resource, fid:", fid)
		} else {
			fmt.Println("found resource failed, fid:", fid)
		}
	}
}

func cendownload(savepath string, fid string) bool {
	fmt.Println("Download: ", savepath, fid)
	if len(fid) != 32 {
		fmt.Println("#1.Fid error!")
		return false
	}
	notify_status(42, fid)

	fname := down(savepath, fid)
	if fname == "" {
		fmt.Println("down() return false.", savepath, fid) //输出执行结果
		notify_status(43, fid)
		return false
	}
	//	md5 := md5sum2(fname)
	md5 := md5sum(fname)
	if md5 != strings.ToLower(fid) {
		fmt.Println("File md5 error, remove.md5:", md5, "fid:", fid, "fname:", fname)
		os.Remove(fname)
		notify_status(44, fid)
		return false
	}

	filesize := get_filesize(fname)

	fmt.Println("======Download Successful. ", md5, fname, filesize)
	notify_status(41, fid)

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

	newname := fmt.Sprintf("%s/%s.mp4", Cf.Listen_path, fid)
	fmt.Println(fname, newname, "-------------------------")
	os.Rename(fname, newname)

	//	cmd := exec.Command("move", fname, newname) //调用Command函数
	//	err := cmd.Run()  //运行指令 ，做判断

	fmt.Println("OK:", DownloadCount_cen)
	return true
}

func down(savepath string, fid string) string {
	fname := fmt.Sprintf("%s/%s.mp4", savepath, fid)
	fmt.Println(fname)

	cmd := exec.Command(Cf.Vendown_name, "-d", "-o", fname, "-f", fid) //调用Command函数

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
