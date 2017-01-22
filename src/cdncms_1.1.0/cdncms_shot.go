package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	. "utils"
)

type CdncmsShotMgmt struct {
	gconf    *GlobalConf     "global config"
	conf     *ShotServerConf "config"
	local_db *sql.DB         "sql connection"
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
	mgmt.local_db = db

	return false
}

func (this *CdncmsShotMgmt) init_conf() {
	return
}

func (this *CdncmsShotMgmt) start() int {
	tx, _ := this.local_db.Begin()
	tx.Exec("update FileUploadResult set gl='0' where gl='9'")
	tx.Commit()

	handler := this.walk_func()
	filepath.Walk(this.gconf.ShotServerBaseDir, handler)
	return 0
}

func (this *CdncmsShotMgmt) walk_func() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == true {
			return nil
		}
		//handle file
		if info.Size() <= 0 {
			s := fmt.Sprintf("File size error[%s][%d]", info.Name(), info.Size())
			return errors.New(s)
		}
		name_slice := strings.Split(info.Name(), ".")
		if len(name_slice) <= 1 {
			s := fmt.Sprintf("File name error[%s]", info.Name())
			return errors.New(s)
		}
		if false == strings.EqualFold(name_slice[len(name_slice)-1], "mp4") {
			s := fmt.Sprintf("File name error[%s]", info.Name())
			return errors.New(s)
		}

		mp4_fid := name_slice[0]
		if len(mp4_fid) != 32 || strings.EqualFold(mp4_fid, "d41d8cd98f00b204e9800998ecf8427e") {
			s := fmt.Sprintf("File mp4_fid error[%s]", mp4_fid)
			return errors.New(s)
		}

		e := this.picture_shot(mp4_fid, "-a", path, "160*90", "2", "10")
		if e != nil {
			return e
		}
		//update db
		return nil
	}
}

func (this *CdncmsShotMgmt) picture_shot(fid, style, video_name, resolution, quality, interval string) error {

	var err error
	var i int

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

	//["-a",file,"160*90", "2", "10"]
	o, _ := exec.Command(this.conf.FfmpegPath, "-i", video_name).Output()
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

	i = 0
	succeed := 1

	os.Mkdir("tmp", 0777)

	prefix := "./tmp/" + fid
	files := prefix + "_*"

	for i < sum {
		//./ffmpeg -ss 10 -i 5c3f6fcd49941b6c304ebe6c76effc4a.mp4 -y -f image2 -q:v 2 -y -t 0.01 -s 160*90 ./5c3f6fcd49941b6c304ebe6c76effc4a.jpg
		pic := prefix + "_" + strconv.Itoa(i) + ".jpg"
		cmd := exec.Command(this.conf.FfmpegPath, "-ss", strconv.Itoa(i), "-i", video_name, "-y", "-f", "image2", "-q:v", quality, "-y", "-t", "0.01", "-s", resolution, pic)
		err = cmd.Run()
		if err != nil {
			succeed = 0
			break
		}
		m, _ := strconv.Atoi(interval)
		i += m
	}

	if succeed == 1 {
		tar_name := prefix + ".tar.gz"
		cmd := exec.Command("tar", "zcvf", tar_name, files)
		err = cmd.Run()
		if err == nil {
			new_name := this.conf.ShotPath + "/" + fid + ".mp4.shot"
			os.Rename(tar_name, new_name)
		}
	}
	exec.Command("rm", "-rf", files).Output()
	return err
}
