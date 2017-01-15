package main

import (
	"database/sql"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbm0 = map[string]string{
		"host": "10.8.16.246",
		//"host":   "218.24.66.80",
		"user":   "cdn",
		"pswd":   "voole.com",
		"dbname": "voole_cdn",
	}

	dbm1 = map[string]string{
		"host":   "123.125.149.11",
		"user":   "cdn",
		"pswd":   "voole.com",
		"dbname": "voole_cdn",
	}

	dbm2 = map[string]string{
		"host":   "119.97.153.69",
		"user":   "cdn",
		"pswd":   "voole.com",
		"dbname": "voole_cdn",
	}
	dbm3 = map[string]string{
		"host":   "112.25.53.202",
		"user":   "cdn",
		"pswd":   "voole.com",
		"dbname": "voole_cdn",
	}
)

var mydb *sql.DB
var mydb0 *sql.DB
var mydb1 *sql.DB
var mydb2 *sql.DB
var mydb3 *sql.DB

func mysql_init(host string, user string, pswd string, dbname string) *sql.DB {
	dburl := user + ":" + pswd + "@tcp(" + host + ":3306)/" + dbname + "?allowOldPasswords=1"
	db, err := sql.Open("mysql", dburl)
	if err != nil {
		fmt.Println("Connect Database Error:", err)
		log.Fatal("Connect Database Error:", err)
		panic(err)
	}
	return db
}

func mysql_sync() {
	for {
		stmt, err := mydb.Prepare("SELECT md5, name, filesize, cpid, oldid, st_sync from content_notify_hg where st_sync != 0x1 and filesize is not null and name is not null")
		if err != nil {
			fmt.Println("Insert error:", err)
			continue
		}

		res, err := stmt.Query()
		if err != nil {
			fmt.Println("Query error:", err, res)
			stmt.Close()
			continue
		}

		for res.Next() {
			var fid string
			var name string
			var filesize string
			var cpid string
			var oldid string
			var st_sync int
			var st_sync_tmp int

			res.Scan(&fid, &name, &filesize, &cpid, &oldid, &st_sync)
			st_sync_tmp = st_sync
			fext := path.Ext(name)
			var format = 0

			if fext == ".zip" {
				format = 0
			}
			if fext == ".apk" {
				format = 0
			}
			if fext == ".ts" {
				format = 6
			}
			if fext == ".mp4" {
				format = 7
			}

			sql_str := fmt.Sprintf("INSERT INTO resource (fid, rawname, filesize, format, cpid, oldid, status) values ('%s', '%s', '%s', %d, '%s', '%s', 20)", fid, name, filesize, format, cpid, oldid)
			if st_sync&0x01 == 0 {
				if 0 == mysql_write(mydb, sql_str) {
					st_sync = st_sync | 0x01
				}
			}

			if st_sync != st_sync_tmp {
				stmt2, err := mydb.Prepare("UPDATE content_notify_hg set st_sync = ?")
				if err != nil {
					fmt.Println("Update error!")
				}

				res, err := stmt2.Exec(st_sync)
				if err != nil {
					fmt.Println("Update st_sync error:", err, res)
				}
				stmt2.Close()
			}
		}

		stmt.Close()
		time.Sleep(time.Second * 5)
	}
}

func mysql_sync_local(md5 string) {
	fmt.Println("Insert into resource 2. md5:", md5)
	stmt, err := mydb.Prepare("SELECT md5, name, filesize, cpid, oldid, st_sync,platinterface_3rd  from content_notify_hg where st_sync != 0x1 and md5 = ?")
	if err != nil {
		fmt.Println("Insert error:", err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Query(md5)
	if err != nil {
		fmt.Println("Query error:", err, res)
		stmt.Close()

	}
	fmt.Println(res)

	for res.Next() {

		var fid string
		var name string
		var filesize string
		var cpid string
		var oldid string
		var st_sync int
		var st_sync_tmp int
		var platinterface_3rd string

		res.Scan(&fid, &name, &filesize, &cpid, &oldid, &st_sync, &platinterface_3rd)
		st_sync_tmp = st_sync
		fext := path.Ext(name)
		var format = 0

		if fext == ".zip" {
			format = 0
		}
		if fext == ".apk" {
			format = 0
		}
		if fext == ".ts" {
			format = 6
		}
		if fext == ".mp4" {
			format = 7
		}
		if fext == ".mp3" {
			format = 9
		}
		if fext == ".aac" {
			format = 10
		}
		if fext == ".wma" {
			format = 11
		}
		if fext == ".wav" {
			format = 21
		}
		if fext == ".flac" {
			format = 22
		}

		sql_str := fmt.Sprintf("INSERT INTO resource (fid, rawname, filesize, format, cpid, oldid, status,platinterface_3rd) values ('%s', '%s', '%s', %d, '%s', '%s', 20,'%s')", fid, name, filesize, format, cpid, oldid, platinterface_3rd)

		if st_sync&0x01 == 0 {

			if 0 == mysql_write(mydb, sql_str) {

				st_sync = st_sync | 0x01
			}
		}

		if st_sync != st_sync_tmp {

			stmt2, err := mydb.Prepare("UPDATE content_notify_hg set st_sync = ? where md5 = ?")
			if err != nil {
				fmt.Println("Update error!")
			}

			res, err := stmt2.Exec(st_sync, md5)
			if err != nil {
				fmt.Println("Update st_sync error:", err, res)
			}
			stmt2.Close()

			timestamp := time.Now().Unix()
			fmt.Println(timestamp)
			tm := time.Unix(timestamp, 0)
			//fmt.Println(tm.Format("2006-01-02 03:04:05"))
			str := tm.Format("2006-01-02 15:04:05")
			fmt.Println(str)

			stmt2, err = mydb.Prepare("UPDATE resource set IssuedDate = ?, rawname=?, filesize=?, format=?, cpid=?, oldid=?,platinterface_3rd=? where fid = ?")
			if err != nil {
				fmt.Println("Update error!")
			}

			res, err = stmt2.Exec(str, name, filesize, format, cpid, oldid, platinterface_3rd, md5)
			if err != nil {
				fmt.Println("Update st_sync error:", err, res)
			}
			stmt2.Close()
		}
	}

}

func mysql_write(db *sql.DB, sql string) int {
	fmt.Println(sql)

	stmt, err := db.Prepare(sql)
	if err != nil {
		fmt.Println("Insert error:", err)
		return -1
	}
	defer stmt.Close()

	res, err := stmt.Exec()
	if err != nil {
		fmt.Println("Insert error:", err, res)
		var err_str string
		err_str = err.Error()

		if len(err_str) > 0 && strings.Index(err_str, "Duplicate entry") > 0 {
			return 0
		}
		return -1
	}
	return 0
}
