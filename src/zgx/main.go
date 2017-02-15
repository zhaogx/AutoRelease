package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	_ "strconv"
	"strings"
	"syscall"
	"time"
	. "utils"
)

func event_cb(event int, fname string) int {
	fmt.Println(event, fname)
	return 0
}

func tcp_server() {
	addr, _ := net.ResolveTCPAddr("tcp", ":10000")
	listener, _ := net.ListenTCP("tcp", addr)

	r := make([]byte, 1024)

	for {
		conn, _ := listener.Accept()
		conn.Read(r)
		fmt.Printf("[recv][%s][%s]%s\n", conn.LocalAddr(), conn.RemoteAddr(), string(r))
		conn.Close()
	}
	return
}

func tcp_client() {
	remote, _ := net.ResolveTCPAddr("tcp", "10.5.6.155:10000")
	conn, _ := net.DialTCP("tcp", nil, remote)
	defer conn.Close()

	str := "hello, server!"
	w := []byte(str)

	var count int
	var ok error
	for {
		count, ok = conn.Write(w)
		if ok != nil {
			fmt.Printf("[send][%s][%s]failed\n", conn.LocalAddr(), conn.RemoteAddr())
			return
		} else {
			fmt.Printf("[send][%s][%s]succeed, count:%d\n", conn.LocalAddr(), conn.RemoteAddr(), count)
			time.Sleep(time.Second)
		}
	}
	return
}

type web_server_t struct {
	/*
		webserver struct {
			host  string
			port  string
			bool  bool
			mysql struct {
				host     string
				port     uint16
				user     string
				password string
				name     string
			}
			errcbhost_3rd string
			threadnum     int
		}
	*/
	scanserver struct {
		host string
		port uint16
		path string
	}
}

func WalkFunc(path string, info os.FileInfo, err error) error {
	fmt.Println(path, info.Name(), info.IsDir())
	return err
}

func error_test_func() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recover[%s]\n", r)
		}
	}()
	panic(errors.New("panic test"))
	return fmt.Errorf("%s[%d]", "fmt test error", 10)
	return errors.New("test error, hello world")
}

func go_func(c chan bool) {
	time.Sleep(time.Second * 1)
	fmt.Println("before in...")
	c <- true
	fmt.Println("finish in...")
}

func main() {

	Vlog_init("Log.json")
	{
		c := make(chan bool)
		go go_func(c)
		time.Sleep(time.Second * 3)
		fmt.Println("before out...")
		<-c
		fmt.Println("finish out...")
		time.Sleep(time.Second * 3)
		return
	}
	{
		dburl := "root:mysql*()@tcp(10.5.6.155:3306)/voole?allowOldPasswords=1"
		db, err := sql.Open("mysql", dburl)
		if err != nil {
			VLOG(VLOG_ERROR, "sql open failed:%s[%s]", dburl, err.Error())
			return
		}
		VLOG(VLOG_MSG, "sql open succeed [%s]", dburl)

		query := "select name from test_table where id = 10"
		rows, err := db.Query(query)
		if err != nil {
			VLOG(VLOG_ERROR, "[%s][failed][%s]", query, err.Error())
			return
		}
		if false == rows.Next() {
			VLOG(VLOG_ERROR, "[%s][failed]", query)
		}
		for rows.Next() {
			var id uint32
			var name string
			rows.Scan(&id, &name)
			VLOG(VLOG_MSG, "result:[%d:%s]", id, name)
		}
		VLOG(VLOG_MSG, "finished.......")

		query = "select count(*) from test_table"
		rows, err = db.Query(query)
		if err != nil {
			VLOG(VLOG_ERROR, "[%s][failed][%s]", query, err.Error())
			return
		}
		for rows.Next() {
			var count uint32
			rows.Scan(&count)
			VLOG(VLOG_MSG, "result:[%d]", count)
		}

		query = "SET NAMES latin1"
		tx, _ := db.Begin()
		_, err = tx.Exec(query)
		if err != nil {
			VLOG(VLOG_ERROR, "[%s] failed", query)
		} else {
			VLOG(VLOG_MSG, "[%s] succeed", query)
		}
		tx.Commit()

		query = "select * from test_table"
		rows, err = db.Query(query)
		if err != nil {
			VLOG(VLOG_ERROR, "[%s][failed][%s]", query, err.Error())
			return
		}
		for rows.Next() {
			var id string
			var name string
			rows.Scan(&id, &name)
			VLOG(VLOG_MSG, "result:[%s:%s]", id, name)
		}
		VLOG(VLOG_MSG, "finished.......")

		return
	}
	{
		e := error_test_func()
		fmt.Printf("%s\n", e)
		return
	}
	filepath.Walk("/tmp", WalkFunc)
	return
	var str_slice = []string{"hello", "test", "world"}
	fmt.Println(str_slice)

	tmp_str := strings.Join(str_slice, "")
	fmt.Println(tmp_str)
	return

	var web_server web_server_t
	//var web_server map[string]interface{}
	ReadConfig("server2.json", &web_server)
	fmt.Println(web_server)
	st := fmt.Sprintln(web_server)
	fmt.Println(st)
	return

	cname := "server2.json"
	conf_mgmt := VooleConfigInit(cname)
	if conf_mgmt == nil {
		fmt.Println("config init failed")
	}
	fmt.Println("====================================================================")
	var result string
	var f bool
	var val_u32 uint32
	var val_f32 float32
	var val_f64 float64

	result, f = conf_mgmt.GetString("WebServer", "host")
	fmt.Println(result, f)
	result, f = conf_mgmt.GetString("WebServer", "Mysql", "host")
	fmt.Println(result, f)
	val_u32, f = conf_mgmt.GetUint32("WebServer", "Threadnum")
	fmt.Println(val_u32, f)

	val_f32, f = conf_mgmt.GetFloat32("WebServer", "thread_id")
	fmt.Println(val_f32, f)
	val_f64, f = conf_mgmt.GetFloat64("WebServer", "thread_id")
	fmt.Println(val_f64, f)

	return
	//socket test
	lock := make(chan int)
	go tcp_server()
	time.Sleep(time.Second)
	go tcp_client()
	<-lock
	return
	//channel range
	ch := make(chan int)
	go func() {
		for i := 0; i < 5; i++ {
			ch <- i
		}
		time.Sleep(time.Second * 5)
		close(ch)
	}()

	for val := range ch {
		fmt.Println(val)
	}
	fmt.Println("ending....")
	return

	//fsnotify
	done := make(chan bool)
	mgmt, _ := VooleNewWatcher("/tmp/foo", event_cb)
	mgmt2, _ := VooleNewWatcher("/tmp/foo2", event_cb)

	time.Sleep(time.Second * 500)
	VooleCloseWatcher(mgmt)
	VooleCloseWatcher(mgmt2)
	<-done
	return

	//notify signal
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGUSR1)
	fmt.Println("waiting signal......")
	s := <-c
	fmt.Println("get signal:", s)
	return

	//cmd args -- os
	fmt.Println(os.Args)
	//cmd args -- flag

	ok := flag.Bool("ok", false, "is ok")
	name := flag.String("name", "", "string name")
	//var str string
	//flag.StringVar(&str, "name", "", "string name")
	flag.Parse()

	fmt.Println("ok:", *ok)
	fmt.Println("name:", *name)

	VLOG(VLOG_ERROR, "error")
	VLOG(VLOG_WARNNING, "warnning")
	VLOG(VLOG_MSG, "msg")
	VLOG(VLOG_DEBUG, "debug")
	Vlog_destory()
}
