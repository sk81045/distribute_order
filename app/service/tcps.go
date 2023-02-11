package service

import (
	redis_method "Hwgen/app/controller/redis"
	sellfood "Hwgen/app/controller/sellfood"
	"Hwgen/app/model"
	"Hwgen/global"
	// helpers "Hwgen/utils"
	"encoding/json"
	"fmt"
	"time"
)

var (
	redis redis_method.RedisStore
	api   sellfood.Api
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

	list_key := "34"
	for {
		llen := redis.LLen(list_key)
		order_list := redis.LRange(list_key, llen-1, llen)
		if len(order_list) == 0 {
			continue
		}
		// fmt.Println("LRange-->", order_list)
		// redis.LRpop(list_key)
		// redis.BRPopLPush(list_key, 5*time.Second)
		ok := api.Mission(order_list[0])
		if ok {
			redis.LRpop(list_key)
		} else {
			redis.BRPopLPush(list_key, list_key+"err", 5*time.Second)
			// redis.SetList("47err", order_list[0])
			fmt.Println("操作失败-->", ok)
		}

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
			ok := redis.SetList(fmt.Sprintf("%d", sid), string(items))
			if ok {
				fmt.Println("redis SetList ok")
				Save(i.ID)
			} else {
				fmt.Println("redis setlist err")
			}
		}
	}
}

func Save(id int) {
	result := global.H_DB.Model(&model.Payorder{}).Where("id = ?", id).Update("sync", 1)
	fmt.Println("Save-->", result.Error)
}

func Payorder(sid int) []model.Payorder {
	var list = []model.Payorder{}
	_ = global.H_DB.Model(&model.Payorder{}).Preload("Students").Where("`sid` = ? AND `sync` = ? AND `price` > ?", sid, 0, 0).Limit(5).Find(&list).Error
	return list
}

func School() []model.School {
	var school_list = []model.School{}
	_ = global.H_DB.Model(&model.School{}).Find(&school_list).Error
	return school_list
}
