package yhcv2

import (
	// "fmt"
	"goskeleton/app/model"
	// "time"
)

func RechargeInfoFactory(sqlType string) *RechargeInfo {
	return &RechargeInfo{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

type RechargeInfo struct {
	*model.BaseModel
	ID           string  `gorm:"column:MemberID" json:"ids"`
	Clockid      string  `gorm:"column:TerminalNo" json:"macID"`
	Ic           string  `gorm:"column:MemberCardNo" json:"ic"`
	Dealtime     string  `gorm:"column:RechargeTime" json:"dealtime"` //实际消费时间
	Createdat    string  `gorm:"column:CreateDate" json:"createdat"`  //创建时间
	OpUser       string  `gorm:"column:op_user" json:"cooperate"`
	CardSequ     int     `gorm:"column:card_sequ" json:"count"`
	Money        float32 `gorm:"column:Amount" json:"price"`
	Balance      float32 `gorm:"column:Balance" json:"balance"`
	BusinessType string  `gorm:"column:BusinessType"`
	Orderid      string  `gorm:"column:SerialNo" json:"orderid"`
	Remark       string  `gorm:"column:Remarks" json:"remark"`
	Terminal
}

func (RechargeInfo) TableName() string {
	return "Recharge_Info"
}

// 查询
func (m *RechargeInfo) List(empID string, Stime string, Etime string) (temp []RechargeInfo) {
	sql := `SELECT TOP 10 Recharge_Info.*,Terminal_Info.TerminalName
		FROM Recharge_Info JOIN Terminal_Info
		ON Recharge_Info.TerminalNo = Terminal_Info.TerminalNo
		WHERE Recharge_Info.MemberID = ?
		AND Recharge_Info.RechargeTime
		BETWEEN ? AND ?
		ORDER BY Recharge_Info.RechargeTime DESC`
	if res := m.Raw(sql, empID, Stime, Etime).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil
}

// func (rs *RecordInfo) Add(payorder model.Payorder) (ok bool) { //充值
// 	employee, err := RecordInfoFactory("").Fetch(payorder.Studentid)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	dealtime := time.Unix(payorder.Dealtime, 0)
// 	createtime := time.Unix(payorder.Created_at, 0)
// 	var blance = employee.AfterPay - payorder.Price
// 	result := rs.Omit("operate_type", "ID", "updated_at", "created_at", "Clock_name", "din_room_name").Create(&RecordInfo{
// 		Clockid:        payorder.Macid,
// 		Empid:          payorder.Studentid,
// 		Opdate:         dealtime.Format("2006-01-02 15:04:05"),
// 		GetTime:        createtime.Format("2006-01-02 15:04:05"),
// 		CardSequ:       employee.CardSequ + 1,
// 		Money:          payorder.Price,
// 		Balance:        blance,
// 		Mealtype:       "6",
// 		Kind:           "6",
// 		Cardid:         payorder.Orderid,
// 		Accountid:      employee.Accountid,
// 		OpUser:         payorder.From,
// 		SubsidyConsume: "0",
// 	})
// 	if result.Error != nil {
// 		fmt.Println("处理交易失败")
// 		return false
// 	} else {
// 		EmployeeFactory("").UpdateEmployee(employee.UserNO, blance, employee.CardSequ+1) //更新人事表
// 		return true
// 	}
// }
