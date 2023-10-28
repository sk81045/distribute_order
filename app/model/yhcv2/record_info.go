package yhcv2

import (
	"fmt"
	"goskeleton/app/model"
	"time"
)

func RecordInfoFactory(sqlType string) *RecordInfo {
	return &RecordInfo{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

type RecordInfo struct {
	*model.BaseModel
	ID        string  `gorm:"column:MemberID" json:"ids"`
	Clockid   string  `gorm:"column:TerminalNo" json:"macID"`
	Ic        string  `gorm:"column:MemberCardNo" json:"ic"`
	Opdate    string  `gorm:"column:ConsumeTime" json:"dealtime"` //实际消费时间
	Createdat string  `gorm:"column:CreateDate" json:"createdat"` //创建时间
	OpUser    string  `gorm:"column:op_user" json:"cooperate"`
	CardSequ  int     `gorm:"column:card_sequ" json:"count"`
	Money     float32 `gorm:"column:Amount" json:"price"`
	Balance   float32 `gorm:"column:Balance" json:"balance"`
	Mealtype  string  `gorm:"column:TerminalType"`
	Orderid   string  `gorm:"column:SerialNo" json:"orderid"`
	Remark    string  `gorm:"column:Remarks" json:"remark"`
	Terminal
}
type Terminal struct { //操作终端
	TerminalName string `gorm:"column:TerminalName" json:"TerminalName"`
}

func (RecordInfo) TableName() string {
	return "Record_Info"
}

// 查询
func (r *RecordInfo) List(empID string, Stime string, Etime string) (temp []RecordInfo) {
	tablename := r.TableTransMean(Etime)
	fmt.Println("tablename", tablename)
	fmt.Println("empID", empID)
	sql := `SELECT  TOP 10 ` + tablename + `.*,Terminal_Info.TerminalName
		FROM ` + tablename + ` JOIN Terminal_Info
		ON ` + tablename + `.TerminalNo = Terminal_Info.TerminalNo
		WHERE ` + tablename + `.MemberID = ?
		AND ` + tablename + `.ConsumeTime
		BETWEEN ? AND ?
		ORDER BY ` + tablename + `.ConsumeTime DESC`
	if res := r.Raw(sql, empID, Stime, Etime).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil
}

func (r *RecordInfo) TableTransMean(times string) string {
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
