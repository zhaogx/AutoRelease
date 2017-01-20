package main

type WebServerConf struct {
	Host          string
	Port          string
	VendownName   string
	ListenPath    string
	SavePath      string
	Errcbhost     string
	Errcbport     string
	Threadnum     uint32
	CheckInterval uint32
	Mysql         struct {
		Host     string
		Port     uint16
		User     string
		Password string
		Name     string
	}
}

type ServerConfig struct {
	WebServer WebServerConf
}
