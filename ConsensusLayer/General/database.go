package General

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

var db *sql.DB

const (
	userName = "root"
	password = "19980618"
	ip       = "127.0.0.1"
	port     = "3306"
	dbName   = "tinyChain"
)

var GormDb *gorm.DB

func Connect() {
	var err error
	path := strings.Join([]string{
		userName, ":",
		password, "@tcp(",
		ip, ":",
		port, ")/",
		dbName,
		"?charset=utf8&parseTime=true&loc=Local"}, "")
	GormDb, err = gorm.Open(mysql.Open(path), nil)
	if err != nil {
		fmt.Println("Database connect fail", err)
		return
	}
	fmt.Println("Database connect success")
}

func ClearTable(table string) {
	GormDb.Exec("truncate table " + table)
}
