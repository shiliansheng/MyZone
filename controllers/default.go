package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"myzone/models"
	"myzone/mzutils"
	"myzone/views"
	"net/http"
	"path/filepath"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/gorilla/websocket"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	var video models.Video
	var spider models.Spider
	c.Data["AllVideoCount"] = video.GetVideoCount()
	c.Data["VideoLastPubtime"] = video.GetLastPubtime()
	c.Data["AllSpiderCount"] = spider.GetCount()
	c.Data["SpiderLastAddtime"] = spider.GetLastAddtime()
	c.Data["HotVideos"] = video.GetVideoList(models.VSORT_HOT, models.VIDEO_LIST_LIMIT_HOME, 1).Data
	c.Data["UpdateVideos"] = video.GetVideoList(models.VSORT_UPDATE, models.VIDEO_LIST_LIMIT_HOME, 1).Data
	c.Data["RecommendVideo"] = video.GetRecommendList().Data
	c.TplName = "home.html"
}

func (c *MainController) Video() {
	ext := c.Ctx.Input.Param(":ext")
	if ext == "json" {
		var resp models.RespData
		var tempv models.Video
		method := c.Ctx.Request.Method
		switch method {
		case "GET":
			id := mzutils.Atoi(c.GetString("id"))
			action := c.GetString("action")
			if action == "videoplay" {
				resp = tempv.GetVideoPlayInfo(id)
			} else {
				resp = tempv.Get(id)
			}
		case "POST":
			id := mzutils.Atoi(c.GetString("id"))
			action := c.GetString("action")
			switch action {
			case "delete":
				id := mzutils.Atoi(c.GetString("id"))
				resp = new(models.Video).Delete(id)
			case "collect":
				id := mzutils.Atoi(c.GetString("id"))
				resp = new(models.Video).Collecting(id)
			case "play":
				id := mzutils.Atoi(c.GetString("id"))
				resp = new(models.Video).Play(id)
			case "cover":
				file, handeler, err := c.GetFile("file")
				id := mzutils.Atoi(c.GetString("id"))
				if err != nil {
					resp.Code = models.ERROR
					resp.Msg = "<span class=\"bg-danger\">获取文件失败！</span>"
					log.Println(resp.Msg, err)
				} else {
					handeler.Filename = mzutils.UniqueId() + filepath.Ext(handeler.Filename)
					resp = models.DownloadFile("pics", &file, handeler)
					if resp.Code == models.SUCCESS {
						msg := resp.Msg
						video := &models.Video{Id: id}
						video.Cover = resp.Data.(string)
						if resp = video.Update(video, "cover"); resp.Code == models.SUCCESS {
							resp.Msg = msg
							resp.Data = video.Cover
						}
					}
				}
			case "time":
				id := mzutils.Atoi(c.GetString("id"))
				time, _ := c.GetFloat("time")
				resp = new(models.Video).AddTimeNode(id, time)
			default:
				video := tempv.Get(id).Data.(models.Video)
				category := mzutils.Atoi(c.GetString("category"))
				video.Title = c.GetString("title")
				video.Path = c.GetString("path")
				video.Actorid = c.GetString("actorid")
				video.Tagid = c.GetString("tagid")
				video.Categoryid = category
				resp = tempv.Update(&video, "categoryid", "title", "path", "actorid", "tagid")
			}
		}
		c.Data["json"] = resp
		c.ServeJSON()
	} else {
		caterStr := c.Ctx.Input.Param(":category")
		category := 0
		limit := models.VIDEO_LIST_LIMIT_DEFAULT
		if caterStr == "" {
			caterStr = "0"
		}
		category = mzutils.Atoi(caterStr)
		if caterStr == "collect" {
			category = models.VIDEO_CATERGORY_COLLECT
			limit = models.VIDEO_LIST_LIMIT_COLLECT
		} else if caterStr == "all" {
			category = models.VIDEO_CATERGORY_ALL
		}
		page := mzutils.Atoi(c.GetString("page"))
		// limit := mzutils.Atoi(c.GetString("limit"))
		videos := new(models.Video).GetVideoList(models.VSORT_DEFAULT, limit, page, category)
		c.Data["Videos"] = videos.Data
		c.Data["Pager"] = views.SetPager("/video", caterStr, videos.Count, limit, page)
		c.Data["Categorys"] = new(models.Category).GetCategoryList(models.MODULE_VIDEO).Data
		c.Data["Category"] = caterStr
		c.Data["Module"] = "video"
		c.Data["ActorList"] = new(models.Actor).GetActorList().Data
		c.Data["TagList"] = new(models.Tag).GetTagList().Data

		c.TplName = "video.html"
	}
}

