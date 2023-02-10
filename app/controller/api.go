package controller

import (
	"Hwgen/app/model"
	"Hwgen/global"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Api struct {
}

type RechargeApiParams struct { //通过接口充值时，需引用的结构体
	MerchantID     string  `json:"merchantID"`
	MemberID       string  `json:"memberID"`
	MemberNo       string  `json:"memberNo"`
	PayTime        string  `json:"payTime"`
	PayType        string  `json:"PayType"`
	Amount         float32 `json:"amount"`
	ReceiptsAmount float32 `json:"receiptsAmount"`
	Remarks        string  `json:"remarks"`
}

func (rs *Api) Rechage(value string) (res string) {
	system := global.H_CONFIG.System
	var order = model.Payorder{}
	// rs.Token()
	_ = json.Unmarshal([]byte(value), &order)

	t := time.Unix(int64(order.Created_at), 0)
	DateStr := t.Format("20060102150405")
	requestBody := RechargeApiParams{ //充值报文
		MerchantID:     "admin",
		MemberID:       "",
		MemberNo:       "xsdssxdsda",
		PayTime:        DateStr,
		PayType:        "2",
		Amount:         order.Price,
		ReceiptsAmount: order.Price,
		Remarks:        "农行线上充值",
	}
	reader_str, _ := json.Marshal(&requestBody)
	reader := bytes.NewReader(reader_str)
	URL := system.SellfoodApiurl + "/OtherPlatformsRecharge?AuthkeyType=1&Authkey=" + system.SellfoodAppid + "|" + SellFoodSystem_Token.Token
	request, _ := http.NewRequest("POST", URL, reader)
	request.Header.Set("Content-Type", "application/json")
	response, _ := http.DefaultClient.Do(request)
	defer func() {
		response.Body.Close()
	}()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("充值接口错误")
	}
	fmt.Println("requestBody", string(body))

	return "ok"
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

func (rs *Api) Token() { //获取Token
	system := global.H_CONFIG.System
	URL := system.SellfoodApiurl + "/GetToken?Appid=" + system.SellfoodAppid + "&Secretkey=" + system.SellfoodSecretkey
	fmt.Println("URL", URL)

	request, _ := http.NewRequest("GET", URL, nil)
	response, _ := http.DefaultClient.Do(request)
	defer func() {
		response.Body.Close()
	}()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("充值接口错误")
	}

	err = json.Unmarshal(body, &SellFoodSystem_Token)
	fmt.Println("SellFoodSystem_Token", SellFoodSystem_Token)

	if err != nil {
		fmt.Println(err, "获取token失败,非预期返回值")
	}
}
