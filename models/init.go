package models

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
)

var (
	Orm              orm.Ormer
	videoDeletePath  string
	videoUpdatePath  string
	videoStorePath   string
	CategoryIdMTitle map[int]string
	ScreenshootList  []Screenshot
)

const (
	STORE_PATH string = ".\\static\\"
)

type RespBase struct {
	Code  int         `json:"code"`
	Error error       `json:"-"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data,omitempty"`
}

// 默认返回Code=ERROR
//
//	@return [*RespBase]
func NewRespBase() *RespBase {
	return &RespBase{
		Code: ERROR,
		Msg:  "",
	}
}

type RespData struct {
	Code  int         `json:"code"`
	Error error       `json:"-"`
	Msg   string      `json:"msg"`
	Count int         `json:"count,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

// 默认返回Code=ERROR
//
//	@return [*RespData]
func NewRespData() *RespData {
	return &RespData{
		Code: ERROR,
		Msg:  "",
	}
}

func init() {
	dbhost, _ := beego.AppConfig.String("dbhost")
	dbport, _ := beego.AppConfig.String("dbport")
	dbuser, _ := beego.AppConfig.String("dbuser")
	dbpassword, _ := beego.AppConfig.String("dbpassword")
	dbname, _ := beego.AppConfig.String("dbname")

	videoDeletePath, _ = beego.AppConfig.String("videodeletepath")
	videoUpdatePath, _ = beego.AppConfig.String("videoupdatepath")
	videoStorePath, _ = beego.AppConfig.String("videostorepath")

	if dbport == "" {
		dbport = "3306"
	}
	dbConnStr := dbuser + ":" + dbpassword + "@tcp(" + dbhost + ":" + dbport + ")/" + dbname + "?charset=utf8&loc=Asia%2FShanghai"

	// 设置日志
	logfile, err := os.OpenFile("./log/"+time.Now().Format("20060102")+".log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println("create log file failed:", err)
	}
	multiOutput := io.MultiWriter(logfile, os.Stdout)
	log.SetOutput(multiOutput)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 开启调试日志
	// orm.Debug = true
	// orm.DebugLog = orm.NewLog(multiOutput)

	if orm.RegisterDataBase("default", "mysql", dbConnStr) != nil {
		fmt.Println("register database(", dbConnStr, ") failed:", err)
	}

	// 注册模型 new Model()
	orm.RegisterModel(new(Video), new(Category), new(Actor), new(Tag), new(Spider), new(Record))

	Orm = orm.NewOrm()

	new(Category).SetMap()

	SetScreenshotList(&ScreenshootList)
	// new(Video).SetDuration()
}

func Update() {
	tiker := time.NewTicker(time.Minute * 5)
	defer tiker.Stop()
	for {
		addLocalVideo()
		<-tiker.C
	}
}
