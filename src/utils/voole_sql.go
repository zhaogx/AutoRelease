package utils

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type VooleSqlMgmt struct {
	host   string
	port   uint16
	user   string
	pwd    string
	dbname string
	db     *sql.DB
}

func VooleSqlInit(host string, port uint16, user, pwd, dbname string) (*VooleSqlMgmt, error) {
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

type VooleQueryCb func(*sql.Rows) int

func (this *VooleSqlMgmt) VooleSqlQueryStmt(cb VooleQueryCb, sql string, args ...interface{}) error {
	stmt, err := this.db.Prepare(sql)
	defer stmt.Close()

	if err != nil {
		VLOG(VLOG_ERROR, "prepare failed![%s][%s]", sql, err)
		return err
	}

	res, err := stmt.Query(args...)
	defer res.Close()

	if err != nil {
		VLOG(VLOG_ERROR, "query failed![%s][%s]", sql, err)
		return err
	}
	cb(res)
	return err
}
