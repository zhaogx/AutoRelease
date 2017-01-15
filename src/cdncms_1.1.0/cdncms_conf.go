package main

type CdncmsConfig struct {
	Host  string
	Port  string
	Mysql []struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	Vendown_name  string
	AppType       string
	Listen_path   string
	Path          string
	Errcbhost     string
	Errcbport     string
	Errcbhost_3rd string
	Errcbport_3rd string
	Threadnum     int
}
