package models

import (
	"context"
	"fmt"
	"log"
	"myzone/mzutils"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type SpiderConfig struct {
}

type Spider struct {
	Id       int         `json:"id"`
	Title    string      `json:"title"`
	Url      string      `json:"url"`
	Pubtime  string      `json:"pubtime"`
	Datastr  string      `json:"-"`
	Data     interface{} `json:"data" orm:"-"`
	Module   int         `json:"module"`
	Section  int         `json:"section"`
	Category int         `json:"category"`
	View     int         `json:"view"`
	Addtime  string      `json:"addtime"`
	Preid    int         `json:"preid"`
}

type SpiderMsg struct {
	Code int         `json:"code"`
	Type int         `json:"type"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var (
	SpiderMsgChan = make(chan SpiderMsg)
)

const (
	// msg type 信息类别
	SPIDER_INFO int = 0
	SPIDER_DATA int = 1
	SPIDER_DONE int = 2

	SPIDER_MODULE_DEFAULT int = 0
	SPIDER_MODULE_SHT     int = 1

	SHT_SECTION_DEFAULT     int = 0
	SHT_SECTION_DOMESTIC    int = 2
	SHT_SECTION_ASIA_NO     int = 36
	SHT_SECTION_ASIA_MOSAIC int = 37

	SHT_TYPE_DEFAILT   int = 0
	SHT_TYPE_D_NOMASIC int = 684
	SHT_TYPE_D_ANCHOR  int = 685
	SHT_TYPE_A_FC2     int = 368
	SHT_TYPE_A_MCRACK  int = 672
	SHT_TYPE_A_LEAKED  int = 654

	SHT_DETELINE_BASE int = 86400

	SHT_BASE_URL string = "https://www.sehuatang.org/"

	SPIDER_TABLE_LIMIT int = 10
)

func (Spider) TableName() string {
	return "spider"
}

func (this Spider) GetCount() int {
	count, _ := Orm.QueryTable(this.TableName()).Count()
	return int(count)
}

func (this Spider) GetLastAddtime() string {
	addtime := ""
	Orm.Raw(`select addtime from spider order by addtime desc limit 1`).QueryRow(&addtime)
	return addtime
}

func setShtUrl(section, typeid, day int) string {
	if typeid == SHT_TYPE_DEFAILT && day != 0 {
		if day != 0 {
			return fmt.Sprintf("%sforum.php?mod=forumdisplay&fid=%d&filter=dateline&dateline=%d", SHT_BASE_URL, section, day*SHT_DETELINE_BASE)
		}
		return fmt.Sprintf("%sforum-%d-1.html", SHT_BASE_URL, section)
	}
	return fmt.Sprintf("%sforum.php?mod=forumdisplay&fid=%d&filter=typeid&typeid=%d&filter=dateline&dateline=%d", SHT_BASE_URL, section, typeid, day*SHT_DETELINE_BASE)
}

func getPageNum(url string) int {
	log.Println("[SHT] Get page number [", url, "]")
	SpiderMsgChan <- SpiderMsg{
		Code: SUCCESS,
		Type: SPIDER_INFO,
		Msg:  "Get page number [ " + url + " ]",
	}
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var numstr string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.TextContent(`#fd_page_top label span`, &numstr),
	); err != nil {
		log.Println("Get page number failed:", err)
		SpiderMsgChan <- SpiderMsg{
			Code: ERROR,
			Type: SPIDER_INFO,
			Msg:  fmt.Sprint("Get page number failed:", err),
		}
		return 0
	}
	//' / 3 页'
	var num int = 0
	num, err := strconv.Atoi(numstr[len(" / ") : len(numstr)-4])
	if err != nil {
		log.Println("CONVERT page number(", numstr, ") failed:", err)
		return 0
	}
	log.Println("[SHT] PAGE NUMBER:", num)
	SpiderMsgChan <- SpiderMsg{
		Code: INFO,
		Type: SPIDER_INFO,
		Msg:  fmt.Sprint("page number <data>", num, "</data>"),
	}
	return num
}

