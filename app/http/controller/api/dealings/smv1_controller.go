package dealings

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goskeleton/app/global/consts"
	"goskeleton/app/global/variable"
	"goskeleton/app/model"
	"goskeleton/app/model/smv1"
	"goskeleton/app/utils/redis_factory"
	"goskeleton/app/utils/response"
	"goskeleton/app/utils/sign"
	"math/rand"
	"time"
)

type Smv1 struct {
}

// 1.门户类首页新闻
func (u *Smv1) News(context *gin.Context) {

	//  由于本项目骨架已经将表单验证器的字段(成员)绑定在上下文，因此可以按照 GetString()、GetInt64()、GetFloat64（）等快捷获取需要的数据类型
	// 当然也可以通过gin框架的上下文原原始方法获取，例如： context.PostForm("name") 获取，这样获取的数据格式为文本，需要自己继续转换
	newsType := context.GetString(consts.ValidatorPrefix + "newsType")
	page := context.GetFloat64(consts.ValidatorPrefix + "page")
	limit := context.GetFloat64(consts.ValidatorPrefix + "limit")
	userIp := context.ClientIP()

	// 这里随便模拟一条数据返回
	response.Success(context, "ok", gin.H{
		"newsType": newsType,
		"page":     page,
		"limit":    limit,
		"userIp":   userIp,
		"title":    "API门户首页公司新闻标题001-ceshi",
		"content":  "门户新闻内容001",
	})

}

// 消费记录
func (u *Smv1) List(context *gin.Context) {

	type Params struct { //类型绑定
		Pid   string `form:"pid"`
		Stime string `form:"stime"`
		Etime string `form:"etime"`
	}
	var p Params
	if context.ShouldBindQuery(&p) != nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Printf("====== 查询参数错误 pid:%s stime:%s etime:%s=====\n", p.Pid, p.Stime, p.Etime)
	}

	chargeList := smv1.MChargeRecordsFactory("").List(p.Pid, p.Stime, p.Etime)
	mealList := smv1.MealRecordsFactory("").List(p.Pid, p.Stime, p.Etime)

	temp1 := make([]model.DealRecord, len(chargeList))
	for k, value := range chargeList {
		// fmt.Println("temp1 value", value)
		temp1[k].ID = value.Id
		temp1[k].User = fmt.Sprint("%s", value.Empid)
		temp1[k].Orderid = value.Cardid
		temp1[k].Macid = value.Clockid
		temp1[k].Counterparty = value.DinRoom_name
		temp1[k].Kind = value.Clock_name
		temp1[k].Cooperation = value.OpUser
		temp1[k].Operate = "1"
		temp1[k].Money = value.Money
		temp1[k].Balance = value.Balance
		temp1[k].Createdat = value.GetTime
		temp1[k].Dealtime = value.Opdate
	}

	temp2 := make([]model.DealRecord, len(mealList))
	// fmt.Println("temp2 value", mealList)
	// return
	for k, value := range mealList {
		temp2[k].ID = value.Id
		temp2[k].User = fmt.Sprint("%s", value.Empid)
		temp2[k].Orderid = value.Cardid
		temp2[k].Macid = value.Clockid
		temp2[k].Counterparty = value.DinRoom_name
		temp2[k].Kind = value.Clock_name
		temp2[k].Cooperation = value.OpUser
		temp2[k].Operate = "-1"
		temp2[k].Money = value.Money
		temp2[k].Balance = value.Balance
		temp2[k].Createdat = value.GetTime
		temp2[k].Dealtime = value.Opdate
	}

	newlist := append(temp1, temp2...)
	fmt.Printf("====== 查询参数 pid:%s stime:%s etime:%s 获取%d条数据===\n", p.Pid, p.Stime, p.Etime, len(newlist))

	if newlist != nil {
		response.Success(context, consts.CurdStatusOkMsg, gin.H{
			"list":         newlist,
			"count":        len(newlist),
			"chrage_count": len(chargeList),
			"meal_count":   len(mealList),
		})
	} else {
		fmt.Println("Fail")
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// 消费记录
func (u *Smv1) UserInfo(context *gin.Context) {
	type Params struct { //类型绑定
		Pid string `form:"pid" json:"pid"  binding:"required"`
	}
	var p *Params

	if context.ShouldBindQuery(&p) != nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Printf("====== 获取用户 参数错误pid:%s =====\n", p.Pid)
		response.Fail(context, consts.CurdSelectFailCode, "参数错误 pid required!", "")
		return
	}

	info, err := smv1.EmployeeFactory("").Employee(p.Pid)
	if err == nil {
		if info.Status == "false" && info.Flag == "0" {
			info.CardState = "正常"
		} else {
			if info.Status == "true" {
				info.CardState = "挂失"
			}
			if info.Flag == "1" {
				info.CardState = "退卡"
			} else if info.Flag == "2" {
				info.CardState = "已补办"
			} else {
				info.CardState = "异常"
			}
		}
		response.Success(context, consts.CurdStatusOkMsg, info)
	} else {
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// 消费记录
func (u *Smv1) SetList(context *gin.Context) {
	type Params struct { //类型绑定
		Pid   string `form:"pid"`
		Num   int    `form:"num"`
		Stime string `form:"stime"`
		Etime string `form:"etime"`
	}
	var p Params
	if context.ShouldBindQuery(&p) == nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Println("====== 添加订单 ======")
	}

	list_key := p.Pid
	redisList(list_key, p.Num)

	// 这里随便模拟一条数据返回
	response.Success(context, "ok", gin.H{
		"PeopleInfo": p.Pid,
		"content":    "门户新闻内容001",
	})
}

// 测试 redis 连接池
func redisList(list_key string, num int) {
	redisClient := redis_factory.GetOneRedisClient()
	for i := 1; i <= num; i++ {
		rand.Seed(time.Now().UnixNano())
		m := fmt.Sprintf("%d", rand.Intn(5))
		r := fmt.Sprintf("%d", rand.Intn(99999))
		n := fmt.Sprintf("%d", time.Now().UnixNano())
		t := fmt.Sprintf("%d", time.Now().Unix()-int64(rand.Intn(1000)))
		ti := rand.Intn(100)

		if ti%2 == 0 {
			ti = 1
		} else {
			ti = 2
		}

		ty := fmt.Sprintf("%d", ti)

		orderid := "JX" + n + r
		id := fmt.Sprintf("%d", i)

		orlist := `{"id":` + id + `,"sid":44,"pid":6019,"lid":0,"student_id":210049,"ic":"","orderid":"` + orderid + `","price":` + m + `,"macid":"150","type":` + ty + `,"from":"农行支付","paystatus":true,"category":"3","sync":false,"created_at":` + t + `,"dealtime":1686153600}`

		sign := sign.Create(orlist, variable.ConfigYml.GetString("App.Secret"))
		fmt.Println("orlist", orlist)
		list := `{"id":` + id + `,"sid":44,"pid":6019,"lid":0,"student_id":210049,"ic":"","orderid":"` + orderid + `","price":` + m + `,"macid":"150","type":` + ty + `,"from":"农行支付","paystatus":true,"category":"3","sync":false,"created_at":` + t + `,"dealtime":1686153600,"sign":"` + sign + `"}`

		_, err := redisClient.Int64(redisClient.Execute("LPUSH", list_key, list))
		if err != nil {
			fmt.Println("err", err)
		}
	}
	redisClient.ReleaseOneRedisClient()
}
