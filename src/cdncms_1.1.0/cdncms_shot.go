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
	"sync"
	"time"
	. "utils"
)

type CdncmsShotMgmt struct {
	gconf     *GlobalConf     "global config"
	conf      *ShotServerConf "config"
	db_local  *sql.DB         "sql connection"
	db_cen    *sql.DB         "sql connection"
	db_online *sql.DB         "sql connection"
	task_ch   chan string     "task chan"

	//增加一个请求列表 避免1个文件被多次遍历到
	mutex    sync.Mutex
	file_map map[string]bool "file list"
}

func CdncmsShotMgmtInit(pconf *ServerConfig, ret chan bool) {
	var err error

	if pconf == nil {
		ret <- false
		return
	}
	mgmt := new(CdncmsShotMgmt)

	mgmt.gconf = &pconf.Global
	mgmt.conf = &pconf.ShotServer

	mgmt.db_local, err = mgmt.initDb(&mgmt.gconf.LocalSqlServer)
	if err != nil {
		ret <- false
		return
	}
	mgmt.db_cen, err = mgmt.initDb(&mgmt.gconf.GlCenSqlServer)
	mgmt.db_online, err = mgmt.initDb(&mgmt.gconf.OnLineSqlServer)

	if mgmt.db_cen == nil && mgmt.db_online == nil {
		mgmt.db_local.Close()
		ret <- false
		return
	}

	mgmt.task_ch = make(chan string)
	mgmt.file_map = make(map[string]bool)

	mgmt.init_conf()
	mgmt.start()
	ret <- true
	return
}

func (this *CdncmsShotMgmt) init_conf() {
	if this.conf.Threadnum <= 0 {
		this.conf.Threadnum = 1
	}
	if this.conf.LoopInterval <= 0 {
		this.conf.LoopInterval = 3
	}
	return
}

func (this *CdncmsShotMgmt) initDb(pconf *SqlServerConf) (*sql.DB, error) {
	if pconf == nil {
		return nil, errors.New("init db. pconf is nil")
	}

	if len(pconf.User) <= 0 || len(pconf.Password) <= 0 || len(pconf.Host) <= 0 || len(pconf.Name) <= 0 {
		return nil, errors.New("db info error")
	}

	dburl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?allowOldPasswords=1",
		pconf.User, pconf.Password, pconf.Host, pconf.Port, pconf.Name)

	db, err := sql.Open("mysql", dburl)
	if err != nil {
		VLOG(VLOG_ERROR, "sql open failed:%s[%s]", dburl, err.Error())
		return nil, err
	}
	VLOG(VLOG_MSG, "sql open succeed [%s]", dburl)
	return db, err
}

func (this *CdncmsShotMgmt) workThread() error {
	mp4_fid := <-this.task_ch
	defer func() {
		this.mutex.Lock()
		delete(this.file_map, mp4_fid)
		this.mutex.Unlock()
	}()

	full_name := this.gconf.ShotServerBaseDir + "/mp4/" + mp4_fid + ".mp4"

	e := this.picture_shot(mp4_fid, "-a", full_name, "160*90", "2", "10")
	if e != nil {
		VLOG(VLOG_ERROR, "%s", e.Error())
		return e
	}

	//update db
	e = this.update_db(mp4_fid)
	if e != nil {
		VLOG(VLOG_ERROR, "update db failed.mp4fid:[%s], error:[%s]", mp4_fid, e.Error())
		return e
	}

	shot_name := this.gconf.ShotServerBaseDir + "/shot/" + mp4_fid + ".mp4.shot"
	shot_fid := md5sum(shot_name)
	dst_name := this.gconf.UploadServerBaseDir + "/waitting/" + shot_fid + ".shot"
	os.Rename(shot_name, dst_name)
	VLOG(VLOG_MSG, "Rename succeed.[%s]->[%s]", shot_name, dst_name)

	return nil
}

