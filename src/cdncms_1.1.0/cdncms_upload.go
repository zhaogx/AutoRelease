package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	. "utils"
)

type CdncmsUploadMgmt struct {
	gconf     *GlobalConf       "global config"
	conf      *UploadServerConf "config"
	db_local  *sql.DB           "sql connection"
	db_cen    *sql.DB           "sql connection"
	db_online *sql.DB           "sql connection"

	task_ch chan string "task chan"

	//增加一个请求列表 避免1个文件被多次遍历到
	mutex    sync.Mutex
	file_map map[string]bool "file list"
}

func CdncmsUploadMgmtInit(pconf *ServerConfig, ret chan bool) {
	var err error

	if pconf == nil {
		ret <- false
		return
	}
	mgmt := new(CdncmsUploadMgmt)

	mgmt.gconf = &pconf.Global
	mgmt.conf = &pconf.UploadServer

	mgmt.db_local, err = mgmt.initDb(&mgmt.gconf.LocalSqlServer)
	if err != nil {
		ret <- false
		return
	}
	mgmt.db_cen, err = mgmt.initDb(&mgmt.gconf.GlCenSqlServer)
	if err != nil {
		mgmt.db_local.Close()
		ret <- false
		return
	}
	mgmt.db_online, err = mgmt.initDb(&mgmt.gconf.OnLineSqlServer)
	if err != nil {
		mgmt.db_local.Close()
		mgmt.db_cen.Close()
		ret <- false
		return
	}

	mgmt.init_conf()
	mgmt.start()
	ret <- true
	return
}

func (this *CdncmsUploadMgmt) init_conf() {
	if this.conf.Threadnum <= 0 {
		this.conf.Threadnum = 1
	}
	if this.conf.LoopInterval <= 0 {
		this.conf.LoopInterval = 3
	}
	return
}

