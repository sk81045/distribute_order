package smv1

import (
	"fmt"
	"goskeleton/app/model"
	"strconv"
	"time"
)

func MealRecordsFactory(sqlType string) *MealRecords {
	return &MealRecords{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

type MealRecords struct {
	*model.BaseModel
	OperateType string `json:"operate_type"`
	// ID             int     `gorm:"column:ID" json:"id"`
	Clockid        string  `gorm:"column:clock_id" json:"macID"`
	Empid          int     `gorm:"column:emp_id" json:"userNO"`
	Opdate         string  `gorm:"column:sign_time" json:"dealtime"` //实际消费时间
	GetTime        string  `gorm:"column:get_time" json:"createdat"` //创建时间
	OpUser         string  `gorm:"column:op_user" json:"cooperate"`
	CardSequ       int     `gorm:"column:card_sequ" json:"count"`
	Money          float64 `gorm:"column:card_consume" json:"price"`
	Balance        float64 `gorm:"column:card_balance" json:"balance"`
	Mealtype       string  `gorm:"column:mealtype"`
	Cardid         string  `gorm:"column:card_id" json:"orderid"`
	Accountid      string  `gorm:"column:account_id"`
	Kind           string  `gorm:"column:kind" json:"kind"`
	SubsidyConsume string  `gorm:"column:subsidy_consume"`
	SubsidyBalance string  `gorm:"column:subsidy_balance"`
	DinRoom
	Clock
}

func (MealRecords) TableName() string {
	return "MealRecordsReal"
}

// 查询
func (m *MealRecords) List(empID string, Stime string, Etime string) (temp []MealRecords) {
	sql := `SELECT top 50 MealRecordsReal.*,DinRoom.DinRoom_name, Clocks.Clock_name
		FROM MealRecordsReal JOIN Clocks
		ON MealRecordsReal.clock_id = Clocks.Clock_id
		JOIN DinRoom
		ON Clocks.DinRoom_id = DinRoom.DinRoom_id 
		WHERE MealRecordsReal.emp_id = ? 
		AND MealRecordsReal.sign_time 
		BETWEEN ? AND ?
		ORDER BY MealRecordsReal.ID DESC`
	if res := m.Raw(sql, empID, Stime, Etime).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil
}

func (rs *MealRecords) Add(payorder model.Payorder) error { //充值
	employee, err := EmployeeFactory("").Fetch(payorder.Studentid)
	if err != nil {
		return err
	}
	dealtime := time.Unix(payorder.Dealtime, 0)
	createtime := time.Unix(payorder.Created_at, 0)
	var money, _ = strconv.ParseFloat(payorder.Price, 64)
	var blance = employee.AfterPay + money
	var mealData = MealRecords{
		Clockid:        payorder.Macid,
		Empid:          payorder.Studentid,
		Opdate:         dealtime.Format("2006-01-02 15:04:05"),
		GetTime:        createtime.Format("2006-01-02 15:04:05"),
		CardSequ:       employee.CardSequ + 1,
		Money:          money,
		Balance:        blance,
		Mealtype:       "6",
		Kind:           "6",
		Cardid:         payorder.Orderid,
		Accountid:      employee.Accountid,
		OpUser:         payorder.From,
		SubsidyConsume: "0",
	}

	result := rs.Omit("operate_type", "ID", "updated_at", "created_at", "Clock_name", "din_room_name").Create(&mealData)
	if result.Error != nil {

		return fmt.Errorf("处理交易失败")
	} else {
		EmployeeFactory("").UpdateEmployee(employee.UserNO, blance, employee.CardSequ+1) //更新人事表
		return nil
	}
}
