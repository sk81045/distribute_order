package order

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"goskeleton/app/model"
	"goskeleton/app/model/yhcv1"
	"goskeleton/app/utils/redis_factory"
	"goskeleton/app/utils/sign"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Yhcv1 struct {
	List_key string
	Cope     chan int
	Order    string
	Count    []int64
	Soap     struct {
		Url string
	}
	RedisClient *redis_factory.RedisClient
}

func (manager *Yhcv1) Run(list_key string, secret string, order interface{}) {
	/**********************************任务初始化*************************/
	manager.List_key = list_key
	manager.Soap.Url = fmt.Sprintf("%s", order.(map[string]interface{})["sellsoap"])
	manager.Cope = make(chan int, 3) //任务执行信号
	manager.Count = make([]int64, 3) //任务计数器

	manager.RedisClient = redis_factory.GetOneRedisClient()
	/*******************************************************************/

	go manager.FailOrderProcess()

	redisClient := manager.RedisClient
	for {
		select {
		case <-manager.Cope:
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
			pay_error := manager.Process(payorder)
			//2. 处理订单
			paytime := time.Unix(payorder.Dealtime, 0)
			if pay_error == nil {
				_, err := redisClient.Int64(redisClient.Execute("LPUSH", "OrderDone", order))
				if err != nil {
					fmt.Println("无法将订单加入已完成列表:", order)
					continue
				}

				fmt.Printf("\033[7;32;2m交易成功|操作%d|$%f|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")

				manager.Count[0]++
			} else {
				//交易失败
				manager.failOrder(list_key, order)
				fmt.Printf("\033[7;31;2m交易失败!|操作%d|$%f|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")
				manager.Count[1]++
				//！！！异常订单处理 在这里设置重发次数和时间间隔
				time.Sleep((10 * time.Duration(manager.Count[1])) * time.Millisecond)
				if manager.Count[1] == 10 {
					fmt.Printf("重试次数超出限制机 %d 加入错误订单\n", manager.Count[1])
					manager.BadOrderProcess(payorder, pay_error)
					manager.Count[1] = 0
				}
			}

			time.Sleep(1 * time.Millisecond)
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

func (h *Yhcv1) BadOrderProcess(bad_order model.Payorder, err_msg error) {
	redisClient := h.RedisClient
	bad_order.Error = err_msg.Error()
	bad, _ := json.Marshal(bad_order)
	redisClient.Int64(redisClient.Execute("LPUSH", h.List_key+"Bad", string(bad)))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key, 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"OrderFail", 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, h.Order))
	// redisClient.ReleaseOneRedisClient()
}

func (h *Yhcv1) FailOrderProcess() {
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

func (h *Yhcv1) failOrder(list_key string, order string) {
	h.RedisClient.Int64(h.RedisClient.Execute("LPUSH", list_key+"OrderFail", order))
}

func (h *Yhcv1) Process(order model.Payorder) error {
	switch order.Type {
	case 1:
		return h.HttpOrder(order, "33")
	case 2:
		return h.HttpOrder(order, "15")
	default:
		return fmt.Errorf("接口错误 非预期返回值")
	}
}

func (h *Yhcv1) HttpOrder(order model.Payorder, kind string) error {
	dealtime := time.Unix(order.Dealtime, 0)
	Body := yhcv1.RechargeParams{ //交易记录报文
		AccountID: order.Studentid,
		CardID:    order.Ic,
		PayMoney:  order.Price,
		PayTime:   dealtime.Format("2006-01-02 15:04:05"),
		MacID:     order.Macid,
		MacType:   "app",
		PayKind:   kind,
		OrderNO:   order.Orderid,
	}
	jsonByte, _ := json.Marshal(Body)
	Soaps := string(jsonByte)
	reqBody := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"  xmlns:xsd="http://www.w3.org/2001/XMLSchema"  xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
<soap:Body>
<PayListAdd xmlns="http://localhost:8081/soap/IUCWebService">
<val>` + Soaps + `</val>
</PayListAdd>
</soap:Body>
</soap:Envelope>`

	res, err := http.Post(h.Soap.Url, "text/xml; charset=UTF-8", strings.NewReader(reqBody))
	if nil != err {
		fmt.Println("http post err:", err)
	}
	defer res.Body.Close()

	if 200 != res.StatusCode {
		fmt.Println("WebService soap1.1 request fail, status: %s\n", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)
	if nil != err {
		fmt.Println("ioutil ReadAll err:", err)
	}

	req := &MyRespEnvelope{}
	_ = xml.Unmarshal(data, req)

	type remsg struct {
		ErrorMsg string `json:"errorMsg"`
		ErrorID  int    `json:"errorID"`
	}

	var ee remsg
	err = json.Unmarshal([]byte(req.Body.GetResponse.MyVar), &ee)
	if nil != err {
		fmt.Println("json.Unmarshal err:", err)
	}

	items, _ := json.Marshal(ee)
	fmt.Println("response", string(items))
	if ee.ErrorID == 0 && ee.ErrorMsg == "交易成功" {
		return nil
	}
	return fmt.Errorf(ee.ErrorMsg)
}

// ============解析XML start=====================
type MyRespEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    Body
}

type Body struct {
	XMLName     xml.Name
	GetResponse completeResponse `xml:"PayListAddResponse"`
}

type completeResponse struct {
	XMLName xml.Name `xml:"PayListAddResponse"`
	MyVar   string   `xml:"return"`
}

func (h *Yhcv1) Buyget(order model.Payorder) error {

	return nil

}

// BLPOP list1 100
func (h *Yhcv1) BLPOP(list_key string) []string {
	res, err := h.RedisClient.Strings(h.RedisClient.Execute("BLPOP", list_key, 3))
	if err != nil {
		fmt.Println("err", err)
	}
	return res
}

func (h *Yhcv1) LLEN(list_key string) int64 {
	res, err := h.RedisClient.Int64(h.RedisClient.Execute("LLEN", list_key))
	if err != nil {
		fmt.Println("err", err)
	}
	return res
}

// LLEN KEY_NAME
func (h *Yhcv1) Lrange(list_key string, START int64, END int64) []string {
	res, err := h.RedisClient.Strings(h.RedisClient.Execute("LRANGE", list_key, START, END))
	if err != nil {
		fmt.Println("Lrange err", err)
	}
	h.RedisClient.ReleaseOneRedisClient()
	return res
}
