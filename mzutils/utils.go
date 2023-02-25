package mzutils

import (
	"bufio"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const (
	baseStorePath string = ".\\static\\"
)

// 将字符串s转换成为int，出错则为0
//
//	@param  s [string]
//	@return [int]
func Atoi(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return res
}

// 生成唯一ID
func UniqueId() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return Md5(base64.URLEncoding.EncodeToString(b))
}

func Md5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func NowDate() string {
	return time.Now().Format("2006-01-02")
}

func NowTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func NowTimeStringLong() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func RandomInt(limit int) int {
	res, _ := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	return int(res.Int64())
}

func DownloadFile(fileBelong string, file *multipart.File, handeler *multipart.FileHeader) (string, error) {
	var rerr error = nil
	fext := filepath.Ext(handeler.Filename)
	storepath := filepath.Join(baseStorePath, fileBelong, UniqueId()+fext)
	storefile, err := os.OpenFile(storepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		rerr = fmt.Errorf("创建本地文件失败: %s", storepath)
		log.Println(rerr, ", err:", err)
		return "", rerr
	}
	reader := bufio.NewReader(*file)
	writer := bufio.NewWriter(storefile)
	buffer := make([]byte, 256)
	for {
		_, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
		} else {
			writer.Write(buffer)
		}
	}
	writer.Flush()
	path := ".\\" + storepath
	storefile.Close()
	log.Println("download file success:", storepath)
	return path, rerr
}

// 使用Chromedp下载链接内容
//
//	@param  fileBelong [string] 文件所属类型
//	@param  url [string]
//	@return [error]
//	@return [string] path
func DownloadFileByUrl(fileBelong, url string) (string, error) {
	log.Println("download file from online by url:", url)
	qmIndex := strings.LastIndex(url, "?")
	var rerr error
	if qmIndex != -1 {
		url = url[:qmIndex]
	}
	storepath := filepath.Join(baseStorePath, fileBelong, UniqueId()+filepath.Ext(url))
	v, err := http.Get(url)
	if err != nil {
		rerr = fmt.Errorf("http get failed(" + url + ")")
		log.Printf("Http get [%v] failed! %v\n", url, err)
		return "", rerr
	}
	defer v.Body.Close()
	content, err := ioutil.ReadAll(v.Body)
	if err != nil {
		rerr = fmt.Errorf("Read http response failed")
		log.Printf("Read http response failed! %v\n", err)
		return "", rerr

	}
	err = ioutil.WriteFile(storepath, content, 0666)
	path := ""
	if err != nil {
		path = ".\\" + storepath
	} else {
		rerr = fmt.Errorf("write to file failed:" + storepath)
	}
	return path, rerr
}

// 使用Chromedp下载链接内容
//
//	@param  fileBelong [string] 文件所属类型
//	@param  url [string]
//	@return [error]
//	@return [string] path
func DownloadUrlByChromedp(fileBelong, url string) (string, error) {
	log.Println("download file from online by url:", url)
	var rerr error = nil
	qmIndex := strings.LastIndex(url, "?")
	if qmIndex != -1 {
		url = url[:qmIndex]
	}
	storepath := filepath.Join(baseStorePath, fileBelong, UniqueId()+filepath.Ext(url))
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// set up a channel so we can block later while we monitor the download
	// progress
	done := make(chan bool)

	// this will be used to capture the request id for matching network events
	var requestID network.RequestID

	// set up a listener to watch the network events and close the channel when
	// complete the request id matching is important both to filter out
	// unwanted network events and to reference the downloaded file later
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			log.Printf("EventRequestWillBeSent: %v: %v", ev.RequestID, ev.Request.URL)
			if ev.Request.URL == url {
				requestID = ev.RequestID
			}
		case *network.EventLoadingFinished:
			log.Printf("EventLoadingFinished: %v", ev.RequestID)
			if ev.RequestID == requestID {
				close(done)
			}
		}
	})

	// all we need to do here is navigate to the download url
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
	); err != nil {
		log.Println(err)
		rerr = fmt.Errorf("run chromedp err")
	}

	// This will block until the chromedp listener closes the channel
	<-done
	// get the downloaded bytes for the request id
	var buf []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		buf, err = network.GetResponseBody(requestID).Do(ctx)
		return err
	})); err != nil {
		log.Println(err)
	}

	// write the file to disk - since we hold the bytes we dictate the name and
	// location
	if err := ioutil.WriteFile(storepath, buf, 0644); err != nil {
		log.Println(err)
	}
	log.Print("download file success, stored to ", storepath)
	return "/" + storepath, rerr
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 下载截图
//
//	@param  fileBelong [string] 文件所属类型
//	@param  videoname [string]
//	@param  file [*multipart.File]
//	@param  handeler [*multipart.FileHeader]
//	@return [error]
//	@return [string] path
func DownloadScreenshoot(fileBelong string, videoname string, file *multipart.File, handeler *multipart.FileHeader) (string, error) {
	var rerr error = nil
	fext := ".png"
	storepath := filepath.Join(baseStorePath, fileBelong)
	os.MkdirAll(storepath, os.ModePerm)
	findex := strings.LastIndex(videoname, ".")
	if findex != -1 {
		videoname = videoname[:findex]
	}
	for i := 1; ; i++ {
		temppath := filepath.Join(storepath, fmt.Sprintf("%s-%d%s", videoname, i, fext))
		if res, _ := PathExists(temppath); !res {
			storepath = temppath
			break
		}
	}
	storefile, err := os.OpenFile(storepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		rerr = fmt.Errorf("创建本地文件失败:" + storepath)
		log.Println(rerr, ", err:", err)
		return "", rerr
	}
	reader := bufio.NewReader(*file)
	writer := bufio.NewWriter(storefile)
	buffer := make([]byte, 256)
	for {
		_, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
		} else {
			writer.Write(buffer)
		}
	}
	writer.Flush()
	storefile.Close()
	log.Println("download file success:", storepath)
	return ".\\" + storepath, rerr
}