func (this *CdncmsUploadMgmt) initDb(pconf *SqlServerConf) (*sql.DB, error) {
	if pconf == nil {
		return nil, errors.New("init db. pconf is nil")
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

func (this *CdncmsUploadMgmt) select_db(mp4_fid string) (*sql.DB, error) {

	if len(mp4_fid) != 32 {
		return nil, fmt.Errorf("select_db fid len(%d) != 32", len(mp4_fid))
	}

	query := fmt.Sprintf("select fid from resource where fid = '%s'", mp4_fid)
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

	rows, err = this.db_online.Query(query)
	if err != nil {
		return nil, fmt.Errorf("[%s][FAILED]", query)
	} else {
		flag := rows.Next()
		rows.Close()
		if flag {
			return this.db_online, nil
		}
	}
	return nil, fmt.Errorf("[%s] can not found %s from resource", query)
}

func (this *CdncmsUploadMgmt) taskDistribute() error {
	var fid string
	for {
		fid = ""
		this.mutex.Lock()
		for k, v := range this.file_map {
			if v == false {
				fid = k
				break
			}
		}
		this.mutex.Unlock()
		if len(fid) == 32 {
			this.task_ch <- fid
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (this *CdncmsUploadMgmt) workThread() error {
	shot_fid := <-this.task_ch
	defer func() {
		this.mutex.Lock()
		delete(this.file_map, shot_fid)
		this.mutex.Unlock()
	}()

	shot_base_name := shot_fid + ".shot"
	shot_full_name := this.gconf.UploadServerBaseDir + "/waitting/" + shot_base_name
	shot_down_name := this.gconf.UploadServerBaseDir + "/down/" + shot_base_name

	count := this.findFileUploadResult(shot_fid)
	if count <= 0 {
		this.insertFileUploadResult(shot_fid, shot_fid)
	}
	gl := this.getFileUploadResult(shot_fid)
	if gl == 0 {
		db, err := this.select_db(shot_fid)
		if err != nil {
			VLOG(VLOG_ERROR, "select db failed! shot_fid:[%s], name:[%s], full_name:[%s]", shot_fid, shot_fid, shot_full_name)
			return err
		}

		this.updateFileUploadResult(shot_fid, 9)

		tool := this.conf.VendownName
		//upload
		cmd_str := fmt.Sprintf("%s -u -i %s -f %s", tool, shot_full_name, shot_fid)
		cmd := exec.Command("/bin/sh", "-c", cmd_str) //调用Command函数
		err = cmd.Run()
		if err != nil {
			VLOG(VLOG_ERROR, "[%s] failed. [%s]", cmd_str, err.Error())
			return err
		}

		//down
		cmd_str = fmt.Sprintf("%s -d -o %s -f %s", tool, shot_down_name, shot_fid)
		cmd = exec.Command("/bin/sh", "-c", cmd_str) //调用Command函数
		err = cmd.Run()
		if err != nil {
			VLOG(VLOG_ERROR, "[%s] failed. [%s]", cmd_str, err.Error())
			return err
		}
		//check md5
		down_md5 := md5sum(shot_down_name)
		cmp := strings.EqualFold(down_md5, shot_fid)
		if cmp == false {
			VLOG(VLOG_ERROR, "md5 check failed. shot_fid:[%s], down_fid:[%s]", shot_fid, down_md5)
			this.updateFileUploadResult(shot_fid, 0)
		} else {
			VLOG(VLOG_MSG, "md5 check succeed. shot_fid:[%s]", shot_fid)
			this.updateFileUploadResult(shot_fid, 1)
			//call back
			this.callback_media(db, shot_fid, 1)
		}
		os.Remove(shot_down_name)
	} else if gl == 1 {
		VLOG(VLOG_WARNNING, "upload has finished. [%s]", shot_base_name)
		os.Remove(shot_base_name)
	} else {
	}
	return nil
}
func (this *CdncmsUploadMgmt) callback_media(db *sql.DB, shot_fid string, status uint32) {
	query := "SET NAMES GBK"
	tx, _ := db.Begin()
	tx.Exec(query)
	tx.Commit()

	var FID, RawName, Resolution, Container, AudioCodec, VideoCodec, Iseq, mp4fid, CPID string
	var Format, published uint8
	var TotalBitrate, Duration uint32
	var FileSize uint64
	var VideoFPS, AudioFPS float64

	query = fmt.Sprintf(
		"SELECT FID,RawName,Format,FileSize,TotalBitrate,Resolution,Container,Duration,AudioCodec,VideoCodec,Iseq,VideoFPS,AudioFPS,mp4fid, CPID, published FROM resource where fid='%s' and status = 20",
		shot_fid)
	rows, err := db.Query(query)
	if err != nil {
		VLOG(VLOG_ERROR, "%s[FAILED]", query)
	} else {
		for rows.Next() {
			rows.Scan(&FID, &RawName, &Format, &FileSize, &TotalBitrate, &Resolution, &Container, &Duration, &AudioCodec, &VideoCodec, &Iseq, &VideoFPS, &AudioFPS, &mp4fid, &CPID, &published)
			break
		}
	}

	var newfile uint8 = 0
	if Format == 20 {
		query = fmt.Sprintf("select fid from content_notify_seek where fid = '%s'", mp4fid)
		rows, err := db.Query(query)
		if err != nil {
			VLOG(VLOG_ERROR, "%s[FAILED]", query)
		} else {
			if rows.Next() == false {
				newfile = 0
			} else {
				newfile = 1
			}
		}
	}
	if newfile == 0 {
		VLOG(VLOG_ERROR, "ERROR:fid or mp4fid :%s not in content_notify_seek", mp4fid)
		return
	}

	var vcodec, fps, resolution, avcconfiglen, avcconfig, acodec, samplerate, language, descr, stream_type, decspecificdescrlen, decspecificdescr, avc_profile, level, nal_unit_size, sps_count, pps_count, cabac, adstatus, fps2, refs, isb2b string

	query = fmt.Sprintf("select VideoCodec,fps,resolution,avcconfiglen,avcconfig,acodec,samplerate,language,descr,stream_type,decspecificdescrlen,decspecificdescr, avc_profile,level,nal_unit_size,sps_count,pps_count,cabac,adstatus,fps2,refs,isb2b,oldid  from resource where fid = '%s'", mp4fid)
	rows, err = db.Query(query)
	if err != nil {
		VLOG(VLOG_ERROR, "%s[FAILED]", query)
	} else {
		if rows.Next() {
			rows.Scan(&vcodec, &fps, &resolution, &avcconfiglen, &avcconfig, &acodec, &samplerate,
				&language, &descr, &stream_type, &decspecificdescrlen, &decspecificdescr, &avc_profile,
				&level, &nal_unit_size, &sps_count, &pps_count, &cabac, &adstatus, &fps2, &refs, &isb2b)
		}
	}

	var url_gz string
	if status == 0 {
		//call back error
		url_gz = fmt.Sprintf(this.gconf.PLATERRORINTERFACE+"mp4fid=%s&status=0", mp4fid)
	} else if status == 1 {
		//call back succeed
		iseq := "00000000000000000000000000000000"
		url_gz = fmt.Sprintf(this.gconf.PLATINTERFACE+"fid=%s&name=%s&format=%d&size=%d&bitrate=%d&resolution=%s"+
			"&ctn=%s&long=%d&audiocodec=%s&videocodec=%s&cdnflag=%s&cpid=%s&dir=%d&iseq=%s&VideoFPS=%.0f&mp4fid=%s"+
			"&vcodec=%s&avcconfiglen=%s&avcconfig=%s&acodec=%s&language=%s&descr=%s&stream_type=%s&decspecificdescrlen=%s"+
			"&decspecificdescr=%s&samplerate=%s&published=%s&avc_profile=%s&level=%s&nal_unit_size=%s&sps_count=%s&pps_count=%s"+
			"&cabac=%s&adstatus=%s&fps2=%s&refs=%s&isb2b=%s",
			FID, RawName, Format, FileSize, TotalBitrate, Resolution, Container,
			Duration, AudioCodec, VideoCodec, "gl", CPID, 0, iseq, VideoFPS, mp4fid,
			vcodec, avcconfiglen, avcconfig, acodec, language, descr, stream_type, decspecificdescrlen, decspecificdescr,
			samplerate, published, avc_profile, level, nal_unit_size, sps_count, pps_count, cabac, adstatus, fps2, refs, isb2b)
	} else {
		return
	}
	resp, err := http.Get(url_gz)
	if err != nil {
		VLOG(VLOG_ERROR, "http Get failed![%s]", url_gz)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		VLOG(VLOG_ERROR, "http Get response failed![%s]", url_gz)
		return
	}
	VLOG(VLOG_MSG, "[CALLBACK SUCCEED][%s][%s]", url_gz, body)
	return
}

func (this *CdncmsUploadMgmt) insertFileUploadResult(md5, name string) {
	query := fmt.Sprintf(
		"insert into fileuploadresult (fid, name, md5name,gl,tj,nj,wh,sx,bj,removed,genm3u8,geninx,genidx,gents,db_sx,db_wh,db_bj,lastblocknum,lastoffset,uploaderr)"+
			" values ('%s','%s','%s','0','0','0','0','0','0','0',0,0,0,0,0,0,0,0,0,0)",
		md5, md5, name)

	tx, _ := this.db_local.Begin()
	tx.Exec(query)
	tx.Commit()
}

func (this *CdncmsUploadMgmt) findFileUploadResult(name string) uint32 {
	var count uint32 = 0
	query := fmt.Sprintf("select count(*) from fileuploadresult where name = '%s'", name)
	rows, err := this.db_local.Query(query)
	if err != nil {
		VLOG(VLOG_ERROR, "%s[FAILED]", query)
	} else {
		rows.Scan(&count)
	}
	return count
}

func (this *CdncmsUploadMgmt) getFileUploadResult(name string) uint32 {
	var gl uint32 = 0
	query := fmt.Sprintf("select gl from fileuploadresult where name = '%s'", name)
	rows, err := this.db_local.Query(query)
	if err != nil {
		VLOG(VLOG_ERROR, "%s[FAILED]", query)
	} else {
		rows.Scan(&gl)
	}
	return gl
}

func (this *CdncmsUploadMgmt) updateFileUploadResult(name string, status uint32) {
	query := fmt.Sprintf("update fileuploadresult set gl=%s where name = '%s'", status, name)
	tx, _ := this.db_local.Begin()
	tx.Exec(query)
	tx.Commit()
}

func (this *CdncmsUploadMgmt) start() int {
	tx, _ := this.db_local.Begin()
	tx.Exec("update FileUploadResult set gl='0' where gl='9'")
	tx.Commit()

	var i uint32
	for i = 0; i < this.conf.Threadnum; i++ {
		go this.workThread()
	}
	go this.taskDistribute()

	dir := this.gconf.UploadServerBaseDir + "/waitting/"
	for {
		handler := this.walk_handler()
		filepath.Walk(dir, handler)
		time.Sleep(time.Second * time.Duration(this.conf.LoopInterval))
	}
	return 0
}

func (this *CdncmsUploadMgmt) walk_handler() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == true {
			return nil
		}
		//handle file
		VLOG(VLOG_MSG, "UploadServer prepare handle [%s]", path)
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

		shot_fid := name_slice[0]
		if len(shot_fid) != 32 || strings.EqualFold(shot_fid, "d41d8cd98f00b204e9800998ecf8427e") {
			s := fmt.Sprintf("File shot_fid error[%s]", shot_fid)
			VLOG(VLOG_ERROR, "%s", s)
			return errors.New(s)
		}
		//insert map
		this.mutex.Lock()
		if _, ok := this.file_map[shot_fid]; ok == false {
			this.file_map[shot_fid] = false
		}
		this.mutex.Unlock()
		return nil
	}
}
