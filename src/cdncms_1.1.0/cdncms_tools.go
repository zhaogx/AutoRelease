package main

import (
	"crypto/md5"
	_ "encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	_ "strings"
	"time"
)

func deamon() bool {
	//fmt.Println(os.Getppid())
	time.Sleep(time.Millisecond * 100)
	if os.Getppid() != 1 {
		//判断当其是否是子进程，当父进程return之后，子进程会被 系统1 号进程接管
		filePath, _ := filepath.Abs(os.Args[0])
		//将命令行参数中执行文件路径转换成可用路径
		cmd := exec.Command(filePath)
		//将其他命令传入生成出的进程
		cmd.Stdin = os.Stdin
		//给新进程设置文件描述符，可以重定向到文件中
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		//开始执行新进程，不等待新进程退出
		cmd.Start()
		return false
	}
	return true
}

func get_filesize(filename string) int64 {
	finfo, err := os.Stat(filename)
	if err == nil {
		return finfo.Size()
	}
	return 0
}

func md5sum(filename string) string {
	var md5new string
	file, inerr := os.Open(filename)
	defer file.Close()

	if inerr == nil {
		md5h := md5.New()
		io.Copy(md5h, file)
		md5new = fmt.Sprintf("%x", md5h.Sum([]byte("")))
	}
	return md5new
}

/*
func md5sum(filename string) string {
	buf := make([]byte, 4096)
	fin, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fin.Close()

	h := md5.New()
	for {
		length, err := fin.Read(buf)
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			if length == 0 {
				break
			}
		}
		if length > 0 {
			h.Write(buf[:length])
		}
	}
	str := h.Sum(nil)
	return strings.ToLower(hex.EncodeToString(str[:]))
}
*/
