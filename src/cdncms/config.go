package main

import (
	"encoding/json"
	"os"
)

type Config struct {
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

var Cf Config

func ReadConfig(name string) error {
	r, err := os.Open(name)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(r)
	err = decoder.Decode(&Cf)
	if err != nil {
		return err
	}
	return nil
}
