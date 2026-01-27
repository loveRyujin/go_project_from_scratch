package main

import (
	"log"

	"github.com/loveRyujin/geeorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	e, err := geeorm.NewEngine("sqlite3", "gee.db")
	if err != nil {
		panic(err)
	}
	defer e.Close()

	sess := e.Session()

	_, _ = sess.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = sess.Raw("CREATE TABLE User(Name text);").Exec()
	result, err := sess.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	if err == nil {
		affected, _ := result.RowsAffected()
		log.Println(affected)
	}
	row := sess.Raw("SELECT Name FROM User LIMIT 1").QueryRow()
	var name string
	if err := row.Scan(&name); err == nil {
		log.Println(name)
	}
}
