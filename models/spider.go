package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"myzone/mzutils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/liuzl/gocc"
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
	Code     int         `json:"code"`
	Msgmod   int         `json:"-"`
	Type     int         `json:"type"`
	Msg      string      `json:"msg"`
	Data     interface{} `json:"data"`
	DataType string      `json:"dataType"`
}

var (
	SpiderMsgChan     = make(chan SpiderMsg)
	chapterTryTimes   = 3
	SPIDER_THREAD_NUM = 10
	wg                sync.WaitGroup
)

const (
	// msg type 信息类别
	SPIDER_INFO int = 0
	SPIDER_DATA int = 1
	SPIDER_DONE int = 2

	// msg module
	SPIDER_MODULE_DEFAULT int = 0
	SPIDER_MODULE_SHT     int = 1
	SPIDER_MODULE_UAA     int = 2
	SPIDER_MODULE_2048    int = 3

	SECTION_DEFAULT         int = 0
	SHT_SECTION_DOMESTIC    int = 2
	SHT_SECTION_ASIA_NO     int = 36
	SHT_SECTION_ASIA_MOSAIC int = 37

	TYPE_DEFAILT       int = 0
	SHT_TYPE_D_NOMASIC int = 684
	SHT_TYPE_D_ANCHOR  int = 685
	SHT_TYPE_A_FC2     int = 368
	SHT_TYPE_A_MCRACK  int = 672
	SHT_TYPE_A_LEAKED  int = 654

	TYPE_2048_A_NOMASIC int = 4
	TYPE_2048_DOMESTIC  int = 15

	SHT_DETELINE_BASE int = 86400

	BASE_URL_SHT    string = "https://www.sehuatang.org/"
	BASE_URL_2048   string = `https://hjd2048.com/2048/`
	bookStorePath          = `./static/novel/`
	bookUaaJsonPath        = `./static/json/book.uaa.json`
	UAA_BASE_URL           = `https://api.uaa.com/novel/app/novel/`

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
	if typeid == TYPE_DEFAILT && day != 0 {
		if day != 0 {
			return fmt.Sprintf("%sforum.php?mod=forumdisplay&fid=%d&filter=dateline&dateline=%d", BASE_URL_SHT, section, day*SHT_DETELINE_BASE)
		}
		return fmt.Sprintf("%sforum-%d-1.html", BASE_URL_SHT, section)
	}
	return fmt.Sprintf("%sforum.php?mod=forumdisplay&fid=%d&filter=typeid&typeid=%d&filter=dateline&dateline=%d", BASE_URL_SHT, section, typeid, day*SHT_DETELINE_BASE)
}