func (this Spider) SpiderSht(section, typeid, day int) {
	url := setShtUrl(section, typeid, day)
	pagenum := getPageNum(url)
	if pagenum == 0 {
		return
	}
	for page := pagenum; page > 0; page-- {
		pageurl := fmt.Sprintf("%s&page=%d", url, page)
		log.Println("[PAGE URL]", pageurl)
		SpiderMsgChan <- SpiderMsg{
			Code: SUCCESS,
			Type: SPIDER_INFO,
			Msg:  fmt.Sprintf("Start to spider page <data>%d</data> with <a href=\"%s\" target=\"_blank\">URL</a>", page, pageurl),
		}
		ctx, cancel := chromedp.NewContext(context.Background())
		defer cancel()
		infoNodes := []*cdp.Node{}
		numNodes := []*cdp.Node{}

		if err := chromedp.Run(ctx,
			chromedp.Navigate(pageurl),
			chromedp.Nodes(`#threadlisttableid .s.xst`, &infoNodes),
			chromedp.Nodes(`#threadlisttableid .num em`, &numNodes),
		); err != nil {
			log.Printf("spider page <data>%d</data> with <a href=\"%s\" target=\"_blank\">URL</a> failed: %v\n", page, pageurl, err)
			SpiderMsgChan <- SpiderMsg{
				Code: ERROR,
				Type: SPIDER_INFO,
				Msg:  fmt.Sprintf("spider page <data>%d</data> with <a href=\"%s\" target=\"_blank\">URL</a> failed: %v", page, pageurl, err),
			}
			continue
		}
		infonum := 0
		for i := range infoNodes {
			if len(infoNodes[i].Children) == 0 || len(numNodes[i].Children) == 0 {
				continue
			}
			spider := Spider{
				Title:    infoNodes[i].Children[0].NodeValue,
				Url:      SHT_BASE_URL + infoNodes[i].AttributeValue("href"),
				View:     mzutils.Atoi(numNodes[i].Children[0].NodeValue),
				Addtime:  mzutils.NowTimeString(),
				Module:   SPIDER_MODULE_SHT,
				Category: typeid,
				Section:  section,
			}
			spider.Url = spider.Url[:strings.Index(spider.Url, "&extra")]
			tempSpider := Spider{Url: spider.Url}
			if serr := Orm.Read(&tempSpider, "url"); serr == nil {
				log.Println("Have the item", tempSpider.Title)
				continue
			}
			_, err := Orm.Insert(&spider)
			if err != nil {
				log.Println("ADD SPIDER FAILED:", err)
				SpiderMsgChan <- SpiderMsg{
					Code: ERROR,
					Type: SPIDER_INFO,
					Msg:  fmt.Sprint("ADD SPIDER FAILED:", err),
				}
				continue
			}
			infonum++
			SpiderMsgChan <- SpiderMsg{
				Code: SUCCESS,
				Type: SPIDER_DATA,
				Data: spider,
			}
		}
		SpiderMsgChan <- SpiderMsg{
			Code: INFO,
			Type: SPIDER_INFO,
			Msg:  fmt.Sprintf("Spider page <data>%d</data> with <a href=\"%s\" target=\"_blank\">URL</a> end. Get information piece number: <data>%d</data>", page, pageurl, infonum),
		}
	}
	SpiderMsgChan <- SpiderMsg{
		Code: INFO,
		Type: SPIDER_DONE,
		Msg:  fmt.Sprintf("SPIDER END..."),
	}
}

func (this Spider) GetSpiderList(page, limit, mod, section, category int, filter ...string) RespData {
	resp := NewRespData()
	log.Println("[SPIDER LIST] module[", mod, "] section[", section, "] category[", category, "] page[", page, "] limit[", limit, "]")
	spiders := []Spider{}
	// if len(filter) != 0 && filter[0] != "" {
	// 	fstrs := strings.Split(filter[0], " ")
	// 	fstr := ""
	// 	for _, s := range fstrs {
	// 		fstr += s + "|"
	// 	}
	// 	fstr = fstr[:len(fstr)-1]
	// 	seter := Orm.Raw("SELECT * FROM `spider` WHERE title REGEXP ? limit ? offset ?", fstr, limit, limit*(page-1))
	// 	count, err := seter.QueryRows(&spiders)
	// 	if err != nil {
	// 		resp.Msg = fmt.Sprint("filter spider failed:", err)
	// 		log.Println(resp.Msg)
	// 		return *resp
	// 	}
	// 	resp.Count = int(count)
	// 	resp.Data = spiders
	// 	resp.Code = SUCCESS
	// 	return *resp
	// }
	cond := orm.NewCondition()
	seter := Orm.QueryTable(this.TableName())
	if mod == SPIDER_MODULE_SHT {
		if section != SHT_SECTION_DEFAULT {
			cond = cond.And("section", section)
			// seter = seter.Filter("section", section)
		}
		if category != SHT_TYPE_DEFAILT {
			cond = cond.And("category", category)
			// seter = seter.Filter("category", category)
		}
	}
	if len(filter) != 0 && filter[0] != "" {
		// fstr := strings.ReplaceAll(filter[0], " ", "%")
		// seter = seter.Filter("title", fstr)
		fstrs := strings.Split(filter[0], " ")
		cond = cond.And("title__icontains", filter[0])
		// fstr := ""
		for _, s := range fstrs {
			// fstr += s + "|"
			cond = cond.Or("title__icontains", s)
		}
		// cond = cond.Raw("REGEXP ?", fstr)
		
	}
	seter = seter.SetCond(cond)
	seter = seter.Distinct()
	seter = seter.OrderBy("-id")
	count, _ := seter.Count()
	seter = seter.Limit(limit, (page-1)*limit)
	if _, err := seter.All(&spiders); err != nil {
		resp.Error = err
		resp.Msg = fmt.Sprint("[ERROR]", err)
		log.Println(resp.Msg)
		return *resp
	}

	resp.Code = SUCCESS
	resp.Count = int(count)
	resp.Data = spiders
	return *resp
}
