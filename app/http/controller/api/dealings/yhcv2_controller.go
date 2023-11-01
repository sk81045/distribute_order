package dealings

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goskeleton/app/global/consts"
	"goskeleton/app/global/variable"
	"goskeleton/app/model"
	"goskeleton/app/model/yhcv2"
	"goskeleton/app/utils/redis_factory"
	"goskeleton/app/utils/response"
	"goskeleton/app/utils/sign"
	"math/rand"
	"time"
)

type Yhcv2 struct {
}

// 消费记录
func (y *Yhcv2) RecordList(context *gin.Context) {

	type Params struct { //类型绑定
		Pid   string `form:"pid"`
		Stime string `form:"stime"`
		Etime string `form:"etime"`
	}
	var p Params
	if context.ShouldBindQuery(&p) != nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Printf("====== yhc查询参数错误 pid:%s stime:%s etime:%s=====\n", p.Pid, p.Stime, p.Etime)
	}

	user, err := yhcv2.MermberFactory("").GetMemberInfo(p.Pid)
	if err != nil {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, err.Error())
		return
	}
	fmt.Println("user", user)

	mealList := yhcv2.RecordInfoFactory("").List(user.ID, p.Stime, p.Etime)
	chargeList := yhcv2.RechargeInfoFactory("").List(user.ID, p.Stime, p.Etime)

	temp1 := make([]model.DealRecord, len(mealList))
	for k, value := range mealList {
		temp1[k].User = value.ID
		temp1[k].Ic = value.Ic
		temp1[k].Orderid = value.Orderid
		temp1[k].Macid = value.Clockid
		temp1[k].Counterparty = value.TerminalName
		temp1[k].Kind = y.TransMeaning1(value.Mealtype)
		temp1[k].Cooperation = value.OpUser
		temp1[k].Operate = "-1" //减款
		temp1[k].Money = value.Money
		temp1[k].Balance = value.Balance
		temp1[k].Createdat = value.Createdat
		temp1[k].Dealtime = value.Opdate
		temp1[k].Remark = value.Remark
	}

	temp2 := make([]model.DealRecord, len(chargeList))
	for k, value := range chargeList {
		temp2[k].User = value.ID
		temp2[k].Ic = value.Ic
		temp2[k].Orderid = value.Orderid
		temp2[k].Macid = value.Clockid
		temp2[k].Counterparty = value.TerminalName
		temp2[k].Kind = value.BusinessType
		temp2[k].Cooperation = value.OpUser
		temp2[k].Operate = "1" //z款
		temp2[k].Money = value.Money
		temp2[k].Balance = value.Balance
		temp2[k].Createdat = value.Createdat
		temp2[k].Dealtime = value.Dealtime
		temp2[k].Remark = value.Remark
	}

	newlist := append(temp1, temp2...)
	// fmt.Printf("====== yhcv2查询参数 pid:%s stime:%s etime:%s 获取%d条数据===\n", p.Pid, p.Stime, p.Etime, len(newlist))

	if newlist != nil {
		response.Success(context, consts.CurdStatusOkMsg, gin.H{
			"count":        len(newlist),
			"chrage_count": len(chargeList),
			"meal_count":   len(mealList),
			"list":         newlist,
		})
	} else {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

func (y *Yhcv2) Season(times string) string {
	s, _ := time.Parse("2006-01-02 15:04:05", times)
	month := int(s.Month())
	season := []int{3, 6, 9, 12}
	for i := 0; i < len(season); i++ {
		if season[i]-month < 0 {
			season = append(season[:i], season[i+1:]...)
			i--
		}
	}

	table := "Record"
	switch season[0] {
	case 3:
		table += fmt.Sprintf("%d", s.Year()) + "01_Info"
	case 6:
		table += fmt.Sprintf("%d", s.Year()) + "02_Info"
	case 9:
		table += fmt.Sprintf("%d", s.Year()) + "03_Info"
	case 12:
		table += fmt.Sprintf("%d", s.Year()) + "04_Info"
	}
	return table
}

// 1 普通消费　2 时段固定消费(后台扣费) 3 菜单消费 4充值机 5退款机 6计时消费 7计次消费 8订餐　9取餐　10退餐
func (y *Yhcv2) TransMeaning1(kind string) string {
	switch kind {
	case "1":
		return "普通消费"
	case "2":
		return "时段固定消费"
	case "3":
		return "菜单消费"
	case "4":
		return "充值机"
	case "5":
		return "退款机"
	case "6":
		return "计时消费"
	case "7":
		return "计次消费"
	case "8":
		return "订餐"
	case "9":
		return "取餐"
	case "10":
		return "退餐"
	default:
		return "未知"
	}
}

// 获取用户
func (u *Yhcv2) UserInfo(context *gin.Context) {
	type Params struct { //类型绑定
		Pid string `form:"pid" json:"pid"  binding:"required"`
	}
	var p *Params

	if context.ShouldBindQuery(&p) != nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Printf("====== 获取用户 参数错误pid:%s =====\n", p.Pid)
		response.Fail(context, consts.CurdSelectFailCode, "参数错误 pid required!", "")
		return
	}

	info, err := yhcv2.MermberFactory("").GetMemberInfo(p.Pid)
	fmt.Printf("====== 获取用户KEY:%s =====\n", p.Pid)
	if err == nil {
		//会员状态 0 注销卡 1 未提交卡 2 坏卡 3 挂失卡 8 正常卡
		switch info.CardState {
		case "0":
			info.CardState = "注销卡"
		case "1":
			info.CardState = "未提交卡"
		case "2":
			info.CardState = "坏卡"
		case "3":
			info.CardState = "挂失卡"
		case "8":
			info.CardState = "正常卡"
		default:
			info.CardState = "未知"
		}
		response.Success(context, consts.CurdStatusOkMsg, info)
	} else {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// 消费记录
func (y *Yhcv2) MemberList(context *gin.Context) {
	type Params struct { //类型绑定
		Page  int `form:"page" json:"page"  binding:"required"`
		Limit int `form:"limit" json:"limit"  binding:"required"`
	}
	var p *Params

	if context.ShouldBindQuery(&p) != nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Printf("====== 获取用户 参数错误pid:%s =====\n", p.Limit)
		response.Fail(context, consts.CurdSelectFailCode, "参数错误 Limit required!", "")
		return
	}

	list, err := yhcv2.MermberFactory("").GetMembers(p.Page, p.Limit)

	for _, v := range list {
		fmt.Println("list", v.UserNO)
	}

	if err == nil {
		response.Success(context, consts.CurdStatusOkMsg, list)
	} else {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// 消费记录
func (y *Yhcv2) SetList(context *gin.Context) {
	type Params struct { //类型绑定
		Sid   string `form:"sid"`
		Pid   string `form:"pid"`
		Num   int    `form:"num"`
		Stime string `form:"stime"`
		Etime string `form:"etime"`
	}
	var p Params
	if context.ShouldBindQuery(&p) == nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Println("====== 添加订单 ======")
	}

	list, _ := yhcv2.MermberFactory("").GetMembers(1, 5)

	for _, v := range list {
		fmt.Println("list", v.UserNO)
		y.redisList(p.Sid, p.Num, fmt.Sprintf("%d", v.UserNO))
	}

	// 这里随便模拟一条数据返回
	response.Success(context, "ok", gin.H{
		"PeopleInfo": p.Pid,
		"content":    "门户新闻内容001",
	})
}

// 测试 redis 连接池
func (y *Yhcv2) redisList(list_key string, num int, pid string) {
	redisClient := redis_factory.GetOneRedisClient()
	for i := 1; i <= num; i++ {
		rand.Seed(time.Now().UnixNano())
		mm := rand.Intn(5)
		if mm == 0 {
			mm = 1
		}
		m := fmt.Sprintf("%d", mm)
		r := fmt.Sprintf("%d", rand.Intn(99999))
		n := fmt.Sprintf("%d", time.Now().UnixNano())
		t := fmt.Sprintf("%d", time.Now().Unix())
		dt := fmt.Sprintf("%d", time.Now().Unix()-int64(rand.Intn(1000)))
		ti := rand.Intn(100)
		if ti%2 == 0 {
			ti = 1
		} else {
			ti = 2
		}

		ty := fmt.Sprintf("%d", ti)

		orderid := "JX" + n + r
		id := fmt.Sprintf("%d", i)
		// list, _ := yhcv2.MermberFactory("").GetMembers(1, p.Limit)

		orlist := `{"id":` + id + `,"sid":44,"pid":6019,"lid":0,"student_id":` + pid + `,"ic":"","orderid":"` + orderid + `","price":` + m + `,"macid":"150","type":` + ty + `,"from":"农行支付","paystatus":true,"category":"3","sync":false,"created_at":` + t + `,"dealtime":` + dt + `}`

		sign := sign.Create(orlist, variable.ConfigYml.GetString("App.Secret"))
		// fmt.Println("orlist", orlist)
		list := `{"id":` + id + `,"sid":44,"pid":6019,"lid":0,"student_id":` + pid + `,"ic":"","orderid":"` + orderid + `","price":` + m + `,"macid":"150","type":` + ty + `,"from":"农行支付","paystatus":true,"category":"3","sync":false,"created_at":` + t + `,"dealtime":` + dt + `,"sign":"` + sign + `"}`

		_, err := redisClient.Int64(redisClient.Execute("LPUSH", list_key, list))
		if err != nil {
			fmt.Println("err", err)
		}
	}
	redisClient.ReleaseOneRedisClient()
}
