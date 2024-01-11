package order

import (
	"encoding/json"
	"fmt"
	"goskeleton/app/model"
	"goskeleton/app/utils/http"
	"goskeleton/app/utils/redis_factory"
	"goskeleton/app/utils/sign"
	"log"
	"strconv"
	"time"
)

type Config struct {
	Url    string
	Appid  string
	Secret string
}

type TokenInfo struct {
	Token        string
	Status       int
	Success      bool
	Message      string
	ExecTime     int
	Authkey      string
	ValidityDate string
}

type Yhcv2 struct {
	List_key string
	Secret   string
	Cope     chan int
	// FailCope   chan *model.Payorder
	FailCope   chan string
	Order      string
	Count      []int64
	ResendNum  int64
	ResendTime int64
	Config
	TokenInfo
	RedisClient *redis_factory.RedisClient
}

func (manager *Yhcv2) Run(list_key string, secret string, order_config interface{}) {
	/**********************************任务初始化*************************/
	manager.List_key = list_key
	manager.Secret = secret
	manager.Config.Url = fmt.Sprintf("%s", order_config.(map[string]interface{})["sellapiurl"])
	manager.Config.Appid = fmt.Sprintf("%s", order_config.(map[string]interface{})["sellappid"])
	manager.Config.Secret = fmt.Sprintf("%s", order_config.(map[string]interface{})["sellsecret"])

	resendnum, _ := strconv.ParseInt(fmt.Sprintf("%d", order_config.(map[string]interface{})["resendnum"]), 10, 64)
	manager.ResendNum = resendnum
	resentime, _ := strconv.ParseInt(fmt.Sprintf("%d", order_config.(map[string]interface{})["resendtime"]), 10, 64)
	manager.ResendTime = resentime

	manager.Cope = make(chan int, 1) //任务执行信号
	// manager.FailCope = make(chan *model.Payorder, 1)
	manager.FailCope = make(chan string, 1) //异常订单执行信号

	manager.Count = make([]int64, 3) //任务计数器
	// manager.RedisClient = redis_factory.GetOneRedisClient()
	/*******************************************************************/
	// redisClient := manager.RedisClient
	// redisClient := redis_factory.GetOneRedisClient()
	go manager.FailOrderProcess()
	for {
		select {
		case <-manager.Cope:
			redisClient := manager.RedisClient
			payorder := model.Payorder{}
			if err := json.Unmarshal([]byte(manager.Order), &payorder); err != nil {
				fmt.Println("序列化失败", err)
				manager.BadOrderProcess(payorder, err)
				continue
			}

			if payorder.Mark != "RESEND" { //重发被标记的订单不做sign验证
				if err := sign.Verify(manager.Order, secret); err != nil {
					fmt.Println("Sign验证失败")
					manager.BadOrderProcess(payorder, err)
					continue
				}
			} else {
				fmt.Println("重发被标记的订单不做sign验证")
			}

			order := manager.Order
			pay_error := manager.Process(payorder)
			paytime := time.Unix(payorder.Dealtime, 0)
			//2. 充值
			if pay_error == nil {
				_, err := redisClient.Int64(redisClient.Execute("LPUSH", "OrderDone", order))
				if err != nil {
					fmt.Println("无法将订单加入已完成列表:", err)
					continue
				}
				fmt.Printf("\033[7;32;2m交易成功|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")
				manager.Count[0]++
			} else {
				//交易失败
				_, err := redisClient.Int64(redisClient.Execute("LPUSH", manager.List_key+"OrderFail", order))
				if err != nil {
					fmt.Println("PUSH IN FAILLIST:", err)
					manager.BadOrderProcess(payorder, err)
					continue
				}

				fmt.Printf("\033[7;31;2m交易失败!|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")

				fmt.Println("重发次数", payorder.Resend)
				//！！！异常订单处理 在这里设置重发次数和时间间隔
				if payorder.Resend >= manager.ResendNum {
					fmt.Printf("\033[7;33;2m重试次数超出限制 %d 已终止\033[0m\n", manager.ResendNum)
					manager.BadOrderProcess(payorder, pay_error)
					continue
				}
				redisClient.ReleaseOneRedisClient()
				time.Sleep(time.Duration(payorder.Resend*100) * time.Millisecond)
				manager.FailCope <- order
			}

			redisClient.ReleaseOneRedisClient()
			time.Sleep(time.Duration(manager.ResendTime) * time.Millisecond) //给接口的反应的时间提高成功率
			fmt.Printf("操作成功 %d 条订单\n", manager.Count[0])
		default:
			manager.RedisClient = redis_factory.GetOneRedisClient()
			redisClient := manager.RedisClient
			res, err := redisClient.Bytes(redisClient.Execute("BRPOPLPUSH", list_key, list_key+"Backups", 30))
			if err != nil {
				log.Println("读取新订单列表数据", err)
				redisClient.ReleaseOneRedisClient()
				continue
			}
			manager.Order = string(res)
			manager.Cope <- 1 //处理订单
		}
	}
}

func (h *Yhcv2) BadOrderProcess(bad_order model.Payorder, err_msg error) {
	redisClient := h.RedisClient
	bad_order.Error = err_msg.Error()
	bad, _ := json.Marshal(bad_order)
	redisClient.Int64(redisClient.Execute("LPUSH", "OrderBad", string(bad)))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key, 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"OrderFail", 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, h.Order))
	redisClient.ReleaseOneRedisClient()
}

