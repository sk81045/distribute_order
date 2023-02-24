package controller

import (
	"Hwgen/app/model"
	"Hwgen/global"
	"Hwgen/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Api struct {
}

func (rs *Api) Mission(value string) (ok bool) {
	var order = model.Payorder{}
	_ = json.Unmarshal([]byte(value), &order)
	switch order.Type {
	case 1:
		ok = rs.Rechage(order)
	case 2:
		ok = rs.Deduction(order)
	default:
		fmt.Println("未识别业务类型->", order.Type)
	}
	return ok
}

type RechargeApiParams struct { //通过接口充值时，需引用的结构体
	MerchantID     string  `json:"merchantID"`
	MemberID       string  `json:"memberID"`
	MemberNo       int     `json:"memberNo"`
	PayTime        string  `json:"payTime"`
	PayType        string  `json:"PayType"`
	Amount         float32 `json:"amount"`
	ReceiptsAmount float32 `json:"receiptsAmount"`
	Remarks        string  `json:"remarks"`
}

type RechargeRes struct {
	Status   int
	Success  bool
	Message  string
	ExecTime int
	Authkey  string
}

func (rs *Api) Rechage(order model.Payorder) (ok bool) { //充值
	system := global.H_CONFIG.System
	t := time.Unix(time.Now().Unix(), 0)
	requestBody := RechargeApiParams{ //充值报文
		MerchantID:     "admin",
		MemberID:       "",
		MemberNo:       order.Students.Studentid,
		PayTime:        t.Format("20060102150405"),
		PayType:        "2",
		Amount:         order.Price,
		ReceiptsAmount: order.Price,
		Remarks:        "农行线上充值",
	}
	URL := system.SellfoodApiurl + "/OtherPlatformsRecharge?AuthkeyType=1&Authkey=" + system.SellfoodAppid + "|" + rs.Token()
	body := utils.HttpPost(URL, requestBody, "application/json")

	var ReqMsg RechargeRes
	err := json.Unmarshal(body, &ReqMsg)
	if err != nil {
		fmt.Println(err, "充值失败 充值接口错误 非预期返回值")
	}

	if ReqMsg.Success == true {
		fmt.Printf("\033[32;4mid %d:%s-> +%f\033[0m\n", order.Students.Studentid, ReqMsg.Message, order.Price)
		return true
	} else {
		fmt.Printf("\033[7;31;40mid %d:%s\033[0m\n", order.Students.Studentid, ReqMsg.Message)
		return false
	}
}

func (rs *Api) Deduction(order model.Payorder) bool { //扣费
	system := global.H_CONFIG.System
	t := time.Unix(time.Now().Unix(), 0)
	requestBody := model.DeductionParams{
		MerchantID:      "admin",
		MemberID:        "",
		MemberNo:        order.Students.Studentid,
		ConsumptionTime: t.Format("20060102150405"),
		TerminalNo:      "25",
		Amount:          order.Price,
		SubsidiesAmount: 0,
		GiftAmount:      0,
		Remarks:         order.From,
		AuthkeyType:     1,
	}

	URL := system.SellfoodApiurl + "/OtherPlatformsConsumption?AuthkeyType=1&Authkey=" + system.SellfoodAppid + "|" + rs.Token()
	res := utils.HttpPost(URL, requestBody, "application/json")

	var ReqMsg RechargeRes
	err := json.Unmarshal(res, &ReqMsg)
	if err != nil {
		fmt.Println(err, "扣费失败 扣费接口错误 非预期返回值")
	}

	if ReqMsg.Success == true {
		fmt.Printf("\033[32;4mid %d:%s-> -%f\033[0m\n", order.Students.Studentid, ReqMsg.Message, order.Price)
		return true
	} else {
		fmt.Printf("\033[7;31;40mid %d:%s\033[0m\n", order.Students.Studentid, ReqMsg.Message)
		return false
	}
}

type SellFoodSystemTokenReq struct {
	Token        string
	Status       int
	Success      bool
	Message      string
	ExecTime     int
	Authkey      string
	ValidityDate string
}

var (
	SellFoodSystem_Token SellFoodSystemTokenReq
)

func (rs *Api) Token() (token string) { //获取Token
	system := global.H_CONFIG.System
	URL := system.SellfoodApiurl + "/GetToken?Appid=" + system.SellfoodAppid + "&Secretkey=" + system.SellfoodSecretkey

	timeout, _ := time.ParseInLocation("2006-01-02 15:04:05", SellFoodSystem_Token.ValidityDate, time.Local)
	if time.Now().Unix() < timeout.Unix() {
		// fmt.Println("//Token not timeout: ", SellFoodSystem_Token)
		return SellFoodSystem_Token.Token
	}

	request, _ := http.NewRequest("GET", URL, nil)
	response, _ := http.DefaultClient.Do(request)
	defer func() {
		response.Body.Close()
	}()
	body, _ := ioutil.ReadAll(response.Body)
	err := json.Unmarshal(body, &SellFoodSystem_Token)
	if err != nil {
		fmt.Println(err, "获取token失败,非预期返回值")
	}

	if SellFoodSystem_Token.Success == true {
		return SellFoodSystem_Token.Token
	} else {
		fmt.Printf("\033[7;31;40m%s\033[0m\n", "获取token失败")
		return "err"
	}
}
