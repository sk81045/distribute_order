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
	FailCope    chan string
	Order       string
	Count       []int64
	ResendNum   int64
	ResendTime  int64
	RedisClient *redis_factory.RedisClient
}

func (h *Smv1) Run(list_key string, secret string, order_config interface{}) {
	/**********************************任务初始化*************************/
	h.List_key = list_key
	h.Cope = make(chan int, 3)        //任务执行信号
	h.FailCope = make(chan string, 1) //异常订单执行信号
	h.Count = make([]int64, 3)        //任务计数器
	resendnum, _ := strconv.ParseInt(fmt.Sprintf("%d", order_config.(map[string]interface{})["resendnum"]), 10, 64)
	h.ResendNum = resendnum
	resentime, _ := strconv.ParseInt(fmt.Sprintf("%d", order_config.(map[string]interface{})["resendtime"]), 10, 64)
	h.ResendTime = resentime
	/*******************************************************************/
	go h.FailOrderProcess()

	for {
		select {
		case <-h.Cope:
			redisClient := h.RedisClient
			payorder := model.Payorder{}
			if err := json.Unmarshal([]byte(h.Order), &payorder); err != nil {
				fmt.Println("序列化失败", err)
				h.BadOrderProcess(payorder, err)
				continue
			}

			if payorder.Mark != "RESEND" { //重发被标记的订单不做sign验证
				if err := sign.Verify(h.Order, secret); err != nil {
					fmt.Println("Sign验证失败")
					h.BadOrderProcess(payorder, err)
					continue
				}
			} else {
				fmt.Println("重发被标记的订单不做sign验证")
			}

			order := h.Order
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

				fmt.Printf("\033[7;32;2m交易成功a|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")

				h.Count[0]++
			} else { //交易失败
				redisClient.Int64(redisClient.Execute("LPUSH", list_key+"OrderFail", order)) //加入异常列表

				//！！！异常订单处理 在这里设置重发次数和时间间隔
				fmt.Printf("\033[7;31;2m交易失败!|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")
				fmt.Println("重发次数", payorder.Resend)
				if payorder.Resend >= h.ResendNum {
					fmt.Printf("\033[7;33;2m重试次数超出限制 %d 已终止\033[0m\n", h.ResendNum)
					h.BadOrderProcess(payorder, pay_error)
					continue
				}
				redisClient.ReleaseOneRedisClient()
				time.Sleep(time.Duration(payorder.Resend*100) * time.Millisecond)
				h.FailCope <- order
			}

			redisClient.ReleaseOneRedisClient()
			time.Sleep(time.Duration(h.ResendTime) * time.Millisecond)
			fmt.Printf("操作成功 %d 条订单\n", h.Count[0])
		default:
			h.RedisClient = redis_factory.GetOneRedisClient()
			redisClient := h.RedisClient
			res, err := redisClient.Bytes(redisClient.Execute("BRPOPLPUSH", list_key, list_key+"Backups", 30))
			if err != nil {
				fmt.Println("读取新订单列表数据", err)
				continue
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
	redisClient.Int64(redisClient.Execute("LPUSH", "BadOrder", string(bad)))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key, 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"OrderFail", 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, h.Order))
	redisClient.ReleaseOneRedisClient()
}

func (h *Smv1) FailOrderProcess() {
	for {
		select {
		case fail_order := <-h.FailCope:
			fmt.Println("尝试处理异常订单...", fail_order)

			order := model.Payorder{}
			if err := json.Unmarshal([]byte(fail_order), &order); err != nil {
				fmt.Println("异常订单序列化失败", err)
				h.BadOrderProcess(order, err)
			}

			redisClient := redis_factory.GetOneRedisClient()
			// defer
			_, err := redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, fail_order))
			if err != nil {
				fmt.Println("从备份里取出异常订单", err)
				h.BadOrderProcess(order, err)
			}
			_, err = redisClient.Int64(redisClient.Execute("LREM", h.List_key+"OrderFail", 0, fail_order))
			if err != nil {
				fmt.Println("取出异常订单", err)
				h.BadOrderProcess(order, err)
			}

			order.Resend = order.Resend + 1
			order.Mark = "RESEND" //重发订单标记
			new_fail_order_str, _ := json.Marshal(order)
			_, err = redisClient.Int64(redisClient.Execute("RPUSH", h.List_key, new_fail_order_str))
			if err != nil {
				fmt.Println("异常订单重新加入列表失败", err)
			}
			redisClient.ReleaseOneRedisClient()
		}
	}
}
