package service

import (
	"fmt"
	"github.com/robfig/cron"
	"log"
	"time"
)

type DelyTask struct {
	Limit int
	Num   int
}

func (de *DelyTask) Run() {
	cronS := cron.New()
	spec := "0 0 04 * * ?" //执行定时任务（每5秒执行一次）
	err := cronS.AddFunc(spec, func() {
		log.Println("DelyTask for run")
		de.Limit = 10
		de.Num = 0
		de.BalanceAlert()
	})
	if err != nil {
		fmt.Println("ER", err)
	}
	cronS.Start()
	defer cronS.Stop()
	select {}
}

func (de *DelyTask) BalanceAlert() {
	time.Sleep(1 * time.Second)
	fmt.Println(" RecursionDeduct Run...")
	de.Num = de.Num + 1
	if de.Num <= de.Limit { //计算偏移liang
		de.BalanceAlert()
	} else {
		de.Num = 0
		fmt.Println("递归结束.......")
	}
}
