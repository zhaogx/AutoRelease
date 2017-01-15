package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	VLOG_CLOSE = iota
	VLOG_ERROR
	VLOG_WARNNING
	VLOG_MSG
	VLOG_DEBUG
)

var log_level_str = []string{
	"",
	"ERROR",
	"WARNNING",
	"MSG",
	"DEBUG"}

type config struct {
	Log_level int
	Log_path  string
}

var cf config
var log_init_succeed uint8 = 0
var g_log_file *os.File = nil

func Vlog_init(conf_path string) int {
	var err error
	err = ReadConfig(conf_path, &cf)
	if err != nil {
		fmt.Println("[Vlog_init]read config file failed")
		goto exit
	}
	_, err = os.Stat(cf.Log_path)
	if err != nil {
		err = os.MkdirAll(cf.Log_path, 0777)
		if err != nil {
			fmt.Println("[Vlog_init]make dir failed.", cf.Log_path)
			goto exit
		}
	}
	log_init_succeed = 1
	return 0
exit:
	log_init_succeed = 0
	return -1
}

func Vlog_set_level(level int) int {
	if level < 0 {
		level = 0
	}
	if level > VLOG_DEBUG {
		level = VLOG_DEBUG
	}
	cf.Log_level = level
	return 0
}

func Vlog_destory() {
	if g_log_file != nil {
		g_log_file.Close()
		g_log_file = nil
	}
	return
}

func VLOG(level int, format string, a ...interface{}) {
	var err error
	var full_name string
	if cf.Log_level <= 0 || cf.Log_level < level || log_init_succeed == 0 {
		return
	}
	t := time.Now()

	fname := fmt.Sprintf("%04d_%02d_%02d.log", t.Year(), t.Month(), t.Day())
	full_name = cf.Log_path + "/" + fname

	_, err = os.Stat(full_name)
	if err != nil {
		if false == os.IsNotExist(err) {
			return
		}
		if g_log_file != nil {
			g_log_file.Close()
			g_log_file = nil
		}
	}
	if g_log_file == nil {
		g_log_file, err = os.OpenFile(full_name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("open file", g_log_file, "failed")
			return
		}
	}

	_, file, line, _ := runtime.Caller(1)
	array := strings.Split(file, "/")
	file = array[len(array)-1]

	tmp_format := fmt.Sprintf("[%s:%d][%s] ", file, line, log_level_str[level]) + format

	pLog := log.New(g_log_file, "", log.Lmicroseconds)
	pLog.Printf(tmp_format, a...)
	return
}
