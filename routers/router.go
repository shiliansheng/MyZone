package routers

import (
	"myzone/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/video", &controllers.MainController{}, "*:Video")
	beego.Router("/video/?:category", &controllers.MainController{}, "*:Video")
	beego.Router("/manage", &controllers.MainController{}, "*:Manage")
	beego.Router("/category", &controllers.MainController{}, "*:Category")
	beego.Router("/tag", &controllers.MainController{}, "*:Tag")
	beego.Router("/actor", &controllers.MainController{}, "*:Tag")
	beego.Router("/search", &controllers.MainController{}, "*:Tag")
	beego.Router("/spider", &controllers.MainController{}, "*:Spider")
	beego.Router("/spider/sht", &controllers.MainController{}, "*:SpiderSht")
	beego.Router("/spidervinfo", &controllers.MainController{}, "*:Spidervinfo")
	beego.Router("/spidervideocover", &controllers.MainController{}, "*:Spidervideocover")
	beego.Router("/downloadfile", &controllers.MainController{}, "*:Downloadfile")
	// beego.Router("/picture/:type", &controllers.MainController{}, "*:Picture")
	// beego.Router("/video/:type", &controllers.MainController{}, "*:Video")
	// beego.Router("/audio/:type", &controllers.MainController{}, "*:Audio")
	// beego.Router("/novel/:type", &controllers.MainController{}, "*:Novel")
	// beego.Router("/picture", &controllers.MainController{}, "*:Picture")
	// beego.Router("/video", &controllers.MainController{}, "*:Video")
	// beego.Router("/audio", &controllers.MainController{}, "*:Audio")
	// beego.Router("/novel", &controllers.MainController{}, "*:Novel")
	// beego.Router("/home", &controllers.MainController{}, "*:Home")
	// beego.Router("/album", &controllers.MainController{}, "*:Album")
	// beego.Router("/typer", &controllers.MainController{}, "*:Typer")
	// beego.Router("/videoinfo", &controllers.MainController{}, "*:Videoinfo")
	// beego.Router("/initer", &controllers.MainController{}, "*:Initer")
	// beego.Router("/actor", &controllers.MainController{}, "*:Actor")
	// beego.Router("/tag", &controllers.MainController{}, "*:Tag")
	// beego.Router("/search", &controllers.MainController{}, "*:Search")
	// beego.Router("/downloader", &controllers.MainController{}, "*:Downloader")
}
