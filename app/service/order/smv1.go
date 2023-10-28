package order

import (
	"encoding/json"
	"fmt"
	"goskeleton/app/model"
	"goskeleton/app/model/smv1"
	"goskeleton/app/utils/redis_factory"
	"goskeleton/app/utils/sign"
	"sync"
	"time"
)

type Smv1 struct {
}

var (
	copeOrder    = make(chan int, 3)
	originOrders []string
)

func (h *Smv1) Run(list_key string, secret string) {
	lock := &sync.RWMutex{}
	for {
		select {
		case <-copeOrder:
			lock.Lock()
			slen := LLEN(list_key)
			originOrders = Lrange(list_key, 0, slen)
			redisClient := redis_factory.GetOneRedisClient()
			payorder := model.Payorder{}
			for i := len(originOrders) - 1; i >= 0; i-- {
				order := originOrders[i]
				//1. 先从原始订单列表删除订单防止重复充值
				_, err := redisClient.Int64(redisClient.Execute("LREM", list_key, 0, order))
				if err != nil {
					failOrder(list_key, order)
					fmt.Println("订单出栈失败:", err)
					continue
				}

				if !sign.Verify(order, secret) {
					failOrder(list_key, order)
					fmt.Println("sign 验证失败")
					time.Sleep(10000 * time.Millisecond)

					continue
				}

				if err := json.Unmarshal([]byte(order), &payorder); err != nil {
					// failOrder(list_key, order)
					fmt.Println("订单序列化失败:", err)
					continue
				}

				// fmt.Println("生成SIGN:", sign.Create(order, "xadasadasdasas"))

				var ok bool
				var msg string
				switch payorder.Type {
				case 1:
					msg = "增款+$"
					ok = Recharge(payorder)
				case 2:
					msg = "减款-$"
					ok = Buyget(payorder)
				default:
					fmt.Println("未识别业务类型")
					continue
				}
				//2. 充值
				if ok {
					_, err := redisClient.Int64(redisClient.Execute("LPUSH", "OrderDone", order))
					if err != nil {
						fmt.Println("无法将订单加入已完成列表:", order)
						continue
					}
					time := time.Unix(payorder.Created_at, 0)
					fmt.Printf("交易成功|%s%f|订单号:%s|用户编号:%d|日期:%s \n", msg, payorder.Price, payorder.Orderid, payorder.Studentid, time.Format("2006-01-02 15:04:05"))
					fmt.Println("======================end==============================")
				} else {
					//交易失败
					failOrder(list_key, order)
					fmt.Println("=====================start=============================")
					fmt.Println("订单:", order)
					fmt.Printf("交易失败!|%s%f|订单号:%s|用户编号:%d \n", msg, payorder.Price, payorder.Orderid, payorder.Studentid)
					fmt.Println("======================end==============================")
				}
				// time.Sleep(5000 * time.Millisecond)
			}
			lock.Unlock()
			fmt.Printf("%d 条订单处理完成\n", len(originOrders))
		default:
			// fmt.Println("通道没有数据")
			slen := LLEN(list_key)
			if slen == 0 {
				// fmt.Println("列表内无订单等待...")
				time.Sleep(1 * time.Second)
				continue
			} else {
				fmt.Printf("接收到新订单 %d 条 \n", slen)
				time.Sleep(1 * time.Second)
				copeOrder <- 1 //处理订单
			}
		}
	}
}

func failOrder(list_key string, order string) {
	redisClient := redis_factory.GetOneRedisClient()
	redisClient.Int64(redisClient.Execute("LPUSH", list_key+"OrderFail", order))
}

func Recharge(order model.Payorder) bool {
	if smv1.MChargeRecordsFactory("").Add(order) {
		return true
	}
	return false
}

func Buyget(order model.Payorder) bool {
	if smv1.MealRecordsFactory("").Add(order) {
		return true
	}
	return false
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
	redisClient.ReleaseOneRedisClient()
	return res
}
