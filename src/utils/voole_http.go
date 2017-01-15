package utils

import (
	"net/http"
)

type RouterHandleFunc func(http.ResponseWriter, *http.Request)

type WebServer struct {
	router      map[string]RouterHandleFunc "Handler Function"
	listen_addr string                      "listen addr"
}

func HttpInit() *WebServer {
	server := new(WebServer)
	server.router = make(map[string]RouterHandleFunc)
	return server
}

func (this *WebServer) HttpSetRouter(pattern string, f RouterHandleFunc) {
	this.router[pattern] = f
}

func (this *WebServer) HttpStart(addr string) error {
	for k, v := range this.router {
		http.HandleFunc(k, v)
	}
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		VLOG(VLOG_ERROR, "Listen %s failed.", addr)
		return err
	}
	return nil
}
