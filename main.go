package main

import (
	"myzone/models"
	_ "myzone/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	go func() { models.Update() }()
	// go func() { models.OperateVideoActor() }()
	beego.Run()
}
