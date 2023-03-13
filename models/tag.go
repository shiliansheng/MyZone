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

func (m Tag) Get(id int) RespData {
	log.Println("[GET TAG] ID:", id)
	resp := NewRespData()
	tag := Tag{Id: id}
	if err := Orm.Read(&tag); err != nil {
		resp.Msg = fmt.Sprintf("获取标签[%d]信息失败!", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	resp.Data = tag
	resp.Code = SUCCESS
	resp.Msg = "获取标签信息成功!"
	return *resp
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

func (m Tag) Update(tag *Tag, cols ...string) RespData {
	log.Println("[TAG UPDATE]", *tag, " cols:", cols)
	resp := NewRespData()

	if tag.Name == "" {
		resp.Msg = "更新名称为空，更新失败："
	}
	if _, err := Orm.Update(tag, cols...); err != nil {
		resp.Msg = "update tag failed"
		log.Println(resp.Msg, err)
	} else {
		resp.Msg = "更新成功！"
		resp.Data = tag.Name
		resp.Code = SUCCESS
	}
	return *resp
}

func (m Tag) Delete(id int) RespData {
	log.Println("[ACTOR DELETE] ID:", id)
	resp := NewRespData()
	tag := Tag{Id: id}
	if _, err := Orm.Delete(&tag); err != nil {
		resp.Msg = fmt.Sprintf("删除标签[%d]失败!", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	resp.Msg = fmt.Sprintf("删除标签[%d]成功!", id)
	resp.Code = SUCCESS
	return *resp
}
