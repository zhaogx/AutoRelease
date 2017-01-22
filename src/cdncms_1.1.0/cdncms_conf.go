package main

type SqlServerConf struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Name     string
}

type GlobalConf struct {
	WebServerBaseDir  string
	ShotServerBaseDir string
	LocalSqlServer    SqlServerConf
}

type WebServerConf struct {
	Host          string
	Port          string
	VendownName   string
	Errcbhost     string
	Errcbport     string
	Threadnum     uint32
	CheckInterval uint32
}

type ShotServerConf struct {
	ShotPath   string
	FfmpegPath string
}

type ServerConfig struct {
	Global     GlobalConf
	WebServer  WebServerConf
	ShotServer ShotServerConf
}