func (this *CdncmsShotMgmt) taskDistribute() error {
	var mp4_fid string
	for {
		mp4_fid = ""
		this.mutex.Lock()
		for k, v := range this.file_map {
			if v == false {
				mp4_fid = k
				this.file_map[k] = true
				break
			}
		}
		this.mutex.Unlock()
		if len(mp4_fid) == 32 {
			this.task_ch <- mp4_fid
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (this *CdncmsShotMgmt) start() int {
	var i uint32
	for i = 0; i < this.conf.Threadnum; i++ {
		go this.workThread()
	}
	go this.taskDistribute()

	dir := this.gconf.ShotServerBaseDir + "/mp4"
	for {
		handler := this.walk_handler()
		filepath.Walk(dir, handler)
		time.Sleep(time.Second * time.Duration(this.conf.LoopInterval))
	}
	return 0
}

func (this *CdncmsShotMgmt) walk_handler() filepath.WalkFunc {
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
		//insert map
		this.mutex.Lock()
		if _, ok := this.file_map[mp4_fid]; ok == false {
			this.file_map[mp4_fid] = false
		}
		this.mutex.Unlock()
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

			new_name = this.gconf.ShotServerBaseDir + "/back/" + "/" + fid + ".mp4"
			os.Rename(video_name, new_name)
			VLOG(VLOG_MSG, "Rename succeed.[%s]->[%s]", video_name, new_name)
		}
	}

	err = os.RemoveAll(shot_dir)
	if err != nil {
		VLOG(VLOG_ERROR, "Remove file [%s] failed.[%s]", shot_dir, err.Error())
	} else {
		VLOG(VLOG_ERROR, "Remove file [%s] succeed.", shot_dir)
	}
	return err
}

func (this *CdncmsShotMgmt) select_db(mp4_fid string) (*sql.DB, error) {

	if len(mp4_fid) != 32 {
		return nil, fmt.Errorf("select_db fid len(%d) != 32", len(mp4_fid))
	}

	query := fmt.Sprintf("select fid from resource where fid = '%s'", mp4_fid)

	if this.db_cen != nil {
		rows, err := this.db_cen.Query(query)
		if err != nil {
			VLOG(VLOG_ERROR, "%s[FAILED]", query)
		} else {
			flag := rows.Next()
			rows.Close()
			if flag {
				return this.db_cen, nil
			}
		}
	}
	if this.db_online != nil {
		rows, err := this.db_online.Query(query)
		if err != nil {
			return nil, fmt.Errorf("[%s][FAILED]", query)
		} else {
			flag := rows.Next()
			rows.Close()
			if flag {
				return this.db_online, nil
			}
		}
	}
	return nil, fmt.Errorf("[%s] can not found %s from resource", query)
}

func (this *CdncmsShotMgmt) commit_db(db *sql.DB, query string) {
	if db == nil || len(query) <= 0 {
		VLOG(VLOG_ERROR, "commit_db failed")
		return
	}
	tx, _ := db.Begin()
	_, err := tx.Exec(query)
	tx.Commit()
	if err != nil {
		VLOG(VLOG_ERROR, "[%s][%s]", query, err.Error())
	} else {
		VLOG(VLOG_MSG, "[%s]", query)
	}
}

func (this *CdncmsShotMgmt) update_db(mp4_fid string) error {

	if len(mp4_fid) != 32 {
		return fmt.Errorf("[update_db][mp4_fid len error][%d][%s]", len(mp4_fid), mp4_fid)
	}
	db, err := this.select_db(mp4_fid)
	if err != nil {
		return err
	}

	shot_base_name := mp4_fid + ".mp4.shot"
	shot_full_name := this.gconf.ShotServerBaseDir + "/shot/" + mp4_fid + ".mp4.shot"
	shot_fid := md5sum(shot_full_name)
	if len(shot_fid) <= 0 {
		return fmt.Errorf("Get md5 failed[%s]", shot_full_name)
	}
	filesize := get_filesize(shot_full_name)

	query := fmt.Sprintf("select fid from resource where fid = '%s'", shot_fid)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("%s[FAILED][shot_fullname:%s]", query, shot_full_name)
	} else {
		if rows.Next() {
			query = fmt.Sprintf("update resource set mp4fid = '%s',rawname = '%s' where fid = '%s'",
				mp4_fid, shot_base_name, shot_fid)
		} else {
			query = fmt.Sprintf("insert into resource (fid, rawname, format, inode, filesize, status, mp4fid, ErrorCode) values('%s', '%s', %d, %d, %d, %d, '%s', 10)",
				shot_fid, shot_base_name, 20, 1234567888, filesize, 20, mp4_fid)
		}
		rows.Close()
	}

	this.commit_db(db, query)

	query = fmt.Sprintf("update resource set preview = '%s' where fid = '%s' and format = 7",
		shot_fid, mp4_fid)
	this.commit_db(db, query)

	var VideoRate, MoovOffset, Duration, TotalBitrate, VideoFPS, AudioFPS, AudioRate int32
	var SegIndex, Proprietary, Container, Resolution, AudioCodec, VideoCodec, rawname, iseq, MediaName, CPID string
	query = fmt.Sprintf("select VideoRate, MoovOffset, SegIndex, Proprietary, Duration,"+
		" Container, TotalBitrate, Resolution, AudioCodec, VideoCodec, VideoFPS,"+
		" AudioFPS, AudioRate, rawname, iseq,MediaName,CPID from resource where fid = '%s'",
		mp4_fid)
	rows, err = db.Query(query)
	if err != nil {
		return fmt.Errorf("%s[FAILED][shot_fullname:%s]", query, shot_full_name)
	} else {
		for rows.Next() {
			err = rows.Scan(
				&VideoRate, &MoovOffset, &SegIndex, &Proprietary,
				&Duration, &Container, &TotalBitrate, &Resolution,
				&AudioCodec, &VideoCodec, &VideoFPS, &AudioFPS,
				&AudioRate, &rawname, &iseq, &MediaName, &CPID)
			break
		}
		rows.Close()
	}

	query = fmt.Sprintf("update resource set VideoRate = %d, MoovOffset =%d, SegIndex='%s', Proprietary = '%s', Duration = %d, Container = '%s', TotalBitrate = %d, Resolution = '%s', AudioCodec = '%s', VideoCodec = '%s', VideoFPS = '%d', AudioFPS = '%d', AudioRate = '%d', rawname = '%s.shot', iseq='%s', MediaName='%s',CPID='%s' where fid = '%s' ",
		VideoRate, MoovOffset, SegIndex, Proprietary,
		Duration, Container, TotalBitrate, Resolution,
		AudioCodec, VideoCodec, VideoFPS, AudioFPS,
		AudioRate, rawname, iseq, MediaName, CPID, shot_fid)
	this.commit_db(db, query)
	return nil
}
