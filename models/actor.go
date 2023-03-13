package models

import (
	"fmt"
	"log"
	"myzone/mzutils"
)

type Actor struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Cover     string `json:"cover"`
	Collect   int    `json:"collect" orm:"default(0)"`
	Pubtime   string `json:"pubtime"`
	Recommend int    `json:"recommend" orm:"default(0)"`
}

func (Actor) TableName() string {
	return "actor"
}

func (m Actor) Get(id int) RespData {
	log.Println("[GET ACTOR] ID:", id)
	resp := NewRespData()
	actor := Actor{Id: id}
	if err := Orm.Read(&actor); err != nil {
		resp.Msg = fmt.Sprintf("获取演员[%d]信息失败!", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	resp.Data = actor
	resp.Code = SUCCESS
	resp.Msg = "获取演员信息成功!"
	return *resp
}

// 获取Actor List
//
//	@param  ids [...string] string类型id数组
//	@return [RespData] RespData.Data = actorlist
func (m Actor) GetActorList(ids ...string) RespData {
	// log.Println("[ACTOR LIST] get actor list")
	resp := NewRespData()
	actorlist := []Actor{}
	seter := Orm.QueryTable(m.TableName())
	if len(ids) != 0 {
		seter = seter.Filter("id__in", ids)
	}
	if _, err := seter.All(&actorlist); err != nil {
		resp.Msg = "get actor list failed"
		log.Println(resp.Msg+":", err)
	}
	resp.Code = SUCCESS
	resp.Data = actorlist
	return *resp
}

// func (m Actor) GetActorNameJson(idArr ...[]string) string {
// 	actorinfos := []ActorInfo{}
// 	actors := []Actor{}
// 	if _, err := Orm.QueryTable(m.TableName()).Filter("id__in", idArr).All(&actors); err != nil {
// 		log.Println("get actor list failed:", err)
// 		return ""
// 	}
// 	for _, pie := range actors {
// 		actorinfos = append(actorinfos, ActorInfo{
// 			Id:   pie.Id,
// 			Name: pie.Name,
// 		})
// 	}
// 	bytes, _ := json.Marshal(actorinfos)
// 	return string(bytes)
// }

func (m Actor) GetActorName(id int) string {
	log.Println("[ACTOR] get actor name:", id)
	actor := &Actor{Id: id}
	if err := Orm.Read(actor); err != nil {
		log.Println("get actor list failed:", err)
		return ""
	}
	return actor.Name
}

// 通过name添加actor返回添加后的actor
//
//	@param  name [string]
//	@return [RespData] RespData.Data = actor
func (m Actor) Add(name string) RespData {
	log.Println("[ACTOR] add actor with name:", name)
	resp := NewRespData()
	actor := &Actor{Name: name, Collect: DATA_INIT}
	if err := Orm.Read(actor, "name"); err == nil {
		log.Println("[ACTOR]", name, "is existed, id =", actor.Id)
		resp.Data = *actor
		resp.Msg = fmt.Sprintf("<span class='bg-info'>%s</span> 已存在！", name)
		resp.Code = EXISTED
		return *resp
	}
	actor.Pubtime = mzutils.NowTimeString()
	actorId, err := Orm.Insert(actor)
	if err != nil {
		resp.Msg = "add actor failed [" + name + "]"
		log.Println("[ACTOR]", resp.Msg, err)
	} else {
		actor.Id = int(actorId)
		resp.Code = SUCCESS
		resp.Data = *actor
	}
	return *resp
}

func (m Actor) GetActorId(name string) int {
	log.Println("[GET ACTOR ID] name:", name)
	actor := &Actor{Name: name}
	Orm.Read(actor, "name")
	log.Println("actorid:", actor)
	return actor.Id
}

func (this Actor) GetNoVideoActor() RespData {
	log.Println("[GET NO VIDEO ACTOR LIST]")
	resp := NewRespData()
	actors := []Actor{}
	novActors := []Actor{}
	Orm.QueryTable(this.TableName()).All(&actors)
	for _, actor := range actors {
		count, _ := Orm.QueryTable(new(Video).TableName()).Filter("state", VALID).Filter("actorid__contains", fmt.Sprintf("\"%d\"", actor.Id)).Count()
		if count == 0 {
			// log.Printf("%3d %s\n", actor.Id, actor.Name)
			novActors = append(novActors, actor)
		}
	}
	resp.Data = novActors
	resp.Code = SUCCESS
	return *resp
}

func (m Actor) Update(actor *Actor, cols ...string) RespData {
	log.Println("[ACTOR UPDATE]", *actor, " cols:", cols)
	resp := NewRespData()

	if actor.Name == "" {
		resp.Msg = "更新名称为空，更新失败："
	}
	if _, err := Orm.Update(actor, cols...); err != nil {
		resp.Msg = "update actor failed"
		log.Println(resp.Msg, err)
	} else {
		resp.Msg = "更新成功！"
		resp.Data = actor.Name
		resp.Code = SUCCESS
	}
	return *resp
}

func (m Actor) Delete(id int) RespData {
	log.Println("[ACTOR DELETE] ID:", id)
	resp := NewRespData()
	actor := Actor{Id: id}
	if _, err := Orm.Delete(&actor); err != nil {
		resp.Msg = fmt.Sprintf("删除演员[%d]失败!", id)
		log.Println(resp.Msg, err)
		return *resp
	}
	resp.Msg = fmt.Sprintf("删除演员[%d]成功!", id)
	resp.Code = SUCCESS
	return *resp
}
