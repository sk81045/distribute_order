package order

import (
	"encoding/json"
	"fmt"
	"goskeleton/app/model"
	"goskeleton/app/model/smv1"
	"goskeleton/app/utils/redis_factory"
	"goskeleton/app/utils/sign"
	"strconv"
	"time"
)

type Smv1 struct {
	List_key    string
	Cope        chan int
	Order       string
	Count       []int64
	ResendNum   int64
	ResendTime  int64
	RedisClient *redis_factory.RedisClient
}

func (h *Smv1) Run(list_key string, secret string, order_config interface{}) {
	/**********************************任务初始化*************************/
	h.List_key = list_key
	h.Cope = make(chan int, 3) //任务执行信号
	h.Count = make([]int64, 3) //任务计数器
	resendnum, _ := strconv.ParseInt(fmt.Sprintf("%d", order_config.(map[string]interface{})["resendnum"]), 10, 64)
	h.ResendNum = resendnum
	resentime, _ := strconv.ParseInt(fmt.Sprintf("%d", order_config.(map[string]interface{})["resendtime"]), 10, 64)
	h.ResendTime = resentime
	h.RedisClient = redis_factory.GetOneRedisClient()
	/*******************************************************************/
	go h.FailOrderProcess()

	redisClient := h.RedisClient
	for {
		select {
		case <-h.Cope:
			payorder := model.Payorder{}
			if err := json.Unmarshal([]byte(h.Order), &payorder); err != nil {
				continue
			}

			if !sign.Verify(h.Order, secret) {
				h.BadOrderProcess(payorder, fmt.Errorf("Sign验证失败!"))
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			order := h.Order
			if err := json.Unmarshal([]byte(order), &payorder); err != nil {
				fmt.Println("订单序列化失败:", err)
				continue
			}

			var pay_error error
			switch payorder.Type {
			case 1:
				pay_error = smv1.MChargeRecordsFactory("").Add(payorder) //recharge
			case 2:
				pay_error = smv1.MealRecordsFactory("").Add(payorder) //decline
			default:
				fmt.Println("未识别业务类型")
				continue
			}
			//2. 充值
			paytime := time.Unix(payorder.Dealtime, 0)
			if pay_error == nil {
				_, err := redisClient.Int64(redisClient.Execute("LPUSH", "OrderDone", order))
				if err != nil {
					fmt.Println("无法将订单加入已完成列表:", order)
					continue
				}

				fmt.Printf("\033[7;32;2m交易成功|操作%d|$%f|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")

				h.Count[0]++
			} else { //交易失败
				h.RedisClient.Int64(h.RedisClient.Execute("LPUSH", list_key+"OrderFail", order))
				fmt.Printf("\033[7;31;2m交易失败!|操作%d|$%f|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")
				h.Count[1]++

				//！！！异常订单处理 在这里设置重发次数和时间间隔
				time.Sleep(time.Duration(h.ResendTime) * time.Duration(h.Count[1]) * time.Millisecond)
				if h.Count[1] == h.ResendNum {
					fmt.Printf("重试次数超出限制机 %d 加入错误订单\n", h.Count[1])
					h.BadOrderProcess(payorder, pay_error)
					h.Count[1] = 0
				}
			}
			time.Sleep(10 * time.Millisecond)
			fmt.Printf("操作成功 %d 条订单\n", h.Count[0])
		default:
			redisClient := redis_factory.GetOneRedisClient()
			res, err := redisClient.Bytes(redisClient.Execute("BRPOPLPUSH", list_key, list_key+"Backups", 0))
			if err != nil {
				fmt.Println("err", err)
			}
			h.Order = string(res)
			h.Cope <- 1 //处理订单
		}
	}
}
func (h *Smv1) BadOrderProcess(bad_order model.Payorder, err_msg error) {
	redisClient := h.RedisClient
	bad_order.Error = err_msg.Error()
	bad, _ := json.Marshal(bad_order)
	redisClient.Int64(redisClient.Execute("LPUSH", h.List_key+"Bad", string(bad)))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key, 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"OrderFail", 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, h.Order))
	// redisClient.ReleaseOneRedisClient()
}

func (h *Smv1) FailOrderProcess() {
	fmt.Println("等待完成")
	redisClient := redis_factory.GetOneRedisClient()
	for {
		select {
		case <-h.Cope:
			fmt.Println("等待正常订单处理完成")
			continue
		default:
			fmt.Println("同步处理异常订单.....")

			res, err := redisClient.StringMap(redisClient.Execute("BRPOP", h.List_key+"OrderFail", 0)) //阻塞等待list

			if err != nil {
				fmt.Println("处理异常订单失败:1", err)
			}
			_, err = redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, res[h.List_key+"OrderFail"]))
			if err != nil {
				fmt.Println("异常订单出栈失败:2", err)
			}

			_, err = redisClient.Int64(redisClient.Execute("RPUSH", h.List_key, res[h.List_key+"OrderFail"]))
			if err != nil {
				fmt.Println("处理异常订单失败:3", err)
			}
		}
	}
}
