package main

import (
	"fmt"
	"log"
	"os"
)

var logger *log.Logger

func init_log_and_db() {
	//init log
	fout, err := os.OpenFile("run.log", os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Open Logfile Error:", "run.log", err)
		os.Exit(-1)
	}

	defer fout.Close()
	logger = log.New(fout, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	logger.Println("Start")
	//init mysql
}
