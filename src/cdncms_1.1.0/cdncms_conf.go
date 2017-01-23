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
	FfmpegPath string
	Threadnum  uint32
}

type ServerConfig struct {
	Global     GlobalConf
	WebServer  WebServerConf
	ShotServer ShotServerConf
}

/*
type ServerConfig2 struct {
	Global struct {
		WebServerBaseDir  string
		ShotServerBaseDir string
		LocalSqlServer    struct {
			Host     string
			Port     uint16
			User     string
			Password string
			Name     string
		}
	}
	WebServer struct {
		Host          string
		Port          string
		VendownName   string
		Errcbhost     string
		Errcbport     string
		Threadnum     uint32
		CheckInterval uint32
	}
	ShotServer struct {
		ShotPath   string
		FfmpegPath string
	}
}
*/
