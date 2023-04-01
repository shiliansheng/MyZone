package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"myzone/mzutils"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
)

type Video struct {
	Id            int       `json:"id"`
	Title         string    `json:"title"`
	Cover         string    `json:"cover" orm:"default('')"`
	Path          string    `json:"path"`
	Duration      string    `json:"duration" orm:"default('00:00:00')"`
	Categoryid    int       `json:"categoryid" orm:"default(0)"` // 类型ID
	CategoryTitle string    `json:"actegorytitle" orm:"-"`
	Actorid       string    `json:"actorid" orm:"default('')"` // 存在actorid
	Actors        []Actor   `json:"actors" orm:"-"`
	Tagid         string    `json:"tagid" orm:"default('')"` // 存放tagid
	Tags          []Tag     `json:"tags" orm:"-"`
	Timenode      string    `json:"timenode" orm:"default('[]')"`
	Timenodes     []float64 `json:"timenodes" orm:"-"`
	Serialid      int       `json:"serialid" orm:"default(0)"`
	Collect       int       `json:"collect" orm:"default(0)"`
	View          int       `json:"view" orm:"default(0)"`
	Pubtime       string    `json:"pubtime"`
	State         int       `json:"state,omitempty" orm:"default(0)"`
}

type VideoPlay struct {
	Id           int          `json:"id"`
	Title        string       `json:"title"`
	Cover        string       `json:"cover" orm:"default('')"`
	Path         string       `json:"path"`
	Duration     string       `json:"duration" orm:"default('00:00:00')"`
	Actors       []Actor      `json:"actors" orm:"-"`
	RelateVideos []Video      `json:"relateVideos"`
	Tags         []Tag        `json:"tags" orm:"-"`
	TimeNodes    []float64    `json:"timenodes"`
	Collect      int          `json:"collect" orm:"default(0)"`
	View         int          `json:"view" orm:"default(0)"`
	Screenshots  []Screenshot `json:"screenshoots"`
	Pubtime      string       `json:"pubtime"`
}

type VideoRecommendRecord struct {
	DateMap      map[string][]int `json:"dateMap"`
	RecommendMap map[int]int      `json:"recommendMap"`
	Threshold    int              `json:"threshold"`
}

const (
	VSORT_DEFAULT              = 0
	VSORT_HOT              int = 1
	VSORT_UPDATE           int = 2
	VSORT_COLLECT          int = 3
	VIDEO_REALTE_THRESHOLD     = 8 // video play relate video maximum number

	VIDEO_CATERGORY_COLLECT int = 102

	VIDEO_LIST_LIMIT_MAX     int = 9999
	VIDEO_LIST_LIMIT_DEFAULT int = 18
	VIDEO_LIST_LIMIT_HOME    int = 8
	VIDEO_LIST_LIMIT_COLLECT int = 24

	VIDEO_RECOMMEND_LIMIT int = 12

	recommendJsonPath string = `./static/json/video-recommend-list.json`
	javbus_baseurl    string = `https://www.seejav.pw`
	javdock_baseurl   string = `https://www2.javdock.com/video`
)

func (Video) TableName() string {
	return "video"
}

func (this Video) GetVideoCount() int64 {
	count, _ := Orm.QueryTable(this.TableName()).Filter("state", VALID).Count()
	return count
}

func (this Video) GetLastPubtime() string {
	videos := []Video{}
	Orm.QueryTable(this.TableName()).Filter("state", VALID).OrderBy("-pubtime").Limit(1, 0).All(&videos)
	return videos[0].Pubtime
}

func (this Video) setVideosActorTag(videos []Video) {
	for i := range videos {
		this.setVideoActorTag(&videos[i])
	}
}

func (this Video) setVideoActorTag(video *Video) {
	actorIds := []string{}
	tagIds := []string{}
	if err := json.Unmarshal([]byte(video.Actorid), &actorIds); err != nil || len(actorIds) == 0 {
		// log.Println("[SET VIDEO TAG] unmarshal json of actor ID failed [", video.Id, video.Actorid, "]", err)
	} else {
		video.Actors = new(Actor).GetActorList(actorIds...).Data.([]Actor)
	}
	if err := json.Unmarshal([]byte(video.Tagid), &tagIds); err != nil || len(tagIds) == 0 {
		// log.Println("[SET VIDEO TAG] unmarshal json of tag ID failed [", video.Id, video.Tagid, "]", err)
		return
	}
	video.Tags = new(Tag).GetTagList(tagIds...).Data.([]Tag)
}

