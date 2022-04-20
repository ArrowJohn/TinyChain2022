package General

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"strings"
)

type DatabaseConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	DBName   string `json:"dbName"`
}

var databaseConfig DatabaseConfig
var GormDb *gorm.DB

func ReadConfig() {
	jsonFile, err := os.Open("ConsensusLayer/General/MysqlConfig.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	_ = json.Unmarshal(byteValue, &databaseConfig)
}

func Connect() {
	//ReadConfig()
	var err error
	path := strings.Join([]string{
		"admin", ":",
		"12345678", "@tcp(",
		"elec5620.ctonlrsgkoqg.ap-northeast-1.rds.amazonaws.com", ":",
		"3306", ")/",
		"tinyChain",
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