// setcookies returns a task to navigate to a host with the passed cookies set
// on the network request.
func setcookies(host string, cookies ...string) chromedp.Tasks {
	if len(cookies)%2 != 0 {
		panic("length of cookies must be divisible by 2")
	}
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// create cookie expiration
			expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			// add cookies to chrome
			for i := 0; i < len(cookies); i += 2 {
				err := network.SetCookie(cookies[i], cookies[i+1]).
					WithExpires(&expr).
					WithDomain("https://www.sehuatang.org/").
					WithHTTPOnly(true).
					Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
	}
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
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies := []string{
				"_safe", "vqd37pjm4p5uodq339yzk6b7jdt6oich",
				"cPNj_2132_atarget", "1",
				"cPNj_2132_saltkey", "QLnum82q",
				"cPNj_2132_st_t", "0|1679119862|11fb5a8e69985aa896f0c6b7b2fdb616",
			}
			// create cookie expiration
			expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			// add cookies to chrome
			for i := 0; i < len(cookies); i += 2 {
				err := network.SetCookie(cookies[i], cookies[i+1]).
					WithExpires(&expr).
					WithDomain("www.sehuatang.org").
					WithHTTPOnly(true).
					Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
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
			chromedp.ActionFunc(func(ctx context.Context) error {
				cookies := []string{
					"_safe", "vqd37pjm4p5uodq339yzk6b7jdt6oich",
					"cPNj_2132_atarget", "1",
					"cPNj_2132_saltkey", "QLnum82q",
					"cPNj_2132_st_t", "0|1679119862|11fb5a8e69985aa896f0c6b7b2fdb616",
				}
				// create cookie expiration
				expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
				// add cookies to chrome
				for i := 0; i < len(cookies); i += 2 {
					err := network.SetCookie(cookies[i], cookies[i+1]).
						WithExpires(&expr).
						WithDomain("www.sehuatang.org").
						WithHTTPOnly(true).
						Do(ctx)
					if err != nil {
						return err
					}
				}
				return nil
			}),
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
		t2s, _ := gocc.New("t2s")
		for i := range infoNodes {
			if len(infoNodes[i].Children) == 0 || len(numNodes[i].Children) == 0 {
				continue
			}
			title := infoNodes[i].Children[0].NodeValue
			out, err := t2s.Convert(title)
			if err == nil {
				title = out
			}
			spider := Spider{
				Title:    title,
				Url:      BASE_URL_SHT + infoNodes[i].AttributeValue("href"),
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
			_, err = Orm.Insert(&spider)
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

func set2048Url(fid, day int) string {
	if day != 0 {
		return fmt.Sprintf("%sthread.php?fid=%d&search=%d", BASE_URL_2048, fid, day)
	}
	return fmt.Sprintf("%sthread.php?fid=%d", BASE_URL_SHT, fid)
}

func get2048PageNum(url string) int {
	log.Println("[2048] Get page number [", url, "]")
	SpiderMsgChan <- SpiderMsg{
		Code: SUCCESS,
		Type: SPIDER_INFO,
		Msg:  "Get page number [ " + url + " ]",
	}
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	numNodes := []*cdp.Node{}
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Nodes(`#main .pagesone span`, &numNodes),
	); err != nil {
		log.Println("Get page number failed:", err)
		SpiderMsgChan <- SpiderMsg{
			Code: ERROR,
			Type: SPIDER_INFO,
			Msg:  fmt.Sprint("Get page number failed:", err),
		}
		return 0
	}
	//'Pages: 2/3014'
	var num int = 0
	str := numNodes[0].Children[0].NodeValue
	num, err := strconv.Atoi(str[strings.LastIndex(str, "/")+1:])
	if err != nil {
		log.Println("CONVERT page number(", str, ") failed:", err)
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

func (this Spider) Spider2048(fid, day int) {
	url := set2048Url(fid, day)
	pagenum := get2048PageNum(url)
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
		timeNodes := []*cdp.Node{}

		if err := chromedp.Run(ctx,
			chromedp.Navigate(pageurl),
			chromedp.Nodes(`#main .tr3.t_one a.subject`, &infoNodes),
			chromedp.Nodes(`#main .tr3.t_one .tal.y-style .f10.gray`, &timeNodes),
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
		t2s, _ := gocc.New("t2s")
		for i := range infoNodes {
			if len(infoNodes[i].Children) == 0 || len(timeNodes[i].Children) == 0 {
				continue
			}
			title := infoNodes[i].Children[0].NodeValue
			out, err := t2s.Convert(title)
			if err == nil {
				title = out
			}
			// 亚洲无码只爬取FC2
			if fid == TYPE_2048_A_NOMASIC && strings.Index(title, "FC2-PPV") == -1 {
				continue
			}
			spider := Spider{
				Title:    title,
				Url:      BASE_URL_2048 + infoNodes[i].AttributeValue("href"),
				Addtime:  mzutils.NowTimeString(),
				Pubtime:  timeNodes[i*2+1].Children[0].NodeValue,
				Module:   SPIDER_MODULE_2048,
				Category: fid,
				Section:  SECTION_DEFAULT,
			}
			tempSpider := Spider{Url: spider.Url}
			if serr := Orm.Read(&tempSpider, "url"); serr == nil {
				log.Println("Have the item", tempSpider.Title)
				continue
			}

			if _, err = Orm.Insert(&spider); err != nil {
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
	cond = cond.And("module", mod)
	seter := Orm.QueryTable(this.TableName())
	if section != SECTION_DEFAULT {
		cond = cond.And("section", section)
		// seter = seter.Filter("section", section)
	}
	if category != TYPE_DEFAILT {
		cond = cond.And("category", category)
		// seter = seter.Filter("category", category)
	}
	if len(filter) != 0 && filter[0] != "" {
		cond = cond.And("title__icontains", filter[0])
		// fstrs := strings.Split(filter[0], " ")
		// for _, s := range fstrs {
		// 	cond = cond.Or("title__icontains", s)
		// }
	}
	seter = seter.SetCond(cond)
	seter = seter.Distinct()
	seter = seter.OrderBy("-addtime", "-id")
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

type Book struct {
	Id           string    `json:"id" orm:"pk"`
	Author       string    `json:"authors"`
	Title        string    `json:"title"`
	Chaptercount int       `json:"chapterCount"`
	Brief        string    `json:"brief"`
	Chapters     []Chapter `json:"chapters" orm:"-"`
	Spidertime   string    `json:"spiderTime"`
}

func (Book) TableName() string {
	return "book"
}

type Chapter struct {
	Id       string    `json:"id"`
	Pid      int       `json:"pid"`
	Title    string    `json:"title"`
	Children []Chapter `json:"children" orm:"-"`
	Content  string    `json:"content"`
}

type UaaBookResp struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Model struct {
		ChapterCount int       `json:"chapterCount"`
		Menus        []Chapter `json:"menus"`
	} `json:"model"`
}

type UaaChapterResp struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Model struct {
		Lines []string `json:"lines"`
	} `json:"model"`
}

type UaaBookIntroResp struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Model struct {
		Book
	} `json:"model"`
}

func (this Book) Add(book *Book) RespData {
	log.Printf("[ADD BOOK] UD[%s] TITLE[%s]", book.Id, book.Title)
	resp := NewRespData()
	if _, err := Orm.Insert(book); err != nil {
		resp.Msg = fmt.Sprintf("ADD BOOK[%s] FAILED: %v", book.Title, err)
		log.Println(resp.Msg)
		return *resp
	}
	resp.Code = SUCCESS
	resp.Msg = fmt.Sprint("add book", book.Title, "success!")
	return *resp
}

func (this Book) GetBookList() RespData {
	log.Println("[BOOK LIST] get book list")
	resp := NewRespData()
	booklist := []Book{}
	if _, err := Orm.QueryTable(this.TableName()).OrderBy("-spidertime").All(&booklist); err != nil {
		resp.Msg = fmt.Sprint("get book list failed:", err)
		log.Println(resp.Msg)
		return *resp
	}
	resp.Code = SUCCESS
	resp.Msg = "get book list success"
	resp.Data = booklist
	return *resp
}

func (this Spider) SpiderUaa(nidarr []string) {
	books := []Book{}
	bytes, err := ioutil.ReadFile(bookUaaJsonPath)
	if err != nil {
		json.Unmarshal(bytes, &books)
	}
	for _, id := range nidarr {
		bk := &Book{Id: id}
		if err := Orm.Read(bk); err == nil {
			log.Printf("BOOK ID[%s] TITLE[%s] is existed", bk.Id, bk.Title)
			SpiderMsgChan <- SpiderMsg{
				Code: ERROR,
				Type: SPIDER_INFO,
				Msg:  fmt.Sprintf("BOOK ID[%s] TITLE[%s] is existed", bk.Id, bk.Title),
			}
			continue
		}
		wg = sync.WaitGroup{}
		book, err := spiderUaa(id)
		if err == nil {
			bk.Add(&book)
		}
		books = append(books, book)
	}
	SpiderMsgChan <- SpiderMsg{
		Code: SUCCESS,
		Type: SPIDER_DONE,
	}
	bytes, _ = json.Marshal(books)
	jsonfile, err := os.OpenFile(bookUaaJsonPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	jsonfile.Write(bytes)
	defer jsonfile.Close()
}

func spiderUaa(id string) (Book, error) {
	book, err := getBook(fmt.Sprintf("%scatalog/%s", UAA_BASE_URL, id))
	if err != nil {
		return book, err
	}
	book.Id = id
	if err = setBookIntroduction(&book, fmt.Sprintf("%sintro?id=%s", UAA_BASE_URL, id)); err != nil {
		return book, err
	}
	SpiderMsgChan <- SpiderMsg{
		Code: INFO,
		Type: SPIDER_DATA,
		Data: book,
	}
	SpiderMsgChan <- SpiderMsg{
		Code: INFO,
		Type: SPIDER_INFO,
		Msg:  fmt.Sprintf("start to get book[%s] chapter count[%d]", book.Title, book.Chaptercount),
	}
	log.Println("=============== GET BOOK", book.Title, "CHAPTER COUNT:", book.Chaptercount)
	filename := fmt.Sprintf("%s%s.txt", bookStorePath, book.Title)
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Printf("open file[%s] failed: %v", filename, err)
		return book, err
	}
	defer file.Close()
	file.WriteString(fmt.Sprint(book.Title, "\n作者：", book.Author, "\n简介：\n", book.Brief, "\n\n"))
	for i, ch := range book.Chapters {
		if len(ch.Children) != 0 {
			len := len(ch.Children)
			for j := 0; j < len; {
				for idx := 0; idx < SPIDER_THREAD_NUM && j+idx < len; idx++ {
					wg.Add(1)
					go getUaaChapter(fmt.Sprintf("%schapter?offset=0&viewId=817988423091621957&id=%s", UAA_BASE_URL, ch.Children[j+idx].Id), ch.Children[j+idx].Title, &(book.Chapters[i].Children[j+idx].Content))
				}
				wg.Wait()
				j += SPIDER_THREAD_NUM
			}
		} else {
			wg.Add(1)
			go getUaaChapter(fmt.Sprintf("%schapter?offset=0&viewId=817988423091621957&id=%s", UAA_BASE_URL, ch.Id), ch.Title, &(book.Chapters[i].Content))
			wg.Wait()
		}
	}
	SpiderMsgChan <- SpiderMsg{
		Code: INFO,
		Type: SPIDER_INFO,
		Msg:  fmt.Sprint("正在组装txt文件"),
	}
	for _, ch := range book.Chapters {
		file.WriteString(ch.Title + "\n\n")
		file.WriteString(ch.Content + "\n\n")
		if len(ch.Children) != 0 {
			for _, c := range ch.Children {
				file.WriteString(c.Title + "\n\n")
				file.WriteString(c.Content + "\n\n")
			}
		}
	}
	SpiderMsgChan <- SpiderMsg{
		Code: SUCCESS,
		Type: SPIDER_INFO,
		Msg:  fmt.Sprintf("组装[%s]txt文件完成", book.Title),
	}
	return book, nil
}

func getBook(url string) (Book, error) {
	book := Book{}
	SpiderMsgChan <- SpiderMsg{
		Code: INFO,
		Type: SPIDER_INFO,
		Msg:  fmt.Sprint("get book info:", url),
	}
	log.Println("[SPIDER UAA] get book info:", url)
	req, err := http.Get(url)
	if err != nil {
		log.Println("get url failed:", err)
		SpiderMsgChan <- SpiderMsg{
			Code: ERROR,
			Type: SPIDER_INFO,
			Msg:  fmt.Sprint("get url failed:", err),
		}
		return book, err
	}
	req.Header.Set("User-Agent", "spacecount-tutorial")
	defer req.Body.Close()
	if req.StatusCode != 200 {
		err = fmt.Errorf("status code is not 200:%s", req.Status)
		SpiderMsgChan <- SpiderMsg{
			Code: ERROR,
			Type: SPIDER_INFO,
			Msg:  fmt.Sprint(err),
		}
		log.Println(err)
		return book, err
	}
	body, _ := ioutil.ReadAll(req.Body)
	resp := UaaBookResp{}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Println("unmarshal uaa response failed:", err)
		return book, err
	}
	book = Book{
		Chaptercount: resp.Model.ChapterCount,
		Chapters:     resp.Model.Menus,
	}
	return book, nil
}

func getUaaChapter(url, title string, chContent *string) {
	log.Println("[SPIDER UAA] chapter url:", url, " ", title)
	SpiderMsgChan <- SpiderMsg{
		Code: INFO,
		Type: SPIDER_INFO,
		Msg:  fmt.Sprintf("get chapter[%s] with <a href=\"%s\">URL</a>", title, url),
	}
	var req *http.Response
	var err error
	var i int
	for i = 1; i <= chapterTryTimes; i++ {
		req, err = http.Get(url)
		if err != nil {
			log.Println("get url failed:", err)
		}
		req.Header.Set("User-Agent", "spacecount-tutorial")
		defer req.Body.Close()
		if req.StatusCode != 200 {
			log.Println("status code is not 200:", req.Status)
		}
		body, _ := ioutil.ReadAll(req.Body)
		resp := UaaChapterResp{}
		if err := json.Unmarshal(body, &resp); err != nil {
			log.Println("unmarshal uaa response failed:", err)
			continue
		}
		content := ""
		content = strings.Join(resp.Model.Lines, "\n")
		if err == nil && req.StatusCode == 200 && content != "" {
			t2s, _ := gocc.New("t2s")
			out, _ := t2s.Convert(content)
			(*chContent) = out
			break
		}
	}
	if i > chapterTryTimes {
		log.Printf("GET CHAPTER[%s] FAILED", title)
		SpiderMsgChan <- SpiderMsg{
			Code: ERROR,
			Type: SPIDER_INFO,
			Msg:  fmt.Sprintf("GET CHAPTER[%s] FAILED", title),
		}
		wg.Done()
		return
	}

	SpiderMsgChan <- SpiderMsg{
		Code: SUCCESS,
		Type: SPIDER_DATA,
		Data: Chapter{
			Title:   title,
			Content: fmt.Sprint(len((*chContent))),
		},
	}
	wg.Done()
}

func setBookIntroduction(book *Book, url string) error {
	log.Println("[SPIDER UAA] get book introduction:", url)
	req, err := http.Get(url)
	if err != nil {
		log.Println("get url failed:", err)
		return err
	}
	req.Header.Set("User-Agent", "spacecount-tutorial")
	defer req.Body.Close()
	if req.StatusCode != 200 {
		log.Println("status code is not 200:", req.Status)
		return err
	}
	body, _ := ioutil.ReadAll(req.Body)
	resp := UaaBookIntroResp{}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Println("unmarshal uaa response failed:", err)
		return err
	}
	(*book).Title = resp.Model.Title
	(*book).Brief = resp.Model.Brief
	(*book).Author = resp.Model.Author
	(*book).Spidertime = mzutils.NowTimeString()
	return nil
}

func (this Spider) SpiderCharsetRegulate() {
	spiders := []Spider{}
	Orm.QueryTable(this.TableName()).All(&spiders)
	for _, s := range spiders {
		t2s, _ := gocc.New("t2s")
		out, err := t2s.Convert(s.Title)
		if out != s.Title && err == nil {
			log.Println("CHANGE:", s.Title, "-->", out)
			s.Title = out
			Orm.Update(&s, "title")
		}
	}
}
