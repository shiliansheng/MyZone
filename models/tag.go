package models

import (
	"fmt"
	"log"
)

type Tag struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Recommend int    `json:"recommend" orm:"default(0)"`
	// Viewnum   int    `json:"viewnum"`
	// Categoryid int    `json:"categoryid"`
}

func (Tag) TableName() string {
	return "tag"
}

// 获取Tag List
//
//	@param  ids [...string]
//	@return [RespData] RespData.Data = taglist
func (m Tag) GetTagList(ids ...string) RespData {
	// log.Println("[TAG LIST] get tag list")
	resp := NewRespData()
	taglist := []Tag{}
	seter := Orm.QueryTable(m.TableName())
	if len(ids) != 0 {
		seter = seter.Filter("id__in", ids)
	}
	if _, err := seter.All(&taglist); err != nil {
		resp.Msg = "get tag list failed"
		log.Println(resp.Msg+":", err)
	}
	resp.Code = SUCCESS
	resp.Data = taglist
	return *resp
}

func (m Tag) GeTagName(id int) string {
	tag := Tag{Id: id}
	if err := Orm.Read(&tag); err != nil {
		log.Println("get tag list failed:", err)
		return ""
	}
	return tag.Name
}

// 通过name添加tag返回添加后的tag
//
//	@param  name [string]
//	@return [RespData] RespData.Data = tag
func (m Tag) Add(name string) RespData {
	log.Println("[Tag] add tag with name:", name)
	resp := NewRespData()
	tag := &Tag{Name: name}
	if err := Orm.Read(tag, "name"); err == nil {
		log.Println("TAG]", name, "is existed, id =", tag.Id)
		resp.Data = *tag
		resp.Msg = fmt.Sprintf("<span class='bg-info'>%s</span> 已存在！", name)
		resp.Code = EXISTED
		return *resp
	}
	// tag.Pubtime = mzutils.NowTimeString()
	tagId, err := Orm.Insert(tag)
	if err != nil {
		resp.Msg = "add tag failed [-" + name + "]"
		log.Println("[ACTOR]", resp.Msg, err)
	} else {
		tag.Id = int(tagId)
		resp.Code = SUCCESS
		resp.Data = *tag
	}
	return *resp
}

func (m Tag) GetTagId(name string) int {
	log.Println("[GET TAG ID] name:", name)
	tag := &Tag{Name: name}
	Orm.Read(tag, "name")
	log.Println("actorid:", tag)
	return tag.Id
}
