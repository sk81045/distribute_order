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
				fmt.Println("序列化失败", err)
				h.BadOrderProcess(payorder, err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if err := sign.Verify(h.Order, secret); err != nil {
				fmt.Println("Sign验证失败")
				h.BadOrderProcess(payorder, err)
				time.Sleep(100 * time.Millisecond)
				continue
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

				fmt.Printf("\033[7;32;2m交易成功|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")

				h.Count[0]++
			} else { //交易失败
				h.RedisClient.Int64(h.RedisClient.Execute("LPUSH", list_key+"OrderFail", order)) //加入异常列表

				fmt.Printf("\033[7;31;2m交易失败!|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")
				h.Count[1]++

				//！！！异常订单处理 在这里设置重发次数和时间间隔
				time.Sleep(time.Duration(h.ResendTime) * time.Duration(h.Count[1]) * time.Millisecond)
				fmt.Printf("第 %d 次尝试\n", h.Count[1])
				if h.Count[1] == h.ResendNum {
					fmt.Printf("\033[7;33;2m重试次数超出限制 %d 已终止\033[0m\n", h.Count[1])

					h.BadOrderProcess(payorder, pay_error)
					h.Count[1] = 0
				}
			}
			time.Sleep(10 * time.Millisecond)
			fmt.Printf("操作成功 %d 条订单\n", h.Count[0])
		default:
			res, err := redisClient.Bytes(redisClient.Execute("BRPOPLPUSH", list_key, list_key+"Backups", 10))
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
	redisClient.Int64(redisClient.Execute("LPUSH", h.List_key+"Bad", string(bad)))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key, 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"OrderFail", 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, h.Order))
	// redisClient.ReleaseOneRedisClient()
}

func (h *Smv1) FailOrderProcess() {
	redisClient := redis_factory.GetOneRedisClient()
	for {
		select {
		case <-h.Cope:
			fmt.Println("等待正常订单处理完成")
			continue
		default:
			res, err := redisClient.StringMap(redisClient.Execute("BRPOP", h.List_key+"OrderFail", 10)) //阻塞等待list
			if err != nil {
				fmt.Println("异常订单列表无数据", err)
				continue
			}
			fmt.Println("尝试处理异常订单...")
			_, err = redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, res[h.List_key+"OrderFail"]))
			if err != nil {
				fmt.Println("删除异常订单", err)
				continue
			}

			_, err = redisClient.Int64(redisClient.Execute("RPUSH", h.List_key, res[h.List_key+"OrderFail"]))
			if err != nil {
				fmt.Println("将异常订单重新加入列表", err)
				continue
			}
		}
	}
}
