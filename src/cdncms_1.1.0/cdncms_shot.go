package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	. "utils"
)

type CdncmsShotMgmt struct {
	gconf    *GlobalConf     "global config"
	conf     *ShotServerConf "config"
	local_db *sql.DB         "sql connection"
	task_ch  chan string     "task chan"
	//增加一个请求列表 避免1个文件被多次遍历到
}

func CdncmsShotMgmtInit(pconf *ServerConfig) bool {
	if pconf == nil {
		return false
	}
	mgmt := new(CdncmsShotMgmt)

	mgmt.gconf = &pconf.Global
	mgmt.conf = &pconf.ShotServer
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
		return false
	}
	VLOG(VLOG_MSG, "sql open succeed [%s]", dburl)
	mgmt.local_db = db
	mgmt.task_ch = make(chan string)

	mgmt.init_conf()
	mgmt.start()
	return false
}

func cdncmsShotWorkThread(mgmt *CdncmsShotMgmt) error {
	if mgmt == nil {
		return nil
	}
	fid := <-mgmt.task_ch
	full_name := mgmt.gconf.ShotServerBaseDir + "/mp4/" + fid + ".mp4"

	e := mgmt.picture_shot(fid, "-a", full_name, "160*90", "2", "10")
	if e != nil {
		VLOG(VLOG_ERROR, "%s", e.Error())
		return e
	}
	//update db
	return nil
}

func (this *CdncmsShotMgmt) init_conf() {
	return
}

func (this *CdncmsShotMgmt) start() int {
	tx, _ := this.local_db.Begin()
	tx.Exec("update FileUploadResult set gl='0' where gl='9'")
	tx.Commit()

	var i uint32
	for i = 0; i < this.conf.Threadnum; i++ {
		go cdncmsShotWorkThread(this)
	}

	dir := this.gconf.ShotServerBaseDir + "/mp4"
	for {
		handler := this.walk_func()
		filepath.Walk(dir, handler)

		info, err := ioutil.ReadDir(dir)
		if len(info) == 0 || err != nil {
			time.Sleep(time.Second * 5)
		}
	}
	return 0
}

func (this *CdncmsShotMgmt) walk_func() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == true {
			return nil
		}
		//handle file
		VLOG(VLOG_MSG, "ShotServer prepare handle [%s]", path)
		if info.Size() <= 0 {
			s := fmt.Sprintf("File size error[%s][%d]", info.Name(), info.Size())
			VLOG(VLOG_ERROR, "%s", s)
			return errors.New(s)
		}
		name_slice := strings.Split(info.Name(), ".")
		if len(name_slice) <= 1 {
			s := fmt.Sprintf("File name error[%s]", info.Name())
			VLOG(VLOG_ERROR, "%s", s)
			return errors.New(s)
		}
		if false == strings.EqualFold(name_slice[len(name_slice)-1], "mp4") {
			s := fmt.Sprintf("File name error[%s]", info.Name())
			VLOG(VLOG_ERROR, "%s", s)
			return errors.New(s)
		}

		mp4_fid := name_slice[0]
		if len(mp4_fid) != 32 || strings.EqualFold(mp4_fid, "d41d8cd98f00b204e9800998ecf8427e") {
			s := fmt.Sprintf("File mp4_fid error[%s]", mp4_fid)
			VLOG(VLOG_ERROR, "%s", s)
			return errors.New(s)
		}

		this.task_ch <- mp4_fid
		return nil
	}
}

func (this *CdncmsShotMgmt) picture_shot(fid, style, video_name, resolution, quality, interval string) error {

	var err error
	var i int
	var cmd_str string

	//param check
	var e error
	if style != "-a" {
		s := fmt.Sprintf("[picture_shot style error][%s][%s]", style, video_name)
		return errors.New(s)
	}
	split := strings.Split(resolution, "*")
	if len(split) < 2 {
		s := fmt.Sprintf("[picture_shot resolution error][%s][%s]", resolution, video_name)
		return errors.New(s)
	}
	_, e = strconv.ParseInt(split[0], 10, 0)
	if e != nil {
		return e
	}
	_, e = strconv.ParseInt(split[1], 10, 0)
	if e != nil {
		return e
	}
	_, e = strconv.ParseInt(interval, 10, 0)
	if e != nil {
		return e
	}
	i64, e := strconv.ParseInt(quality, 10, 0)
	if e != nil {
		return e
	}
	if i64 > 60 {
		s := fmt.Sprintf("[picture_shot quality error][%s][%s]", quality, video_name)
		return errors.New(s)
	}

	cmd_str = fmt.Sprintf("%s -i %s", this.conf.FfmpegPath, video_name)
	o, _ := exec.Command("/bin/sh", "-c", cmd_str).CombinedOutput()
	out := string(o)

	dur_str := "Duration: "
	i = strings.Index(out, dur_str)
	if i < 0 {
		s := fmt.Sprintf("[picture_shot get duration failed!!!][%s][%s]", video_name, out)
		return errors.New(s)
	}
	t := out[i+len(dur_str) : i+len(dur_str)+8]
	slice := strings.Split(t, ":")
	hour, _ := strconv.Atoi(slice[0])
	min, _ := strconv.Atoi(slice[1])
	sec, _ := strconv.Atoi(slice[2])
	sum := hour*3600 + min*60 + sec

	VLOG(VLOG_MSG, "duration:[%s], [%d:%d:%d] [sum:%d]", t, hour, min, sec, sum)

	i = 0
	succeed := 1

	shot_dir := fmt.Sprintf("./%s", fid)
	os.Mkdir(shot_dir, 0777)
	prefix := shot_dir + "/" + fid

	for i < sum {
		pic := prefix + "_" + strconv.Itoa(i) + ".jpg"
		cmd_str = fmt.Sprintf("%s -ss %d -i %s -y -f image2 -q:v %s -y -t 0.01 -s %s %s", this.conf.FfmpegPath, i, video_name, quality, resolution, pic)
		cmd := exec.Command("/bin/sh", "-c", cmd_str)
		err = cmd.Run()
		if err != nil {
			succeed = 0
			break
		}
		m, _ := strconv.Atoi(interval)
		i += m
	}

	if succeed == 1 {
		source_name := prefix + ".tar.gz"
		cmd_str = fmt.Sprintf("tar zcvf %s %s*", source_name, prefix)
		cmd := exec.Command("/bin/sh", "-c", cmd_str)
		err = cmd.Run()
		if err == nil {
			VLOG(VLOG_MSG, "[%s] succeed.", cmd_str)
			new_name := this.gconf.ShotServerBaseDir + "/shot/" + "/" + fid + ".mp4.shot"
			os.Rename(source_name, new_name)
			VLOG(VLOG_MSG, "Rename succeed.[%s]->[%s]", source_name, new_name)

			source_name = video_name
			new_name = this.gconf.ShotServerBaseDir + "/back/" + "/" + fid + ".mp4"
			os.Rename(source_name, new_name)
			VLOG(VLOG_MSG, "Rename succeed.[%s]->[%s]", source_name, new_name)
		}
	}
	cmd_str = fmt.Sprintf("rm -rf %s", shot_dir)
	o, err = exec.Command("/bin/sh", "-c", cmd_str).CombinedOutput()
	if err != nil {
		VLOG(VLOG_ERROR, "[%s] failed.[%s][%s]", cmd_str, string(o), err.Error())
	}
	VLOG(VLOG_MSG, "[%s] succeed.", cmd_str)
	return err
}
