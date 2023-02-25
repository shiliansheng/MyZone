package models

import "log"

type Category struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Path   string `json:"path"`
	Module int    `json:"module"`
	State  int    `json:"state"`
}

func (Category) TableName() string {
	return "category"
}

func (m Category) Add(title string, module int, path string) RespData {
	log.Printf("添加Type: title(%s), module(%d), path(%s)\n", title, module, path)
	resp := NewRespData()
	if title == "" || module == 0 {
		resp.Msg = "添加 type 失败：缺号关键内容"
		log.Println(resp.Msg)
		return *resp
	}
	ctg := &Category{
		Title:  title,
		Module: module,
		Path:   path,
	}
	if _, err := Orm.Insert(ctg); err != nil {
		resp.Msg = "添加 Category 失败"
		log.Println(resp.Msg+":", err)
	} else {
		resp.Msg = "添加 Category 成功"
		log.Println(resp.Msg, *ctg)
	}
	return *resp
}

func (m Category) GetCategoryList(moudle int) RespData {
	resp := NewRespData()
	ctgs := []Category{}
	seter := Orm.QueryTable(m.TableName())
	if moudle != MODULE_DEFAULT {
		seter = seter.Filter("module", moudle)
	}
	count, err := seter.All(&ctgs)
	if err != nil {
		resp.Msg = "get category failed"
		log.Println("[ERROR]"+resp.Msg, err)
		return *resp
	}
	resp.Code = SUCCESS
	resp.Count = int(count)
	resp.Data = ctgs
	return *resp
}

func (this Category) SetMap() {
	CategoryIdMTitle = map[int]string{}
	ctgs := []Category{}
	Orm.QueryTable(this.TableName()).All(&ctgs)
	for _, c := range ctgs {
		CategoryIdMTitle[c.Id] = c.Title
	}
}
