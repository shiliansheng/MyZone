package models

import (
	"fmt"
	"log"
	"myzone/mzutils"
)

type Record struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Category int    `json:"category"`
	Content  string `json:"content"`
	Top      int    `json:"top"`
	State    int    `json:"state"`
	Addtime  string `json:"addtime"`
	Deltime  string `json:"deltime"`
}

func (Record) TableName() string {
	return "record"
}

// 获取Record List
//
//	@param  state [int] record状态，有效，无效，默认
//	@param  category [int] record类型
//	@return [RespData] RespData.Data = recordlist
func (m Record) GetRecordList(state, category int) RespData {
	log.Println("[RECORD LIST] get record list with state[", state, "]")
	resp := NewRespData()
	reclist := []Record{}
	seter := Orm.QueryTable(m.TableName())
	if state != STATE_DEFAULT {
		seter = seter.Filter("state", state)
	}
	if category != CATEGORY_ALL {
		seter = seter.Filter("category", category)
	}
	if _, err := seter.All(&reclist); err != nil {
		resp.Msg = "get record list failed"
		log.Println(resp.Msg+":", err)
	}
	resp.Code = SUCCESS
	resp.Data = reclist
	return *resp
}

func (this Record) Get(id int) RespData {
	log.Println("[GET RECORD] ID[", id, "]")
	resp := NewRespData()
	rec := &Record{Id: id}
	if err := Orm.Read(rec); err != nil {
		resp.Msg = fmt.Sprintf("get record ID[%d] failed: %v", id, err)
		log.Println(resp.Msg)
		return *resp
	}
	resp.Data = *rec
	resp.Code = SUCCESS
	return *resp
}

func (this Record) Add(category, top int, title, detail string, content string) RespData {
	log.Printf("[ADD RECORD] title[%s] category[%d] detail[%s] content[%s]", title, category, detail, content)
	resp := NewRespData()
	if title == "" {
		resp.Msg = "所需内容为空"
		log.Println(resp.Msg)
		return *resp
	}
	rec := &Record{
		Title:    title,
		Category: category,
		Top:      top,
		Detail:   detail,
		Content:  content,
		Addtime:  mzutils.NowTimeString(),
		State:    VALID,
	}
	if err := Orm.Read(rec, "title"); err == nil {
		log.Println("[ADD RECOTD]", title, "is existed, id =", rec.Id)
		resp.Data = *rec
		resp.Msg = fmt.Sprintf("<span class='bg-info'>%s</span> 已存在！", title)
		resp.Code = EXISTED
		return *resp
	}
	recId, err := Orm.Insert(rec)
	if err != nil {
		resp.Msg = "add record failed [" + title + "]"
		log.Println("[ADD RECORD]", resp.Msg, err)
	} else {
		rec.Id = int(recId)
		resp.Code = SUCCESS
		resp.Data = *rec
	}
	return *resp
}

func (this Record) Update(record *Record, cols ...string) RespData {
	log.Printf("[UPDATE RECORD] id[%d] cols%v", record.Id, cols)
	resp := NewRespData()
	if len(cols) == 0 || record.Id == 0 {
		resp.Msg = "更新列或ID为空"
		log.Println(resp.Msg)
		return *resp
	}
	if _, err := Orm.Update(record, cols...); err != nil {
		resp.Msg = fmt.Sprint("update record failed:", err)
		log.Println(resp.Msg)
	} else {
		resp.Msg = "更新记录成功"
		resp.Code = SUCCESS
	}
	return *resp
}

func (this Record) Delete(id int) RespData {
	log.Printf("[DELETE RECORD] id[%d]", id)
	resp := NewRespData()
	rec := &Record{Id: id, State: INVALID, Deltime: mzutils.NowTimeString()}
	*resp = this.Update(rec, "state", "deltime")
	return *resp
}
