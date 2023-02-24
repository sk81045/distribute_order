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
		llen := redis.LLen(list_key)
		order_list := redis.LRange(list_key, llen-1, llen)
		if len(order_list) == 0 {
			continue
		}

		var ok bool
		if apitype == "api" {
			fmt.Println("api")
			ok = api.Mission(order_list[0])

		} else {
			fmt.Println("soap")
			ok = soap.Mission(order_list[0])
		}

		if ok {
			redis.LRpop(list_key)
			// global.H_LOG.Info("Successfully->", zap.String("", order_list[0]))
		} else {
			redis.BRPopLPush(list_key, list_key+"err", 5*time.Second)
			global.H_LOG.Warn("Order failed!", zap.String("", order_list[0]))
		}

		time.Sleep(1 * time.Second)
	}
}