func (c *MainController) Manage() {
	c.Data["Module"] = "manage"
	c.Data["ModuleArr"] = models.ModuleValueArr
	c.TplName = "manage.html"
}

func (c *MainController) Category() {
	ext := c.Ctx.Input.Param(":ext")
	if ext == "json" {
		var resp models.RespData
		method := c.Ctx.Request.Method
		switch method {
		case "GET":
		case "POST":
			module := mzutils.Atoi(c.GetString("module"))
			title := c.GetString("title")
			path := c.GetString("path")
			resp = new(models.Category).Add(title, module, path)
		case "DELETE":
		}
		c.Data["json"] = resp
		c.ServeJSON()
	} else {
		page := mzutils.Atoi(c.GetString("page"))
		// limit := mzutils.Atoi(c.GetString("limit"))
		limit := 18
		c.Data["AllVideoCount"] = 1024
		c.Data["LastPubtime"] = time.Now().Format("2006-01-02 15:03:04")
		videos := new(models.Video).GetVideoList(models.VSORT_DEFAULT, limit, page, 0)
		c.Data["Videos"] = videos.Data
		c.Data["Pager"] = views.SetPager("video", "", videos.Count, limit, page)
		c.TplName = "video.html"
	}
}

func (c *MainController) Spider() {
	ext := c.Ctx.Input.Param(":ext")
	resp := models.RespData{}
	if ext == "json" {
		method := c.Ctx.Request.Method
		switch method {
		case "GET":
			mod := mzutils.Atoi(c.GetString("mod"))
			section := mzutils.Atoi(c.GetString("section"))
			category := mzutils.Atoi(c.GetString("category"))
			page := mzutils.Atoi(c.GetString("page"))
			filter := c.GetString("filter")
			resp = new(models.Spider).GetSpiderList(page, models.SPIDER_TABLE_LIMIT, mod, section, category, filter)
		case "POST":
		case "DELETE":
		}
		c.Data["json"] = resp
		c.ServeJSON()
	} else {
		resp = new(models.Spider).GetSpiderList(1, models.SPIDER_TABLE_LIMIT, models.SPIDER_MODULE_SHT, models.SHT_SECTION_DEFAULT, models.SHT_TYPE_DEFAILT)
		c.Data["ShtList"] = resp.Data
		c.Data["ShtListCount"] = resp.Count
		c.TplName = "spider/spider.html"
	}
}

func (c *MainController) SpiderSht() {
	ext := c.Ctx.Input.Param(":ext")
	if ext == "json" {
		ws, err := websocket.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(c.Ctx.ResponseWriter, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println("Cannot setup WebSocket connection:", err)
			return
		}
		section := mzutils.Atoi(c.GetString("section"))
		typeid := mzutils.Atoi(c.GetString("typeid"))
		day := mzutils.Atoi(c.GetString("day"))
		go new(models.Spider).SpiderSht(section, typeid, day)
		for {
			select {
			case msg := <-models.SpiderMsgChan:
				bytes, _ := json.Marshal(msg)
				if err := ws.WriteMessage(websocket.TextMessage, bytes); err != nil {
					log.Println("send message [", msg.Code, msg.Type, msg.Msg, "] failed:", err)
				}
			}
		}
	} else {
		c.TplName = "spider/sehuatang.html"
	}
}

