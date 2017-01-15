package utils

import (
	"database/sql"
	"fmt"
)

type VooleSqlMgmt struct {
	host   string
	port   uint16
	user   string
	pwd    string
	dbname string
	db     *sql.DB
}

func VooleSqlInit(host string, port uint16, user string, pwd string, dbname string) (*VooleSqlMgmt, error) {
	if port <= 0 {
		port = 3306
	}

	mgmt := &VooleSqlMgmt{
		host:   host,
		port:   port,
		user:   user,
		pwd:    pwd,
		dbname: dbname}

	dburl := fmt.Sprintf("%s:%s@tcp(%s:%u)/%s?allowOldPasswords=1",
		mgmt.user, mgmt.pwd, mgmt.host, mgmt.port, mgmt.dbname)

	db, err := sql.Open("mysql", dburl)
	if err != nil {
		VLOG(VLOG_ERROR, "sql open failed:%s", dburl)
		return nil, err
	}
	mgmt.db = db

	return mgmt, err
}

func VooleSqlQuery() {
	return
}
