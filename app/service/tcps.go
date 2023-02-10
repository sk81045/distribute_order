package service

import (
	"Hwgen/app/controller"
	"Hwgen/app/model"
	"Hwgen/global"
	// helpers "Hwgen/utils"
	"encoding/json"
	"fmt"

	"time"
)

var (
	redis controller.RedisStore
	api   controller.Api
)

type Message struct {
	Type     int
	Describe string
	Content  string
	Pid      string
	Sid      int
}

func Run() {

	Server()

	list_key := "47"
	for {
		llen := redis.LLen(list_key)
		order_list := redis.LRange(list_key, llen-1, llen)
		if len(order_list) == 0 {
			continue
		}
		// fmt.Println("LRange-->", order_list)
		// redis.LRpop(list_key)
		// redis.BRPopLPush(list_key, 5*time.Second)
		res := api.Rechage(order_list[0])
		// api.Token()
		if res == "ok" {
			redis.LRpop(list_key)
		} else {
			redis.BRPopLPush(list_key, list_key+"err", 5*time.Second)

			// redis.SetList("47err", order_list[0])
		}
		fmt.Println("res-->", res)

		time.Sleep(3 * time.Second)
	}
}

func Server() {

	school_list_key := School()
	for _, s := range school_list_key {

		go Addlist(s.ID)
		fmt.Println("学校-->", s.Name)
	}

}

func Addlist(sid int) {
	for {
		time.Sleep(5 * time.Second)

		list := Payorder(sid)
		for _, i := range list {
			items, _ := json.Marshal(i)
			redis.SetList(fmt.Sprintf("%d", sid), string(items))
			fmt.Println("Save-->", i)
			Save(i.ID)
		}
	}
}

func Save(id int) {
	global.H_DB.Model(&model.Payorder{}).Where("id = ?", id).Update("sync", 1)
}

func Payorder(sid int) []model.Payorder {
	var list = []model.Payorder{}
	_ = global.H_DB.Model(&model.Payorder{}).Where("`sid` = ? AND `sync` = ? ", sid, 0).Limit(5).Find(&list).Error
	// if list.ID == 0 {
	// 	// panic("func Payorder():没有获取到此设备信息")
	// }
	return list
}

func School() []model.School {
	var school_list = []model.School{}
	_ = global.H_DB.Model(&model.School{}).Find(&school_list).Error
	// if list.ID == 0 {
	// 	// panic("func Payorder():没有获取到此设备信息")
	// }
	return school_list
}
