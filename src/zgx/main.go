package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
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

func main() {
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
