package models

import (
	"bufio"
	"io"
	"log"
	"mime/multipart"
	"myzone/mzutils"
	"os"
	"path/filepath"
)

// 下载截图
//
//	@param  fileBelong [string] 文件所属类型
//	@param  file [*multipart.File]
//	@param  handeler [*multipart.FileHeader]
//	@return [RespData]
func DownloadFile(fileBelong string, file *multipart.File, handeler *multipart.FileHeader) RespData {
	log.Println("[DOWNLOAD FILE] start......")
	resp := NewRespData()
	storepath := filepath.Join(STORE_PATH, fileBelong, handeler.Filename)
	os.MkdirAll(filepath.Join(STORE_PATH, fileBelong), os.ModePerm)
	if ok, err := mzutils.PathExists(storepath); ok {
		if err != nil {
			log.Println(err)
		}
		resp.Msg = "<span class=\"bg-info\">FILE[ " + handeler.Filename + " ]已存在</span>"
		return *resp
	}
	storefile, err := os.OpenFile(storepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		resp.Msg = "创建本地文件失败:" + storepath
		log.Println(resp.Msg, ", err:", err)
		return *resp
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
	resp.Code = SUCCESS
	resp.Data = "/" + storepath
	storefile.Close()
	resp.Msg = "<span class=\"bg-info\">下载文件 [ " + handeler.Filename + " ] 成功！</span>"

	log.Println("[DOWNLOAD FILE] success:", storepath)
	return *resp
}
