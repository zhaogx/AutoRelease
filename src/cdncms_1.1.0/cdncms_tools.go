package main

import (
	"crypto/md5"
	_ "encoding/hex"
	"fmt"
	"io"
	"os"
	_ "strings"
)

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