func (this Video) GetVideoList(sort int, limit, page int, categoryid ...int) RespData {
	resp := NewRespData()
	log.Println("[VIDEO LIST] category[", categoryid, "] page[", page, "] limit[", limit, "] sort[", sort, "]")
	videos := []Video{}
	seter := Orm.QueryTable(this.TableName()).Filter("state", VALID)

	if len(categoryid) != 0 && categoryid[0] != CATEGORY_ALL {
		switch categoryid[0] {
		case CATEGORY_DEFAULT:
			seter = seter.Filter("categoryid__in", categoryid)
		case VIDEO_CATERGORY_COLLECT:
			seter = seter.Filter("collect__gt", 0)
			limit = 24
			sort = VSORT_COLLECT
		default:
			seter = seter.Filter("categoryid__in", categoryid)
		}
	}

	switch sort {
	case VSORT_HOT:
		seter = seter.OrderBy("-view", "id")
	case VSORT_UPDATE:
		seter = seter.OrderBy("-pubtime", "-id")
	case VSORT_COLLECT:
		seter = seter.OrderBy("-collect", "id")
	}

	count, _ := seter.Count()
	seter = seter.Limit(limit, (page-1)*limit)

	if _, err := seter.All(&videos); err != nil {
		resp.Error = err
		resp.Msg = fmt.Sprint("[ERROR]", err)
		log.Println(resp.Msg)
		return *resp
	}

	for i := range videos {
		this.setVideoActorTag(&videos[i])
		videos[i].CategoryTitle = CategoryIdMTitle[videos[i].Categoryid]
	}

	resp.Code = SUCCESS
	resp.Count = int(count)
	resp.Data = videos
	return *resp
}

func (this Video) GetVideoListByTagFilter(tagName string, limit int, id string) RespData {
	resp := NewRespData()
	log.Println("[VIDEO LIST] filter by tag with tag name [", tagName, "] limt:", limit)
	videos := []Video{}
	seter := Orm.QueryTable(this.TableName()).Filter("state", VALID).Filter(tagName+"id__icontains", "\""+id+"\"").OrderBy("-view", "id").Limit(limit)

	if _, err := seter.All(&videos); err != nil {
		resp.Error = err
		resp.Msg = fmt.Sprint("[ERROR]", err)
		log.Println(resp.Msg)
		return *resp
	}
	for i := range videos {
		this.setVideoActorTag(&videos[i])
		videos[i].CategoryTitle = CategoryIdMTitle[videos[i].Categoryid]
	}

	resp.Code = SUCCESS
	resp.Data = videos
	return *resp
}

func (this Video) SearchVideo(key string) RespData {
	key = strings.TrimSpace(key)
	resp := NewRespData()
	log.Println("[VIDEO SEARCH] search key:", key)
	videos := []Video{}
	seter := Orm.QueryTable(this.TableName()).Filter("state", VALID).Filter("title__icontains", key)

	count, _ := seter.Count()
	if _, err := seter.All(&videos); err != nil {
		resp.Error = err
		resp.Msg = fmt.Sprint("[ERROR]", err)
		log.Println(resp.Msg)
		return *resp
	}

	for i := range videos {
		this.setVideoActorTag(&videos[i])
		videos[i].CategoryTitle = CategoryIdMTitle[videos[i].Categoryid]
	}

	resp.Code = SUCCESS
	resp.Count = int(count)
	resp.Data = videos
	return *resp
}

