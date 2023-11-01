package dealings

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goskeleton/app/global/consts"
	"goskeleton/app/global/variable"
	"goskeleton/app/model"
	"goskeleton/app/model/yhcv1"
	"goskeleton/app/utils/redis_factory"
	"goskeleton/app/utils/response"
	"goskeleton/app/utils/sign"
	"math/rand"
	"time"
)

type Yhcv1 struct {
}

// 消费记录
func (y *Yhcv1) RecordList(context *gin.Context) {

	type Params struct { //类型绑定
		Pid   int64  `form:"pid"`
		Stime string `form:"stime"`
		Etime string `form:"etime"`
	}
	var p Params
	if context.ShouldBindQuery(&p) != nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Printf("====== yhc查询参数错误 pid:%s stime:%s etime:%s=====\n", p.Pid, p.Stime, p.Etime)
	}

	fmt.Printf("====== yhc查询参数 pid:%d stime:%s etime:%s=====\n", p.Pid, p.Stime, p.Etime)

	user, err := yhcv1.MermberFactory("").GetMemberInfo(p.Pid)
	if err != nil {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, err.Error())
		return
	}

	list := yhcv1.RecordInfoFactory("").List(user.UserNO, p.Stime, p.Etime)
	temp1 := make([]model.DealRecord, len(list))
	for k, value := range list {
		temp1[k].ID = value.ID
		temp1[k].User = value.UserNO
		temp1[k].Ic = value.Ic
		temp1[k].Orderid = value.Orderid
		temp1[k].Macid = value.Clockid
		temp1[k].Counterparty = value.Terminal
		temp1[k].Kind = value.Terminaltype
		temp1[k].Cooperation = value.Cooperation
		temp1[k].Operate = value.Cooperate //减款
		temp1[k].Money = value.Money
		temp1[k].Balance = value.Balance
		temp1[k].Createdat = value.Createdat
		temp1[k].Dealtime = value.Opdate
		temp1[k].Remark = value.Remark
	}

	if list != nil {
		response.Success(context, consts.CurdStatusOkMsg, gin.H{
			"list": temp1,
		})
	} else {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// 获取用户
func (u *Yhcv1) UserInfo(context *gin.Context) {
	type Params struct { //类型绑定
		Pid int64 `form:"pid" json:"pid"  binding:"required"`
	}
	var p *Params

	if context.ShouldBindQuery(&p) != nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Printf("====== 获取用户 参数错误pid:%s =====\n", p.Pid)
		response.Fail(context, consts.CurdSelectFailCode, "参数错误 pid required!", "")
		return
	}

	info, err := yhcv1.MermberFactory("").GetMemberInfo(p.Pid)
	fmt.Printf("====== 获取用户KEY:%d =====\n", p.Pid)
	if err == nil {
		//会员状态  //1.正常 2.挂失 3.注销
		switch info.CardState {
		case "1":
			info.CardState = "正常"
		case "2":
			info.CardState = "挂失"
		case "3":
			info.CardState = "注销"
		default:
			info.CardState = "未知"
		}
		response.Success(context, consts.CurdStatusOkMsg, info)
	} else {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// 用户列表
func (y *Yhcv1) MemberList(context *gin.Context) {
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

	list, err := yhcv1.MermberFactory("").GetMembers(p.Page, p.Limit)

	if err == nil {
		response.Success(context, consts.CurdStatusOkMsg, list)
	} else {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

func (y *Yhcv1) SetList(context *gin.Context) {
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

	// y.redisList(p.Sid, p.Num, p.Pid)
	// return

	list, _ := yhcv1.MermberFactory("").GetMembers(1, 5)

	for _, v := range list {
		fmt.Println("list", v.UserNO)
		y.redisList(p.Sid, p.Num, fmt.Sprintf("%d", v.UserNO), v.Cardid)
	}

	// 这里随便模拟一条数据返回
	response.Success(context, "ok", gin.H{
		"PeopleInfo": p.Pid,
		"content":    "门户新闻内容001",
	})
}

// 测试 redis 连接池
func (y *Yhcv1) redisList(list_key string, num int, pid string, ic string) {
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
		macid := "0"
		if ti%2 == 0 {
			ti = 1
			macid = "150"
		} else {
			macid = "100"
			ti = 2
		}

		ty := fmt.Sprintf("%d", ti)

		orderid := "JX" + n + r
		id := fmt.Sprintf("%d", i)
		// list, _ := yhcv2.MermberFactory("").GetMembers(1, p.Limit)

		orlist := `{"id":` + id + `,"sid":44,"pid":6019,"lid":0,"student_id":` + pid + `,"ic":"` + ic + `","orderid":"` + orderid + `","price":` + m + `,"macid":"` + macid + `","type":` + ty + `,"from":"农行支付","paystatus":true,"category":"3","sync":false,"created_at":` + t + `,"dealtime":` + dt + `}`

		sign := sign.Create(orlist, variable.ConfigYml.GetString("App.Secret"))
		// fmt.Println("orlist", orlist)
		list := `{"id":` + id + `,"sid":44,"pid":6019,"lid":0,"student_id":` + pid + `,"ic":"` + ic + `","orderid":"` + orderid + `","price":` + m + `,"macid":"` + macid + `","type":` + ty + `,"from":"农行支付","paystatus":true,"category":"3","sync":false,"created_at":` + t + `,"dealtime":` + dt + `,"sign":"` + sign + `"}`

		_, err := redisClient.Int64(redisClient.Execute("LPUSH", list_key, list))
		if err != nil {
			fmt.Println("err", err)
		}
	}
	redisClient.ReleaseOneRedisClient()
}
