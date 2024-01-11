package delyTask

import (
	// "encoding/json"
	"fmt"
	"github.com/robfig/cron"
	// "goskeleton/app/model/yhcv1"
	// "goskeleton/app/model/yhcv2"
	"goskeleton/app/model/proxydeal"
	"goskeleton/app/utils/http"
	"log"
	"time"
)

type TimedMsg struct {
	Sid   string
	Limit int64
	Num   int64
	Count int64
}

func (de *TimedMsg) Run(exe_time string, sid string) {
	de.Sid = sid
	fmt.Println("公共消息推送", exe_time)
	cronS := cron.New()
	spec := "0 0 05 * * ?" //执行定时任务（每5秒执行一次）
	// spec := "*/10 * * * * *"
	log.Println("TimedMsg for run", exe_time)

	err := cronS.AddFunc(spec, func() {
		de.Limit = 10
		de.Num = 0
		de.Count = (&proxydeal.ProxyDeal{}).MemberNums("cash < 1")
		// de.Count = yhcv2.MermberFactory("").MemberNums("Balance < 50")
		de.BalanceAlert()
	})

	de.Limit = 10
	de.Num = 0
	// de.Count = yhcv2.MermberFactory("").MemberNums("Balance < 50")
	de.Count = (&proxydeal.ProxyDeal{}).MemberNums("cash < 1")
	de.BalanceAlert()
	if err != nil {
		fmt.Println("ER", err)
	}
	cronS.Start()
	defer cronS.Stop()
	select {}
}

type requestBody struct {
	Type string      `json:"type"`
	Sid  string      `json:"sid"`
	Data interface{} `json:"data"`
	// Data []yhcv2.MemberInfo `json:"data"`
}

func (de *TimedMsg) BalanceAlert() {
	// time.Sleep(5000 * time.Millisecond)
	de.Num = de.Num + 1
	fmt.Println("de.Num", de.Num)

	// list, err := yhcv2.MermberFactory("").GetMembers(de.Num, de.Limit, "Balance < 50")
	list, err := (&proxydeal.ProxyDeal{}).GetMembers(de.Num, de.Limit, "cash < 1")
	if err != nil {
		fmt.Println("list is not")
		return
	}
	fmt.Println("list", list)

	requestBody := requestBody{
		Type: "balance",
		Sid:  de.Sid,
		Data: list,
	}
	URL := "http://xb36.yoozhi.cn/api/outside/pubmsg/main"
	res, err := http.HttpPost(URL, requestBody, "application/json")
	fmt.Println(string(res))
	fmt.Println("最大次数", de.Count/de.Limit)

	if de.Num < de.Count/de.Limit { //计算偏移liang
		time.Sleep(10000 * time.Millisecond)
		de.BalanceAlert()
	} else {
		de.Num = 0
		fmt.Println("递归结束.......")
	}
}