// 获取视频播放相关信息
//
//	@param  id [int]
//	@return [RespData] RespData.Data = VideoPlayInfo
func (this Video) GetVideoPlayInfo(id int) RespData {
	log.Println("[VIDEO] get video play info Id:", id)
	video := &Video{Id: id}
	resp := NewRespData()
	if err := Orm.Read(video); err != nil {
		resp.Msg = fmt.Sprintf("获取视频[ID:%d]信息失败.", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	this.setVideoActorTag(video)
	vplay := VideoPlay{
		Id:       video.Id,
		Title:    video.Title,
		Cover:    video.Cover,
		Path:     video.Path,
		Duration: video.Duration,
		Collect:  video.Collect,
		View:     video.View,
		Pubtime:  video.Pubtime,
		Actors:   video.Actors,
		Tags:     video.Tags,
	}
	varr := []Video{}
	idarr := []string{}
	idMint := map[int]int{id: 1}
	var tresp RespData

	if json.Unmarshal([]byte(video.Actorid), &idarr); len(idarr) != 0 {
		tvarr := []Video{}
		for _, id := range idarr {
			tresp = this.GetVideoListByTagFilter("actor", VIDEO_LIST_LIMIT_MAX, id)
			tvarr = append(tvarr, tresp.Data.([]Video)...)
		}
		for _, v := range tvarr {
			if _, ok := idMint[v.Id]; !ok {
				varr = append(varr, v)
				idMint[v.Id] = 1
			}
		}
	}
	// rvcnt := len(varr)
	// actorVcnt := rvcnt
	idarr = []string{}
	if json.Unmarshal([]byte(video.Tagid), &idarr); len(idarr) != 0 {
		tvarr := []Video{}
		for _, id := range idarr {
			tresp = this.GetVideoListByTagFilter("tag", VIDEO_REALTE_THRESHOLD, id)
			tmpVArr := tresp.Data.([]Video)
			if len(tmpVArr) > 3 {
				tmpVArr = tmpVArr[:3]
			}
			tvarr = append(tvarr, tmpVArr...)
		}
		for _, v := range tvarr {
			if _, ok := idMint[v.Id]; !ok {
				varr = append(varr, v)
				idMint[v.Id] = 1
				// rvcnt++
			}
		}
	}
	// allVcnt := len(varr)
	// if allVcnt > VIDEO_REALTE_THRESHOLD {
	// 	tagVcnt := allVcnt - actorVcnt
	// 	if tagVcnt > VIDEO_REALTE_THRESHOLD/2 {
	// 		varr = varr[:VIDEO_REALTE_THRESHOLD] // 引用类型交换
	// 	} else if actorVcnt > VIDEO_REALTE_THRESHOLD/2 {
	// 		tagVarr := varr[actorVcnt:]
	// 		actorVcnt = tagVcnt + actorVcnt - VIDEO_REALTE_THRESHOLD
	// 		varr = varr[:actorVcnt]
	// 		varr = append(varr, tagVarr...)
	// 	}
	// }
	json.Unmarshal([]byte(video.Timenode), &(vplay.TimeNodes))
	vplay.RelateVideos = varr
	shList := []Screenshot{}
	for _, sc := range ScreenshotList {
		if !strings.HasPrefix(sc.Title, fmt.Sprintf("[%d]", id)) {
			continue
		}
		shList = append(shList, sc)
	}
	vplay.Screenshots = shList

	resp.Code = SUCCESS
	resp.Msg = "获取视频播放信息成功！"
	resp.Data = vplay
	return *resp
}

// 获取推荐列表
//
//	@return [RespData] RespData.Data = RecommendVideoList
func (this Video) GetRecommendList(date string, refresh bool) RespData {
	resp := NewRespData()
	log.Println("[VIDEO LIST] get video recommended list...")
	recRecord := VideoRecommendRecord{}
	videos := []Video{}
	if refresh {
		idRawlist := []int{}
		idlist := []int{}
		if _, err := Orm.Raw("SELECT id FROM video where state != 1").QueryRows(&idRawlist); err != nil {
			log.Println("read table video id list failed:", err)
			return *resp
		}
		videoCount := len(idRawlist)
		for _, v := range idRawlist {
			if recRecord.RecommendMap[v] < recRecord.Threshold {
				idlist = append(idlist, v)
			}
		}
		for i := 0; i < VIDEO_RECOMMEND_LIMIT; {
			idx := mzutils.RandomInt(videoCount)
			idlist = append(idlist, idRawlist[idx])
			i++
		}
		Orm.QueryTable(this.TableName()).Filter("id__in", idlist).All(&videos)
		for i := range videos {
			this.setVideoActorTag(&videos[i])
			videos[i].CategoryTitle = CategoryIdMTitle[videos[i].Categoryid]
		}
		resp.Code = SUCCESS
		resp.Data = videos
		return *resp
	}

	var jsonfile *os.File
	var fileExist bool
	var err error
	if fileExist, _ = mzutils.PathExists(recommendJsonPath); !fileExist {
		if jsonfile, err = os.OpenFile(recommendJsonPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err != nil {
			log.Println("create video recommended json file [", recommendJsonPath, "] failed:", err)
			return *resp
		}
	}

	// 读取 video recommend json file
	if bytes, err := ioutil.ReadFile(recommendJsonPath); err != nil {
		log.Println("read video recommended json file [", recommendJsonPath, "] failed:", err)
		return *resp
	} else {
		json.Unmarshal(bytes, &recRecord)
	}
	if recRecord.Threshold == 0 {
		recRecord.Threshold = 1
	}
	// 判断当天是否已经推荐过，已经推荐过直接取出id list，否则进行随机寻找 推荐次数在在阈值之下的 id 添加到 list
	if recRecord.DateMap == nil {
		recRecord.DateMap = make(map[string][]int)
		recRecord.RecommendMap = map[int]int{}
	}
	nowdate := mzutils.NowDate()
	if date != "" {
		nowdate = date
		if _, ok := recRecord.DateMap[nowdate]; !ok {
			resp.Msg = fmt.Sprintf("日期[%s]未找到视频推荐列表", date)
			log.Println(resp.Msg)
			return *resp
		}
	}
	if _, ok := recRecord.DateMap[nowdate]; !ok {
		recRecord.DateMap[nowdate] = []int{}
		idRawlist := []int{}
		if _, err = Orm.Raw("SELECT id FROM video where state != 1").QueryRows(&idRawlist); err != nil {
			log.Println("read table video id list failed:", err)
			return *resp
		}
		idlist := []int{}
		for _, v := range idRawlist {
			if recRecord.RecommendMap[v] < recRecord.Threshold {
				idlist = append(idlist, v)
			}
		}
		videoCount := len(idlist)
		for i := 0; i < VIDEO_RECOMMEND_LIMIT; {
			idx := mzutils.RandomInt(videoCount)
			recTimes := 0
			if v, ok := recRecord.RecommendMap[idlist[idx]]; ok && v >= recRecord.Threshold {
				continue
			} else if ok {
				recTimes = v
			}
			recRecord.DateMap[nowdate] = append(recRecord.DateMap[nowdate], idlist[idx])
			recRecord.RecommendMap[idlist[idx]] = recTimes + 1
			i++
		}
		if videoCount < 2*VIDEO_RECOMMEND_LIMIT {
			recRecord.Threshold++
		}
	}
	Orm.QueryTable(this.TableName()).Filter("state", VALID).Filter("id__in", recRecord.DateMap[nowdate]).All(&videos)
	for i := range videos {
		this.setVideoActorTag(&videos[i])
		videos[i].CategoryTitle = CategoryIdMTitle[videos[i].Categoryid]
	}
	if fileExist {
		if jsonfile, err = os.OpenFile(recommendJsonPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err != nil {
			log.Println("open video recommended json file [", recommendJsonPath, "] failed:", err)
			return *resp
		}
	}
	bytes, _ := json.Marshal(recRecord)
	jsonfile.Write(bytes)
	defer jsonfile.Close()
	resp.Code = SUCCESS
	resp.Data = videos
	return *resp
}

// 获取视频
//
//	@param  id [int]
//	@return [RespData] RespData.Data = video
func (this Video) Get(id int) RespData {
	log.Println("[VIDEO] get video info Id:", id)
	resp := NewRespData()
	video := &Video{Id: id}
	if err := Orm.Read(video); err != nil {
		resp.Msg = fmt.Sprintf("获取视频[ID:%d]信息失败.", id)
		log.Println(resp.Msg)
		return *resp
	}
	resp.Data = *video
	resp.Code = SUCCESS
	return *resp
}

func (this Video) Add(videos ...Video) RespBase {
	resp := NewRespBase()
	log.Println("[VIDEO ADD]", videos)
	for _, video := range videos {
		if video.Path == "" {
			resp.Msg = fmt.Sprintf("video[%s] insert failed: path is space!\n", video.Title)
			log.Println(resp.Msg)
			return *resp
		}
		if _, err := Orm.Insert(&video); err != nil {
			resp.Msg = fmt.Sprintf("video[%s] insert failed: %v\n", video.Title, err)
			log.Println(resp.Msg)
			return *resp
		}
		resp.Msg += fmt.Sprintf("video[%s] insert success!\n", video.Title)
		log.Println(resp.Msg)
	}
	resp.Code = SUCCESS
	return *resp
}

func (m Video) Update(video *Video, cols ...string) RespData {
	log.Println("[VIDEO UPDATE]", *video, " cols:", cols)
	resp := NewRespData()

	if video.Title != "" {
		rawv := &Video{Id: video.Id}
		Orm.Read(rawv)
		rawTitle := rawv.Path[strings.LastIndex(rawv.Path, "\\")+1 : strings.LastIndex(rawv.Path, ".")]
		if video.Title != rawTitle && video.Path == rawv.Path {
			prepath := rawv.Path
			newpath := strings.Replace(prepath, rawTitle, video.Title, -1)
			log.Println("need to change path:", prepath, "-->", newpath)
			if err := os.Rename("."+prepath, "."+newpath); err == nil {
				video.Path = newpath
			} else {
				log.Println(err)
				return *resp
			}
		}
	}
	if _, err := Orm.Update(video, cols...); err != nil {
		resp.Msg = "update video failed"
		log.Println(resp.Msg, err)
	} else {
		resp.Code = SUCCESS

	}
	return *resp
}

func (m Video) Delete(id int) RespData {
	log.Println("[VIDEO DELETE]", id)
	resp := NewRespData()
	video := Video{Id: id}
	var oldpath, newpath string
	if err := Orm.Read(&video); err != nil {
		resp.Msg = fmt.Sprint("Not Found Video By ID:", id)
		log.Println(resp.Msg, err)
		return *resp
	} else {
		oldpath = video.Path
		_, fname := filepath.Split(oldpath)
		newpath = filepath.Join(videoDeletePath, fname)
		if err := os.Rename(".\\"+oldpath, ".\\"+newpath); err != nil {
			resp.Msg = fmt.Sprint("[VIDEO DELETE] failed:[", video.Id, "]", video.Path, "::", err)
			log.Println(resp.Msg)
			return *resp
		}
	}
	video.Path = strings.Replace(video.Path, "片库", "删除", -1)
	video.State = INVALID
	*resp = video.Update(&video, "state", "path")
	return *resp
}

// 收藏视频
//
//	@param  id [int]
//	@return [*]
func (m Video) Collecting(id int) RespData {
	log.Println("[VIDEO COLLECT]", id)
	resp := NewRespData()
	video := &Video{Id: id}

	if err := Orm.Read(video); err != nil {
		resp.Msg = fmt.Sprint("Not Found Video By ID:", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	video.Collect = video.Collect + 1
	*resp = video.Update(video, "collect")
	if resp.Code == SUCCESS {
		resp.Msg = "收藏视频成功！"
	}
	return *resp
}

// 收藏视频
//
//	@param  id [int]
//	@return [*]
func (m Video) AddTimeNode(id int, time float64) RespData {
	log.Println("[VIDEO TIMENODE] ID[", id, "] ADD TIME=", time)
	resp := NewRespData()
	video := &Video{Id: id}
	if err := Orm.Read(video); err != nil {
		resp.Msg = fmt.Sprint("Not Found Video By ID:", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	timenodes := []float64{}
	if video.Timenode != "" {
		if err := json.Unmarshal([]byte(video.Timenode), &timenodes); err != nil {
			resp.Msg = fmt.Sprint("Unmarshal video timenode failed:", err)
			log.Println(resp.Msg)
			return *resp
		}
	}
	for _, t := range timenodes {
		if int64(t) == int64(time) {
			resp.Msg = "时间结点已存在，添加失败!"
			return *resp
		}
	}
	timenodes = append(timenodes, time)
	bytes, _ := json.Marshal(timenodes)
	video.Timenode = string(bytes)
	*resp = video.Update(video, "timenode")
	if resp.Code == SUCCESS {
		resp.Msg = "添加时间结点成功！"
	}
	return *resp
}

func (m Video) DeleteTimeNode(id int, time float64) RespData {
	log.Println("[VIDEO TIMENODE] ID[", id, "] DEL TIME=", time)
	resp := NewRespData()
	video := &Video{Id: id}
	if err := Orm.Read(video); err != nil {
		resp.Msg = fmt.Sprint("Not Found Video By ID:", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	timenodes := []float64{}
	if video.Timenode != "" {
		if err := json.Unmarshal([]byte(video.Timenode), &timenodes); err != nil {
			resp.Msg = fmt.Sprint("Unmarshal video timenode failed:", err)
			log.Println(resp.Msg)
			return *resp
		}
	}
	timeCpy := []float64{}
	for _, t := range timenodes {
		if int64(t) == int64(time) {
			continue
		} else {
			timeCpy = append(timeCpy, t)
		}
	}
	bytes, _ := json.Marshal(timeCpy)
	video.Timenode = string(bytes)
	*resp = video.Update(video, "timenode")
	if resp.Code == SUCCESS {
		resp.Msg = "删除时间结点成功！"
	}
	return *resp
}

// 给视频添加播放量
//
//	@param  id [int]
//	@return [*]
func (m Video) Play(id int) RespData {
	log.Println("[VIDEO VIEW]", id)
	resp := NewRespData()
	video := &Video{Id: id}

	if err := Orm.Read(video); err != nil {
		resp.Msg = fmt.Sprint("Not Found Video By ID:", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	video.View = video.View + 1
	*resp = video.Update(video, "view")
	return *resp
}

// 爬取封面并下载更新 JAVBUS
//
//	@param  id [int]
//	@param  url [string]
//	@return [*]
func JavbusCoverDownUp(id int, url string) RespData {
	log.Println("[SPIDER IMG] url:", url)
	ctx, cancel := chromedp.NewContext(context.Background())
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var imgNodes []*cdp.Node
	var actorNodes []*cdp.Node

	resp := NewRespData()
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Nodes(`.container .screencap .bigImage`, &imgNodes),
		chromedp.Nodes(`#star-div .avatar-box .photo-frame img`, &actorNodes),
	)
	if err != nil {
		resp.Msg = fmt.Sprintf(" <span class=\"bg-warning\">%s GET FAILED</span> ", url)
		log.Println(resp.Msg, err)
		return *resp
	}
	resp.Code = SUCCESS

	durl := javbus_baseurl + imgNodes[0].AttributeValue("href")
	lurl, derr := mzutils.DownloadUrlByChromedp("pics", durl)
	video := &Video{Id: id}
	actorIds := []string{}
	for _, node := range actorNodes {
		if resp := new(Actor).Add(node.AttributeValue("title")); resp.Code != ERROR {
			actorIds = append(actorIds, fmt.Sprint(resp.Data.(Actor).Id))
		}
	}
	if len(actorIds) > 1 {
		sort.Slice(actorIds, func(i, j int) bool {
			return actorIds[i] < actorIds[j]
		})
	}
	bytes, _ := json.Marshal(actorIds)
	video.Actorid = string(bytes)
	if derr != nil {
		log.Println("[DOWNLOAD FAILED] download cover", durl, "failed:", derr)
	} else {
		video.Cover = lurl
		video.Update(video, "cover", "actorid")
	}
	resp.Data = lurl
	resp.Msg = fmt.Sprintf(" <span class=\"bg-info\">%s</span> ", resp.Data)
	return *resp
}

// 封面下载并更新 JAVPORN
//
//	@param  id [int]
//	@param  url [string] 网页上的图片，以 .ext 结尾
//	@return [RespData]
func WebCoverDownUp(id int, url string) RespData {
	log.Println("[WEB] download cover:", url, " for ID:", id)
	resp := NewRespData()
	lurl, derr := mzutils.DownloadUrlByChromedp("pics", url)
	video := &Video{Id: id}
	if derr != nil {
		resp.Msg = fmt.Sprint("[DOWNLOAD] download cover", url, "failed.")
		log.Println(resp.Msg, derr)
		return *resp
	}
	video.Cover = lurl
	video.Update(video, "cover")
	resp.Code = SUCCESS
	resp.Data = lurl
	resp.Msg = "下载封面链接成功!"
	return *resp
}

// 爬取封面并下载更新 JAVDOE
//
//	@param  id [int]
//	@param  url [string]
//	@return [*]
func JavdoeCoverDownUp(id int, url string) RespData {
	log.Println("[JAVDOE] download cover:", url, " for ID:", id)
	resp := NewRespData()
	col := colly.NewCollector()
	col.OnError(func(r *colly.Response, err error) {
		resp.Msg = fmt.Sprint("download cover[", url, "] failed:", err)
		log.Println(resp.Msg)
	})
	// 定位标签。注册该函数，框架内部回调
	col.OnHTML("#video-player img", func(elem *colly.HTMLElement) {
		coverUrl := elem.Attr("src")
		var c = colly.NewCollector()
		c.OnResponse(func(r *colly.Response) {
			reader := bytes.NewReader(r.Body)
			body, _ := ioutil.ReadAll(reader)
			//读取图片内容
			ext := coverUrl[strings.LastIndex(coverUrl, "."):]
			filepath := filepath.Join("./static/pics/", mzutils.UniqueId()+ext)
			if _, err := os.Create(filepath); err != nil {
				log.Println(err)
			}
			err := ioutil.WriteFile(filepath, body, 0755)
			if err != nil {
				log.Println(coverUrl, err)
			} else {
				video := &Video{Id: id, Cover: "/" + filepath}
				video.Update(video, "cover")
				resp.Code = SUCCESS
				resp.Data = "/" + filepath
				resp.Msg = "下载封面成功！"
			}
		})
		c.Visit(coverUrl)
	})
	col.Visit(url)
	return *resp
}

// 爬取封面并下载更新 JAVPORN
//
//	@param  id [int]
//	@param  url [string]
//	@return [*]
func JavpornCoverDownUp(id int, url string) RespData {
	log.Println("[SPIDER IMG] url:", url)
	ctx, cancel := chromedp.NewContext(context.Background())
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var imgNodes []*cdp.Node
	resp := NewRespData()
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Nodes(`#video-player img`, &imgNodes),
	)
	if err != nil {
		resp.Msg = fmt.Sprintf(" <span class=\"bg-warning\">%s</span> ", url)
		log.Println(resp.Msg, err)
		return *resp
	}
	resp.Code = SUCCESS
	durl := imgNodes[0].AttributeValue("src")
	lurl, derr := mzutils.DownloadUrlByChromedp("pics", durl)
	video := &Video{Id: id}
	if derr != nil {
		log.Println("[DOWNLOAD FAILED] download cover", durl, "failed:", derr)
	} else {
		video.Cover = lurl
		video.Update(video, "cover")
	}
	resp.Data = lurl
	resp.Msg = fmt.Sprintf(" <span class=\"bg-info\">%s</span> ", resp.Data)
	return *resp
}

// 添加本地视频文件
func addLocalVideo() {
	upvPreffix := "[VIDEO ADD LOCAL]"
	files, err := ioutil.ReadDir(videoUpdatePath)
	firstLog := true
	if err != nil {
		log.Println("get directory failed:", err)
	}
	for _, fi := range files {
		if !fi.IsDir() {
			ext := strings.ToLower(filepath.Ext(fi.Name()))[1:]
			if ext == "mp4" {
				if firstLog {
					log.Println(upvPreffix, "add new video from path:", videoUpdatePath)
					firstLog = false
				}
				editAddKLocalVideo(fi, "")
			} else {
				log.Println(upvPreffix, "not mp4 file:", filepath.Join(videoUpdatePath, fi.Name()))
			}
		} else {
			// log.Println("read dir", fi.Name())
			subdir, err := ioutil.ReadDir(filepath.Join(videoUpdatePath, fi.Name()))
			if err != nil {
				log.Println("get directory failed:", err)
				continue
			}
			dirname := fi.Name()
			for _, vfile := range subdir {
				ext := strings.ToLower(filepath.Ext(vfile.Name()))[1:]
				if ext == "mp4" {
					editAddKLocalVideo(vfile, dirname)
				}
			}
		}
	}
}

func editAddKLocalVideo(vfile fs.FileInfo, own string) {
	upvPreffix := "[VIDEO ADD LOCAL]"
	fname := vfile.Name()
	vname := fname[:strings.LastIndex(fname, ".")]
	log.Printf("%s %s [%dMB]\n", upvPreffix, fname, vfile.Size()/1024/1024)
	storepath := filepath.Join(videoStorePath, fname)
	if own != "" {
		for i := 1; ; i++ {
			ext := strings.ToLower(filepath.Ext(vfile.Name()))[1:]
			vname = fmt.Sprintf("%s-%d.%s", own, i, ext)
			storepath = filepath.Join(videoStorePath, vname)
			if res, _ := mzutils.PathExists(".\\" + storepath); !res {
				break
			}
		}
	} else if res, _ := mzutils.PathExists(".\\" + storepath); res {
		log.Printf("PATH[%s] is existed. add video[%s] failed", storepath, fname)
		return
	}
	// move video file to storepath
	if err := os.Rename(".\\"+filepath.Join(videoUpdatePath, own, fname), ".\\"+storepath); err != nil {
		log.Println(upvPreffix, "[FILE MOVE] move file", fname, "failed ::", err)
		return
	}
	// get video duration
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf(`/c %s`, fmt.Sprintf("ffmpeg -i \"./%s\"", storepath)), HideWindow: true}
	out, _ := cmd.CombinedOutput()
	arr := strings.Split(string(out), "\n")
	if len(arr) < 4 {
		log.Println("get video duration failed.")
		log.Println(string(out))
	}
	duration := ""
	for _, s := range arr {
		if strings.HasPrefix(s, "  Duration") {
			duration = s[12:20]
			break
		}
	}
	log.Printf("[VIDEO DURATION] PATH[%s] DURATION[%s]", storepath, duration)
	video := Video{
		Title:    vname,
		Path:     "\\" + storepath,
		Pubtime:  time.Now().Format("2006-01-02 15:04:05"),
		Duration: duration,
	}
	video.Add(video)
}

func (this Video) SetDuration() {
	log.Println("[SET VIDEO DURATION]")
	videos := []Video{}
	Orm.QueryTable(this.TableName()).Filter("state", VALID).Filter("duration", "").All(&videos)
	for _, v := range videos {
		cmd := exec.Command("cmd")
		cmdstr := fmt.Sprintf("ffmpeg -i \"./%s\"", v.Path)
		cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf(`/c %s`, cmdstr), HideWindow: true}
		out, _ := cmd.CombinedOutput()
		arr := strings.Split(string(out), "\n")
		for _, s := range arr {
			if strings.HasPrefix(s, "  Duration") {
				v.Duration = s[12:20]
				break
			}
		}
		if v.Duration == "" {
			log.Printf("GET VIDEO[%d][%s] duration failed:\n%v", v.Id, v.Title, string(out))
			continue
		}
		log.Printf("[%s] %s", v.Title, v.Duration)
		v.Update(&v, "duration")
	}
}

type MessageChan struct {
	Code int    `json:"code"`
	Time string `json:"time"`
	Msg  string `json:"msg"`
}

var MsgChan = make(chan MessageChan)

func (this Video) SpiderCover() {
	log.Println("[SPIDER VIDEOS' COVER LOOP]")
	videos := []Video{}
	if _, err := Orm.QueryTable(this.TableName()).Filter("state", VALID).Filter("cover", "").All(&videos); err != nil {
		log.Println("read videos failed:", err)
		return
	}
	for _, v := range videos {
		if !strings.HasPrefix(v.Title, "FC2") {
			continue
		}
		urlTitle := v.Title
		urlTitle = strings.ToUpper(urlTitle)
		urlTitle = strings.Replace(urlTitle, "FC2 PPV", "FC2-PPV", -1)
		urlTitle = strings.Replace(urlTitle, "PPV ", "PPV-", -1)
		resp := collyGetCover(javdock_baseurl + "/" + urlTitle + "/")
		MsgChan <- MessageChan{
			Code: resp.Code,
			Time: mzutils.NowTimeString(),
			Msg:  fmt.Sprintf("ID[%03d] TITLE[%s] %s", v.Id, v.Title, resp.Msg),
		}
		if resp.Code != SUCCESS {
			continue
		} else {
			v.Cover = resp.Data.(string)
			this.Update(&v, "cover")
		}
	}
	MsgChan <- MessageChan{
		Code: ENDING,
		Time: mzutils.NowTimeString(),
		Msg:  fmt.Sprint("SPIDER END"),
	}
}

func collyGetCover(url string) RespData {
	log.Println("[SPIDER IMG] url:", url)
	resp := NewRespData()
	col := colly.NewCollector()
	col.OnError(func(r *colly.Response, err error) {
		resp.Msg = fmt.Sprint("get url[", url, "] failed:", err)
		log.Println(resp.Msg)
		// wg.Done()
	})
	// 定位标签。注册该函数，框架内部回调
	col.OnHTML("#video-player img", func(elem *colly.HTMLElement) {
		coverUrl := elem.Attr("src")
		var c = colly.NewCollector()
		c.OnResponse(func(r *colly.Response) {
			reader := bytes.NewReader(r.Body)
			body, _ := ioutil.ReadAll(reader)
			//读取图片内容
			ext := coverUrl[strings.LastIndex(coverUrl, "."):]
			filepath := filepath.Join("./static/pics/", mzutils.UniqueId()+ext)
			if _, err := os.Create(filepath); err != nil {
				log.Println(err)
			}
			err := ioutil.WriteFile(filepath, body, 0755)
			if err != nil {
				log.Println(coverUrl, err)
			} else {
				resp.Code = SUCCESS
				resp.Data = "/" + filepath
				resp.Msg = "获取图片成功！"
			}
		})
		c.Visit(coverUrl)
	})
	col.Visit(url)
	return *resp
}
