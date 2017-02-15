package main

type SqlServerConf struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Name     string
}

type GlobalConf struct {
	PLATERRORINTERFACE  string
	PLATINTERFACE       string
	WebServerBaseDir    string
	ShotServerBaseDir   string
	UploadServerBaseDir string
	LocalSqlServer      SqlServerConf
	GlCenSqlServer      SqlServerConf
	OnLineSqlServer     SqlServerConf
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
	FfmpegPath   string
	Threadnum    uint32
	LoopInterval uint32
}

type UploadServerConf struct {
	VendownName  string
	Threadnum    uint32
	LoopInterval uint32
}

type ServerConfig struct {
	Global       GlobalConf
	WebServer    WebServerConf
	ShotServer   ShotServerConf
	UploadServer UploadServerConf
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
