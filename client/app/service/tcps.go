package service

import (
	redis_method "Hwgen/app/controller/redis"
	sellfood "Hwgen/app/controller/sellfood"
	"Hwgen/global"
	// helpers "Hwgen/utils"
	"fmt"
	"go.uber.org/zap"
	"time"
)

var (
	redis redis_method.RedisStore
	api   sellfood.Api
	soap  sellfood.Soap
	lock  = false
)

type Message struct {
	Type     int
	Describe string
	Content  string
	Pid      string
	Sid      int
}

func Run() {
	system := global.H_CONFIG.System
	sid := system.SchoolId            //学校ID
	apitype := system.SellfoodApiType //售饭接口类型
	Client(sid, apitype)
}

func Client(sid string, apitype string) {
	fmt.Println("服务启动,通信接口类型:", global.H_CONFIG.System.SellfoodApiType)
	list_key := sid //学校ID
	for {
		if lock {
			fmt.Println("接口错误: var lock = true")
			continue
		}
		if redis.LLen(list_key) == 0 {
			continue
		}

		order_list := redis.BRPopLPush(list_key, list_key+"ChargeBuffer", 5*time.Second)
		fmt.Println("order_list", order_list)
		ok := false
		lock = true
		if apitype == "api" {
			fmt.Println("type->api")
			ok = api.Mission(order_list)
			lock = false
		} else {
			fmt.Println("type->soap")
			ok = soap.Mission(order_list)
			lock = false
		}

		if ok {
			redis.LRpop(list_key + "ChargeBuffer")
		} else {
			redis.BRPopLPush(list_key+"ChargeBuffer", list_key+"ChargeErr", 5*time.Second)
			global.H_LOG.Warn("Order failed!", zap.String("", order_list))
		}
		time.Sleep(5 * time.Second)
	}
}
