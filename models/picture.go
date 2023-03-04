package models

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type Screenshot struct {
	Title string `json:"title"`
	Path  string `json:"path"`
	Time  string `json:"time"`
}

const (
	screenshotPath string = `./static/screenshot/`
)

func GetScreenshotList() RespData {
	resp := NewRespData()
	resp.Data = ScreenshootList
	resp.Code = SUCCESS
	return *resp
}

func SetScreenshotList(list *[]Screenshot) {
	files, err := ioutil.ReadDir(screenshotPath)
	if err != nil {
		log.Println("read screenshoot dir failed:", err)
		return
	}
	for _, f := range files {
		ftitle := f.Name()
		time := strings.Split(ftitle, "]")[2]
		*list = append(*list, Screenshot{
			Title: ftitle[:strings.LastIndex(ftitle, ".")],
			Path:  "\\" + filepath.Join(screenshotPath, ftitle),
			Time:  time[:strings.LastIndex(time, ".")],
		})
	}
}
