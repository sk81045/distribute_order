package order

import (
	"encoding/json"
	"fmt"
	"goskeleton/app/model"
	"goskeleton/app/utils/http"
	"goskeleton/app/utils/redis_factory"
	"goskeleton/app/utils/sign"
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
	List_key    string
	Secret      string
	Cope        chan int
	Cope2       chan int
	OriginOrder []string
	Order       string
	Count       []int64
	Config
	TokenInfo
}

var (
	RedisClient = redis_factory.GetOneRedisClient()
)

func (manager *Yhcv2) Run(list_key string, secret string, order interface{}) {
	/**********************************任务初始化*************************/
	manager.List_key = list_key
	manager.Secret = secret
	manager.Config.Url = fmt.Sprintf("%s", order.(map[string]interface{})["sellapiurl"])
	manager.Config.Appid = fmt.Sprintf("%s", order.(map[string]interface{})["sellappid"])
	manager.Config.Secret = fmt.Sprintf("%s", order.(map[string]interface{})["sellsecret"])

	manager.Cope = make(chan int, 3)  //任务执行信号
	manager.Cope2 = make(chan int, 3) //任务执行信号
	manager.Count = make([]int64, 3)  //任务计数器
	/*******************************************************************/
	go manager.FailOrderProcess()
	for {
		select {
		case <-manager.Cope:
			redisClient := redis_factory.GetOneRedisClient()

			payorder := model.Payorder{}
			if err := json.Unmarshal([]byte(manager.Order), &payorder); err != nil {
				continue
			}

			if !sign.Verify(manager.Order, secret) {
				manager.BadOrderProcess(payorder, fmt.Errorf("Sign验证失败!"))
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			order := manager.Order
			msg := "$"
			pay_error := manager.Process(payorder)
			//2. 充值
			if pay_error == nil {
				_, err := redisClient.Int64(redisClient.Execute("LPUSH", "OrderDone", order))
				if err != nil {
					fmt.Println("无法将订单加入已完成列表:", order)
					continue
				}
				time := time.Unix(payorder.Created_at, 0)
				fmt.Printf("交易成功|%s%f|订单号:%s|用户编号:%d|日期:%s \n", msg, payorder.Price, payorder.Orderid, payorder.Studentid, time.Format("2006-01-02 15:04:05"))
				fmt.Println("======================end==============================")
				manager.Count[0]++
			} else {
				//交易失败
				manager.failOrder(list_key, order)
				fmt.Println("=====================start=============================")
				fmt.Printf("交易失败!|%s%f|订单号:%s|用户编号:%d \n", msg, payorder.Price, payorder.Orderid, payorder.Studentid)
				fmt.Println("======================end==============================")
				manager.Count[1]++
				//！！！异常订单处理 在这里设置重发次数和时间间隔
				time.Sleep((1000 * time.Duration(manager.Count[1])) * time.Millisecond)
				if manager.Count[1] == 10 {
					fmt.Printf("重试次数超出限制机 %d 加入错误订单\n", manager.Count[1])
					manager.BadOrderProcess(payorder, pay_error)
					manager.Count[1] = 0
				}
			}

			time.Sleep(10 * time.Millisecond)
			fmt.Printf("操作成功 %d 条订单\n", manager.Count[0])

		default:
			redisClient := redis_factory.GetOneRedisClient()
			res, err := redisClient.Bytes(redisClient.Execute("BRPOPLPUSH", list_key, list_key+"Backups", 0))
			if err != nil {
				fmt.Println("err", err)
			}
			manager.Order = string(res)
			manager.Cope <- 1 //处理订单
		}
	}
}

func (h *Yhcv2) BadOrderProcess(bad_order model.Payorder, err_msg error) {
	bad_order.Error = err_msg.Error()
	bad, _ := json.Marshal(bad_order)
	RedisClient.Int64(RedisClient.Execute("LPUSH", h.List_key+"Bad", string(bad)))
	RedisClient.Int64(RedisClient.Execute("LREM", h.List_key, 0, h.Order))
	RedisClient.Int64(RedisClient.Execute("LREM", h.List_key+"OrderFail", 0, h.Order))
	RedisClient.Int64(RedisClient.Execute("LREM", h.List_key+"Backups", 0, h.Order))
}

func (h *Yhcv2) FailOrderProcess() {
	redisClient := redis_factory.GetOneRedisClient()
	for {
		select {
		case <-h.Cope:
			fmt.Println("等待订单")
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

func (h *Yhcv2) failOrder(list_key string, order string) {
	redisClient := redis_factory.GetOneRedisClient()
	redisClient.Int64(redisClient.Execute("LPUSH", list_key+"OrderFail", order))
}

type OrderParams struct {
	MerchantID      string  `json:"MerchantID"`
	MemberID        string  `json:"MemberID"`
	MemberNo        int     `json:"MemberNo"`
	PayTime         string  `json:"payTime"`
	ConsumptionTime string  `json:"ConsumptionTime"`
	ReceiptsAmount  float32 `json:"receiptsAmount"`
	TerminalNo      string  `json:"TerminalNo"`
	Amount          float32 `json:"Amount"`
	SubsidiesAmount float32 `json:"SubsidiesAmount"`
	GiftAmount      float32 `json:"GiftAmount"`
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
		return fmt.Errorf("接口错误 非预期返回值")
	}
}

func (h *Yhcv2) Recharge(order model.Payorder) error {
	dealtime := time.Unix(order.Dealtime, 0)
	requestBody := OrderParams{ //充值报文
		MerchantID:     "admin",
		MemberID:       "",
		MemberNo:       order.Studentid,
		PayTime:        dealtime.Format("20060102150405"),
		PayType:        "2",
		Amount:         order.Price,
		ReceiptsAmount: order.Price,
		Remarks:        order.From + "|" + order.Orderid,
	}
	var token, _ = h.Token()
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
		fmt.Printf("\033[32;4mid %d:%s-> +%f\033[0m\n", order.Studentid, ReqMsg.Message, order.Price)
		return nil
	} else {
		fmt.Printf("\033[7;31;40mid %d:%s\033[0m\n", order.Studentid, ReqMsg.Message)
		return fmt.Errorf(ReqMsg.Message)
	}

}

func (h *Yhcv2) Buyget(order model.Payorder) error {
	dealtime := time.Unix(order.Dealtime, 0)
	requestBody := OrderParams{
		MerchantID:      "admin",
		MemberID:        "",
		MemberNo:        order.Studentid,
		ConsumptionTime: dealtime.Format("20060102150405"),
		TerminalNo:      order.Macid,
		Amount:          order.Price,
		SubsidiesAmount: 0,
		GiftAmount:      0,
		Remarks:         order.From + "|" + order.Orderid,
		AuthkeyType:     1,
	}

	var token, _ = h.Token()
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
		fmt.Printf("\033[32;4mid %d:%s-> -%f\033[0m\n", order.Studentid, ReqMsg.Message, order.Price)
		return nil
	} else {
		fmt.Printf("\033[7;31;40mid %d:%s\033[0m\n", order.Studentid, ReqMsg.Message)
		return fmt.Errorf(ReqMsg.Message)
	}

}

func (h *Yhcv2) Token() (string, error) { //获取Token
	timeout, _ := time.ParseInLocation("2006-01-02 15:04:05", (*h).TokenInfo.ValidityDate, time.Local)
	if time.Now().Unix() < timeout.Unix() {
		fmt.Println("Token 未超时")
		return (*h).TokenInfo.Token, nil
	}

	var token = &(*h).TokenInfo
	body, err := http.HttpGet((*h).Config.Url + "/GetToken?Appid=" + (*h).Config.Appid + "&Secretkey=" + (*h).Config.Secret)
	if nil != err {
		fmt.Println("Http get request err:", err)
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
