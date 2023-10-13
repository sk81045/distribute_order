package order

import (
	// "encoding/json"
	"fmt"
	// "goskeleton/app/model"
	"goskeleton/app/model/smv1"
	"goskeleton/app/utils/redis_factory"
	"math/rand"
	"time"
)

type Hub struct {
}

func (h *Hub) Run() {
	fmt.Println("Order->启动")
	list_key := "nzh_"
	// Client("99", "ych_v2", 5)
	// var i = 0
	// SetList(list_key, 100)

	BRPopLPush(list_key)
	// time.Sleep(1000 * time.Millisecond)
	// }

	return
	for {
		slen := LLEN(list_key)
		if slen == 0 {
			// fmt.Println("订单列表空：", list_key)
			continue
		}
		payorder := Lrange(list_key, 0, slen)
		fmt.Println("订单列表：", payorder)

		redisClient := redis_factory.GetOneRedisClient()
		for _, v := range payorder {
			if smv1.EmployeeFactory("").Rechage(v) {
				//加入已完成订单
				bufferorder, err := redisClient.Int64(redisClient.Execute("LPUSH", list_key+"OrderDone", v))
				if err != nil {
					fmt.Println("SetList error:", err)
				}
				fmt.Println("LPUSH Success Add to Done_order", bufferorder)
				//从订单列表删除元素
				deleteOrder, err := redisClient.Int64(redisClient.Execute("LREM", list_key, 0, v))
				if err != nil {
					fmt.Println("deleteOrder error:", err)
				}
				fmt.Println("deleteOrder Success Add to Done_order", deleteOrder)

			}
			time.Sleep(1 * time.Second)
		}
		time.Sleep(110 * time.Second)
	}

}

func Client(sid string, apitype string, sleep_time int) {
	for {
		fmt.Println("服务启动,一卡通软件版本:")
		fmt.Println("sid", sid)
		fmt.Println("apitype", apitype)
		fmt.Println("sleep_time", sleep_time)
		time.Sleep(time.Duration(sleep_time) * time.Second)
	}
}

// 测试 redis 连接池
func SetList(list_key string, num int) {
	redisClient := redis_factory.GetOneRedisClient()
	for i := 1; i <= num; i++ {
		rand.Seed(time.Now().UnixNano())
		r := fmt.Sprintf("%d", rand.Intn(1000))
		fmt.Println(r)
		orderid := "JX2023092612193810003364" + r

		id := fmt.Sprintf("%d", i)
		list := `{"ID":` + id + `,"Sid":42,"Pid":0,"Lid":0,"Studentid":210049,"Ic":"3566612813","Orderid":"` + orderid + `","Macid":"150","Price":5,"Type":2,"From":"家校互联","Paystatus":false,"Category":"3","Sync":false,"Created_at":1695701978}`
		res, err := redisClient.Int64(redisClient.Execute("LPUSH", list_key, list))
		if err != nil {
			fmt.Println("err", err)
		} else {
			fmt.Println("success", res)
		}

	}
	redisClient.ReleaseOneRedisClient()
}

func BRPopLPush(list_key string) {
	redisClient := redis_factory.GetOneRedisClient()
	for {
		payorder, err := redisClient.String(redisClient.Execute("RPOPLPUSH", list_key, list_key+"OrderBuffer"))

		if err != nil {
			continue
		} else {
			// redisClient.Int64(redisClient.Execute("LPUSH", list_key+"bacd", payorder)) //debug
			if smv1.EmployeeFactory("").Rechage(payorder) {
				bufferorder, _ := redisClient.String(redisClient.Execute("RPOPLPUSH", list_key+"OrderBuffer", "OrderDone"))
				// fmt.Println("RPOPLPUSH Success Add to buffer_order", bufferorder)
				fmt.Println("交易成功", bufferorder)
			} else {
				failerorder, _ := redisClient.String(redisClient.Execute("RPOPLPUSH", list_key+"OrderBuffer", "OrderFail"))
				fmt.Println("交易失败", failerorder)
			}
		}

		time.Sleep(1000 * time.Millisecond)
	}
	// redisClient.ReleaseOneRedisClient()
}

func LLEN(list_key string) int64 {
	redisClient := redis_factory.GetOneRedisClient()
	res, err := redisClient.Int64(redisClient.Execute("LLEN", list_key))
	if err != nil {
		fmt.Println("err", err)
	}
	return res
}

// LLEN KEY_NAME
func Lrange(list_key string, START int64, END int64) []string {
	redisClient := redis_factory.GetOneRedisClient()
	res, err := redisClient.Strings(redisClient.Execute("LRANGE", list_key, START, END))
	// var order = make([]model.Payorder, END-START+1)
	if err != nil {
		fmt.Println("Lrange err", err)
	}
	// } else {
	// 	for k, v := range res {
	// 		if err := json.Unmarshal([]byte(v), &order[k]); err != nil {
	// 			fmt.Errorf("json unma err:%v", err)
	// 		}
	// 	}
	// }
	redisClient.ReleaseOneRedisClient()
	return res
}
