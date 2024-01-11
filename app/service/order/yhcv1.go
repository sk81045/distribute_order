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
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Yhcv1 struct {
	List_key   string
	Cope       chan int
	FailCope   chan string
	Order      string
	ResendNum  int64
	ResendTime int64
	Count      []int64
	Soap       struct {
		Url string
	}
	RedisClient *redis_factory.RedisClient
}

func (manager *Yhcv1) Run(list_key string, secret string, order_config interface{}) {
	/**********************************任务初始化*************************/
	manager.List_key = list_key
	manager.Soap.Url = fmt.Sprintf("%s", order_config.(map[string]interface{})["sellsoap"])
	resendnum, _ := strconv.ParseInt(fmt.Sprintf("%d", order_config.(map[string]interface{})["resendnum"]), 10, 64)
	manager.ResendNum = resendnum
	resentime, _ := strconv.ParseInt(fmt.Sprintf("%d", order_config.(map[string]interface{})["resendtime"]), 10, 64)
	manager.ResendTime = resentime
	manager.Cope = make(chan int, 3)        //任务执行信号
	manager.Count = make([]int64, 3)        //任务计数器
	manager.FailCope = make(chan string, 1) //异常订单执行信号
	/*******************************************************************/

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
			//2. 处理订单
			paytime := time.Unix(payorder.Dealtime, 0)
			if pay_error == nil {
				_, err := redisClient.Int64(redisClient.Execute("LPUSH", "OrderDone", order))
				if err != nil {
					fmt.Println("无法将订单加入已完成列表:", order)
					continue
				}

				fmt.Printf("\033[7;32;2m交易成功|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")

				manager.Count[0]++
			} else {
				//交易失败
				manager.failOrder(list_key, order)
				//！！！异常订单处理 在这里设置重发次数和时间间隔
				fmt.Printf("\033[7;31;2m交易失败!|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
				fmt.Println("====================================================")
				fmt.Println("重发次数", payorder.Resend)
				if payorder.Resend >= manager.ResendNum {
					fmt.Printf("\033[7;33;2m重试次数超出限制 %d 已终止\033[0m\n", manager.ResendNum)
					manager.BadOrderProcess(payorder, pay_error)
					continue
				}
				time.Sleep(time.Duration(payorder.Resend*100) * time.Millisecond)
				manager.FailCope <- order
			}

			time.Sleep(time.Duration(manager.ResendTime) * time.Millisecond)
			fmt.Printf("操作成功 %d 条订单\n", manager.Count[0])
		default:
			manager.RedisClient = redis_factory.GetOneRedisClient()
			redisClient := manager.RedisClient
			res, err := redisClient.Bytes(redisClient.Execute("BRPOPLPUSH", list_key, list_key+"Backups", 30))
			if err != nil {
				log.Println("读取新订单列表数据", err)
				continue
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
	redisClient.Int64(redisClient.Execute("LPUSH", "OrderBad", string(bad)))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key, 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"OrderFail", 0, h.Order))
	redisClient.Int64(redisClient.Execute("LREM", h.List_key+"Backups", 0, h.Order))
	redisClient.ReleaseOneRedisClient()
}

func (h *Yhcv1) FailOrderProcess() {
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

func (h *Yhcv1) failOrder(list_key string, order string) {
	h.RedisClient.Int64(h.RedisClient.Execute("LPUSH", list_key+"OrderFail", order))
}

func (h *Yhcv1) Process(order model.Payorder) error {
	switch order.Type {
	case 1:
		return h.HttpOrder(order, "33") //增款
	case 2:
		return h.HttpOrder(order, "15") //减款
	default:
		return fmt.Errorf("接口错误 非预期返回值")
	}
}

func (h *Yhcv1) HttpOrder(order model.Payorder, kind string) error {
	dealtime := time.Unix(order.Dealtime, 0)
	fmt.Println("交易时间", dealtime.Format("2006-01-02 15:04:05"))
	money, _ := strconv.ParseFloat(order.Price, 64)
	Body := yhcv1.RechargeParams{ //交易记录报文
		AccountID: order.Studentid,
		CardID:    order.Ic,
		PayMoney:  money,
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
