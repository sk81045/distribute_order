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
	"sync"
	"time"
)

var (
	redis redis_method.RedisStore
	api   sellfood.Api
	mu    sync.Mutex
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
}

func Server() {
	school_list_key := School()
	for _, s := range school_list_key {
		time.Sleep(2 * time.Second)
		go Addlist(s.ID)
		fmt.Println("客户端加入-->", s.Name)
	}
	// go Addlist(34)
	for {

	}
}

func Addlist(sid int) {
	for {
		time.Sleep(10 * time.Second)

		list := Payorder(sid)
		for _, i := range list {
			items, _ := json.Marshal(i)
			ok := redis.SetList(fmt.Sprintf("%d", sid), string(items))
			if ok {
				SaveSync(i.ID)
				fmt.Printf("Order added to queue successfully id:%d, school_id:%d\n", i.ID, i.Sid)
			} else {
				redis.BRPopLPush(fmt.Sprintf("%d", sid), fmt.Sprintf("%d", sid)+"SetListErr", 5*time.Second)
				fmt.Println("Redis setlist failed id:", i.ID)
				global.H_LOG.Warn("订单加入队列失败!:", zap.String("redis setlist method err:", string(items)))
			}
		}
	}
}

// func ErrOrder() {
// 	for {
// 		school_list_key := School()
// 		for _, s := range school_list_key {
// 			go Addlist(s.ID)
// 			fmt.Println("学校-->", s.Name)
// 		}
// 	}
// }

func SaveSync(id int) {
	mu.Lock()
	defer mu.Unlock()
	global.H_DB.Model(&model.Payorder{}).Where("id = ?", id).Update("sync", 1)
}

func Payorder(sid int) []model.Payorder {
	mu.Lock()
	defer mu.Unlock()
	var list = []model.Payorder{}
	_ = global.H_DB.Model(&model.Payorder{}).Where("`paystatus` = ? AND `sid` = ? AND `sync` = ? AND `price` > ?", 0, sid, 0, 0).Limit(50).Find(&list).Error
	return list
}

func School() []model.School {
	var school_list = []model.School{}
	_ = global.H_DB.Model(&model.School{}).Find(&school_list).Error
	return school_list
}
