package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	//Read config
	err1 := ReadConfig("server.json")
	if err1 != nil {
		fmt.Println("Read config error: ", err1)
	}

	mydb = mysql_init(Cf.Mysql[0].Host, Cf.Mysql[0].User, Cf.Mysql[0].Password, Cf.Mysql[0].Name)
	mydb1 = mysql_init(Cf.Mysql[1].Host, Cf.Mysql[1].User, Cf.Mysql[1].Password, Cf.Mysql[1].Name)
	mydb2 = mysql_init(Cf.Mysql[2].Host, Cf.Mysql[2].User, Cf.Mysql[2].Password, Cf.Mysql[2].Name)
	mydb3 = mysql_init(Cf.Mysql[3].Host, Cf.Mysql[3].User, Cf.Mysql[3].Password, Cf.Mysql[3].Name)
	fmt.Println("Mysql connected, OK.")

	init_log_and_db()

	//cen content download
	go cen_content_download()
	fmt.Println("Start cen download, OK.")

	Router()
	fmt.Printf("Listening port:%s ...............\n", Cf.Port)

	err := http.ListenAndServe(":"+Cf.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	if mydb != nil {
		mydb.Close()
	}
	os.Exit(0)
}
