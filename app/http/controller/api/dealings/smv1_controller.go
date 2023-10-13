package dealings

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goskeleton/app/global/consts"
	"goskeleton/app/model"
	"goskeleton/app/utils/response"
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
	pid := context.GetString(consts.ValidatorPrefix + "pid")
	stime := context.GetString(consts.ValidatorPrefix + "stime")
	etime := context.GetString(consts.ValidatorPrefix + "etime")
	fmt.Println("pse", pid, stime, etime)
	showlist := model.MealRecordsFactory("").List(pid, stime, etime)
	if showlist != nil {
		response.Success(context, consts.CurdStatusOkMsg, gin.H{"list": showlist})
	} else {
		fmt.Println("Fail")
		response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// 消费记录
func (u *Smv1) PeopleInfo(context *gin.Context) {
	type Params struct { //类型绑定
		Pid   string `form:"pid"`
		Stime string `form:"stime"`
		Etime string `form:"etime"`
	}
	var p Params
	if context.ShouldBindQuery(&p) == nil { //ShouldBindQuery 函数只绑定Get参数
		fmt.Println("====== 消费-查询参数 ======")
	}

	// 这里随便模拟一条数据返回
	response.Success(context, "ok", gin.H{
		"PeopleInfo": p.Pid,
		"content":    "门户新闻内容001",
	})

}