func (h *Yhcv2) FailOrderProcess() {
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

type OrderParams struct {
	MerchantID      string  `json:"MerchantID"`
	MemberID        string  `json:"MemberID"`
	MemberNo        int     `json:"MemberNo"`
	PayTime         string  `json:"payTime"`
	ConsumptionTime string  `json:"ConsumptionTime"`
	ReceiptsAmount  float64 `json:"receiptsAmount"`
	TerminalNo      string  `json:"TerminalNo"`
	Amount          float64 `json:"Amount"`
	SubsidiesAmount float64 `json:"SubsidiesAmount"`
	GiftAmount      float64 `json:"GiftAmount"`
	Times           int     `json:"Times"`
	Remarks         string  `json:"Remarks"`
	AuthkeyType     int     `json:"AuthkeyType"`
	Authkey         string  `json:"Authkey"`
	PayType         string  `json:"PayType"`
}

type RechargeRes struct {
	Status   int
	Success  bool
	Message  string
	ExecTime int
	Authkey  string
}

func (h *Yhcv2) Process(order model.Payorder) error {
	switch order.Type {
	case 1:
		return h.Recharge(order)
	case 2:
		return h.Buyget(order)
	default:
		return fmt.Errorf("错误 未识别业务类型")
	}
}

func (h *Yhcv2) Recharge(order model.Payorder) error {
	dealtime := time.Unix(order.Dealtime, 0)
	Amount, _ := strconv.ParseFloat(order.Price, 64)
	requestBody := OrderParams{ //充值报文
		MerchantID:     "admin",
		MemberID:       "",
		MemberNo:       order.Studentid,
		PayTime:        dealtime.Format("20060102150405"),
		PayType:        "2",
		Amount:         Amount,
		ReceiptsAmount: Amount,
		Remarks:        order.Orderid,
	}
	var token, err = h.Token()
	if err != nil {
		return err
	}
	URL := (*h).Config.Url + "/OtherPlatformsRecharge?AuthkeyType=1&Authkey=" + (*h).Config.Appid + "|" + token
	res, err := http.HttpPost(URL, requestBody, "application/json")
	if nil != err {
		fmt.Println("ioutil ReadAll err:", err)
	}
	var ReqMsg RechargeRes
	err = json.Unmarshal(res, &ReqMsg)
	if err != nil {
		return fmt.Errorf("接口错误 非预期返回值")
	}

	if ReqMsg.Success == true {
		return nil
	} else {
		return fmt.Errorf(ReqMsg.Message)
	}

}

func (h *Yhcv2) Buyget(order model.Payorder) error {
	dealtime := time.Unix(order.Dealtime, 0)
	Amount, _ := strconv.ParseFloat(order.Price, 64)
	requestBody := OrderParams{
		MerchantID:      "admin",
		MemberID:        "",
		MemberNo:        order.Studentid,
		ConsumptionTime: dealtime.Format("20060102150405"),
		TerminalNo:      order.Macid,
		Amount:          Amount,
		SubsidiesAmount: 0,
		GiftAmount:      0,
		Remarks:         order.Orderid,
		AuthkeyType:     1,
	}

	var token, err = h.Token()
	if err != nil {
		return err
	}
	URL := (*h).Config.Url + "/OtherPlatformsConsumption?AuthkeyType=1&Authkey=" + (*h).Config.Appid + "|" + token
	res, err := http.HttpPost(URL, requestBody, "application/json")
	if nil != err {
		fmt.Println("ioutil ReadAll err:", err)
	}
	var ReqMsg RechargeRes
	err = json.Unmarshal(res, &ReqMsg)
	if err != nil {
		return fmt.Errorf("接口错误 非预期返回值")
	}

	if ReqMsg.Success == true {
		return nil
	} else {
		return fmt.Errorf(ReqMsg.Message)
	}

}

func (h *Yhcv2) Token() (string, error) { //获取Token
	timeout, _ := time.ParseInLocation("2006-01-02 15:04:05", (*h).TokenInfo.ValidityDate, time.Local)
	if time.Now().Unix() < timeout.Unix() {
		fmt.Println("Token", (*h).TokenInfo.Token)
		return (*h).TokenInfo.Token, nil
	}

	var token = &(*h).TokenInfo
	body, err := http.HttpGet((*h).Config.Url + "/GetToken?Appid=" + (*h).Config.Appid + "&Secretkey=" + (*h).Config.Secret)
	// http.TestGet()
	if nil != err {
		return "", fmt.Errorf("网络原因 获取token失败")
	}

	err = json.Unmarshal([]byte(body), &token)
	if err != nil {
		return "", err
	}
	if !token.Success {
		return "", fmt.Errorf("获取token失败")
	}

	return token.Token, nil
}

// BLPOP list1 100
func (h *Yhcv2) BLPOP(list_key string) []string {
	redisClient := redis_factory.GetOneRedisClient()
	res, err := redisClient.Strings(redisClient.Execute("BLPOP", list_key, 3))
	if err != nil {
		fmt.Println("err", err)
	}
	return res
}

func (h *Yhcv2) LLEN(list_key string) int64 {
	redisClient := redis_factory.GetOneRedisClient()
	res, err := redisClient.Int64(redisClient.Execute("LLEN", list_key))
	if err != nil {
		fmt.Println("err", err)
	}
	return res
}

// LLEN KEY_NAME
func (h *Yhcv2) Lrange(list_key string, START int64, END int64) []string {
	redisClient := redis_factory.GetOneRedisClient()
	res, err := redisClient.Strings(redisClient.Execute("LRANGE", list_key, START, END))
	if err != nil {
		fmt.Println("Lrange err", err)
	}
	redisClient.ReleaseOneRedisClient()
	return res
}
