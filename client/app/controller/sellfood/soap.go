package controller

import (
	"Hwgen/app/model"
	"Hwgen/global"
	// "Hwgen/utils"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Soap struct {
}

type SoapRechargeParams struct { //通过webserver接口充值时，需引用的结构体
	AccountID int     `json:"AccountID"`
	CardID    string  `json:"CardID"`
	PayMoney  float32 `json:"PayMoney"`
	PayTime   string  `json:"PayTime"`
	MacID     string  `json:"MacID"`
	MacType   string  `json:"MacType"`
	PayKind   string  `json:"PayKind"`
	OrderNO   string  `json:"OrderNO"`
}

//============解析XML start=====================
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

func (sp *Soap) Mission(value string) (ok bool) {
	var order = model.Payorder{}
	_ = json.Unmarshal([]byte(value), &order)
	switch order.Type {
	case 1:
		ok = sp.AddRecord(order, "33") //增款
	case 2:
		ok = sp.AddRecord(order, "15") //减款
	default:
		fmt.Println("未识别业务类型->", order.Type)
	}
	return ok
}

func (sp *Soap) AddRecord(order model.Payorder, mhd string) bool { //二十三、添加交易记录接口
	MacId := global.H_CONFIG.System.MacId
	Body := SoapRechargeParams{ //充值报文
		AccountID: order.Studentid,
		CardID:    order.Ic,
		PayMoney:  order.Price,
		PayTime:   time.Unix(order.Created_at, 0).Format("2006-01-02 15:04:05"),
		MacID:     MacId,
		MacType:   "app",
		PayKind:   mhd,
		OrderNO:   order.Orderid,
	}
	jsonByte, _ := json.Marshal(Body)
	Soaps := string(jsonByte)
	fmt.Println("Soaps", Soaps)
	reqBody := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"  xmlns:xsd="http://www.w3.org/2001/XMLSchema"  xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
<soap:Body>
<PayListAdd xmlns="http://localhost:8081/soap/IUCWebService">
<val>` + Soaps + `</val>
</PayListAdd>
</soap:Body>
</soap:Envelope>`
	URL := global.H_CONFIG.System.SellfoodSoap
	res, err := http.Post(URL, "text/xml; charset=UTF-8", strings.NewReader(reqBody))
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
		return true
	} else {
		return false
	}
}
