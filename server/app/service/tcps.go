package service

import (
	redis_method "Hwgen/app/controller/redis"
	sellfood "Hwgen/app/controller/sellfood"
	"Hwgen/app/model"
	"Hwgen/global"
	// helpers "Hwgen/utils"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
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
	// system := global.H_CONFIG.System
	// sid := system.SchoolId //学校ID
	// fmt.Println("User:", system.School, "ID:", sid)
	// Client(sid)
}

func Client(sid string) {
	list_key := sid //学校ID
	for {
		llen := redis.LLen(list_key)
		order_list := redis.LRange(list_key, llen-1, llen)
		if len(order_list) == 0 {
			continue
		}
		ok := api.Mission(order_list[0])
		if ok {
			redis.LRpop(list_key)
			global.H_LOG.Info("Successfully->", zap.String("", order_list[0]))
		} else {
			redis.BRPopLPush(list_key, list_key+"err", 5*time.Second)
			global.H_LOG.Warn("Order failed!", zap.String("", order_list[0]))
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
	for {

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
				Save(i.ID)
				fmt.Printf("Order added to queue successfully id:%d, school_id:%d\n", i.ID, i.Sid)
			} else {
				fmt.Println("Redis setlist failed id:", i.ID)
				global.H_LOG.Warn("订单加入队列失败!:", zap.String("redis setlist method err:", string(items)))
			}
		}
	}
}

func Save(id int) {
	global.H_DB.Model(&model.Payorder{}).Where("id = ?", id).Update("sync", 1)
}

func Payorder(sid int) []model.Payorder {
	var list = []model.Payorder{}
	_ = global.H_DB.Model(&model.Payorder{}).Where("`sid` = ? AND `sync` = ? AND `price` > ?", sid, 0, 0).Limit(50).Find(&list).Error
	return list
}

func School() []model.School {
	var school_list = []model.School{}
	_ = global.H_DB.Model(&model.School{}).Find(&school_list).Error
	return school_list
}
