package proxydeal

import (
	"fmt"
	"goskeleton/app/global/variable"
	"goskeleton/app/model/smv1"
	"goskeleton/app/model/yhcv1"
	"goskeleton/app/model/yhcv2"
)

type ProxyDeal struct {
	Yhcv1 yhcv1.MemberInfo
	Yhcv2 yhcv2.MemberInfo
	Smv1  smv1.Employee
}

// 查询
func (p *ProxyDeal) GetMembers(page int64, limit int64, where string) (interface{}, error) {
	fmt.Println("where", where)
	switch variable.ConfigYml.GetString("Order.SellVersion") {
	case "smv1":
		fmt.Println("smv1类型!")
		list, err := smv1.EmployeeFactory("").GetMembers(page, limit, "card_balance < 50")
		return list, err
	case "yhcv1":
		fmt.Println("yhcv1类型!")
		list, err := yhcv1.MermberFactory("").GetMembers(page, limit, "cash < 50")
		return list, err
	case "yhcv2":
		fmt.Println("yhcv2类型!")
		list, err := yhcv2.MermberFactory("").GetMembers(page, limit, "Balance < 50")
		return list, err
	default:
		fmt.Println("未识别 SellVersion 类型!")
		return nil, nil
	}
}

// 用户总数
func (e *ProxyDeal) MemberNums(where string) int64 {
	switch variable.ConfigYml.GetString("Order.SellVersion") {
	case "smv1":
		return smv1.EmployeeFactory("").MemberNums("card_balance < 50")
	case "yhcv1":
		return yhcv1.MermberFactory("").MemberNums("cash < 50")
	case "yhcv2":
		return yhcv2.MermberFactory("").MemberNums("Balance < 50")
	default:
		fmt.Println("未识别 SellVersion 类型!")
		return 0
	}
}