// 获取video info
func (c *MainController) Spidervinfo() {
	if c.Ctx.Input.Param(":ext") == "json" {
		var resp models.RespData
		method := c.Ctx.Request.Method
		switch method {
		case "GET":
			action := c.GetString("action")
			switch action {
			case "cover":
				id := mzutils.Atoi(c.GetString("id"))
				url := c.GetString("url")
				target := c.GetString("target")
				switch target {
				case "javbus":
					resp = models.JavbusCoverDownUp(id, url)
				case "javdoe":
					resp = models.JavdoeCoverDownUp(id, url)
				case "javporn":
					resp = models.JavpornCoverDownUp(id, url)
				}
			case "coverdownload":
				id := mzutils.Atoi(c.GetString("id"))
				url := c.GetString("url")
				resp = models.WebCoverDownUp(id, url)
			}
		case "POST":

		case "DELETE":
		}
		c.Data["json"] = resp
		c.ServeJSON()
	}
}

func (c *MainController) Spidervideocover() {
	ws, err := websocket.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(c.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println("Cannot setup WebSocket connection:", err)
		return
	}
	go new(models.Video).SpiderCover()
	for {
		select {
		case msg := <-models.MsgChan:
			bytes, _ := json.Marshal(msg)
			if err := ws.WriteMessage(websocket.TextMessage, bytes); err != nil {
				log.Println("send message [", msg, "] failed:", err)
			}
		}
	}
}

func (c *MainController) Tag() {
	if c.Ctx.Input.Param(":ext") == "json" {
		var resp models.RespData
		method := c.Ctx.Request.Method
		switch method {
		case "GET":

		case "POST":
			action := c.GetString("action")
			switch action {
			case "add":
				obj := c.GetString("obj")
				name := c.GetString("name")
				switch obj {
				case "actor":
					resp = new(models.Actor).Add(name)
				case "tag":
					resp = new(models.Tag).Add(name)
				}
			case "tag":

			}
		case "DELETE":
		}
		c.Data["json"] = resp
		c.ServeJSON()
	} else {
		tagType := c.Ctx.Input.URL()[1:]
		tagName := ""
		var id string
		switch tagType {
		case "actor":
			id = c.GetString("id")
			tagName = new(models.Actor).GetActorName(mzutils.Atoi(id))
			c.Data["Videos"] = new(models.Video).GetVideoListByTagFilter(tagType, 9999, id).Data
		case "tag":
			id = c.GetString("id")
			tagName = new(models.Tag).GeTagName(mzutils.Atoi(id))
			c.Data["Videos"] = new(models.Video).GetVideoListByTagFilter(tagType, 9999, id).Data
		case "search":
			sstr := c.GetString("s")
			tagName = fmt.Sprintf("%s  查找结果为：", sstr)
			c.Data["Videos"] = new(models.Video).SearchVideo(sstr).Data
		}
		c.Data["ActorList"] = new(models.Actor).GetActorList().Data
		c.Data["TagList"] = new(models.Tag).GetTagList().Data
		c.Data["Categorys"] = new(models.Category).GetCategoryList(models.MODULE_VIDEO).Data
		c.Data["TagName"] = tagName
		c.TplName = "tag.html"
	}
}

func (c *MainController) Downloadfile() {
	ext := c.Ctx.Input.Param(":ext")
	if ext == "json" {
		fbelong := c.GetString("belong")
		file, handeler, err := c.GetFile("file")
		resp := models.NewRespData()
		if err != nil {
			resp.Code = models.ERROR
			resp.Msg = "get file failed"
			log.Println(resp.Msg, err)
		} else {
			*resp = models.DownloadFile(fbelong, &file, handeler)
		}
		c.Data["json"] = *resp
		c.ServeJSON()
	} else {

	}
}
