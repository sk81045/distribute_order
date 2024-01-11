package smv1

import (
	"fmt"
	"goskeleton/app/global/variable"
	"goskeleton/app/model"
	"strconv"
	"time"
)

func MChargeRecordsFactory(sqlType string) *MChargeRecords {
	return &MChargeRecords{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

/*
CREATE TABLE
  DinRoom (
    DinRoom_id nvarchar(30) NOT NULL,
    DinRoom_name nvarchar(20) NULL,
    Mem nvarchar(50) NULL,
    id int IDENTITY(1, 1) NOT NULL,
  )
ALTER TABLE
  DinRoom
ADD
  CONSTRAINT PK_DinRoom PRIMARY KEY (DinRoom_id)
*/

type MChargeRecords struct {
	*model.BaseModel
	OperateType string `json:"operate_type"`
	// ID             int     `gorm:"column:ID" json:"id"`
	Clockid        string  `gorm:"column:clock_id" json:"macID"`
	Empid          int     `gorm:"column:emp_id" json:"userNO"`
	Opdate         string  `gorm:"column:op_date" json:"dealtime"`   //实际消费时间
	GetTime        string  `gorm:"column:get_time" json:"createdat"` //创建时间
	OpUser         string  `gorm:"column:op_user" json:"cooperate"`
	CardSequ       int     `gorm:"column:card_sequ" json:"count"`
	Money          float64 `gorm:"column:charge_money" json:"price"`
	Balance        float64 `gorm:"column:card_balance" json:"balance"`
	Kind           string  `gorm:"column:Kind" json:"kind"`
	ChargeKind     string  `gorm:"column:charge_Kind" json:"Chargekind"`
	Cardid         string  `gorm:"column:card_id" json:"orderid"`
	Accountid      string  `gorm:"column:account_id"`
	SubsidyConsume string  `gorm:"column:subsidy_consume"`
	SubsidyBalance string  `gorm:"column:subsidy_balance"`
	Remark         string  `gorm:"column:remark"`
	DinRoom
	Clock
}

type DinRoom struct { //交易方
	DinRoom_name string `json:"DinRoom_name"`
}
type Clock struct { //操作终端
	Clock_name string `json:"Clock_name"`
}

func (MChargeRecords) TableName() string {
	return "MChargeRecords"
}

// 查询（根据关键词模糊查询）
func (m *MChargeRecords) List(empID string, Stime string, Etime string) (temp []MChargeRecords) {
	sql := `SELECT top 300 *
FROM MChargeRecords JOIN Clocks
ON MChargeRecords.clock_id = Clocks.Clock_id
JOIN DinRoom
ON Clocks.DinRoom_id = DinRoom.DinRoom_id 
WHERE MChargeRecords.emp_id = ? 
AND MChargeRecords.op_date 
BETWEEN ? AND ?
ORDER BY MChargeRecords.ID DESC`
	if res := m.Raw(sql, empID, Stime, Etime).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil
}

func (m *MChargeRecords) GetOrder(empID string, Oid string) (temp []MChargeRecords) {
	sql := `SELECT top 300 *
FROM MChargeRecords JOIN Clocks
ON MChargeRecords.clock_id = Clocks.Clock_id
JOIN DinRoom
ON Clocks.DinRoom_id = DinRoom.DinRoom_id 
WHERE MChargeRecords.emp_id = ? OR MChargeRecords.card_id = ?
ORDER BY MChargeRecords.ID DESC`
	if res := m.Raw(sql, empID, Oid).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil
}

func (rs *MChargeRecords) Add(payorder model.Payorder) error { //充值
	employee, err := EmployeeFactory("").Fetch(payorder.Studentid)
	if err != nil {
		return err
	}
	dealtime := time.Unix(payorder.Dealtime, 0)
	createtime := time.Unix(payorder.Created_at, 0)
	var money, _ = strconv.ParseFloat(payorder.Price, 64)
	//增款===============================================
	var blance = employee.AfterPay + money
	//===================================================
	var chargeData = MChargeRecords{
		Clockid:    payorder.Macid,
		Empid:      payorder.Studentid,
		Opdate:     dealtime.Format("2006-01-02 15:04:05"),
		GetTime:    createtime.Format("2006-01-02 15:04:05"),
		CardSequ:   employee.CardSequ + 1,
		Money:      money,
		Balance:    blance,
		Kind:       variable.ConfigYml.GetString("Order.ChargeKind"),
		ChargeKind: variable.ConfigYml.GetString("Order.ChargeKind"),
		Cardid:     payorder.Orderid,
		Accountid:  employee.Accountid,
		OpUser:     payorder.From,
		Remark:     payorder.From,
	}

	result := rs.Omit("operate_type", "ID", "updated_at", "created_at", "Clock_name", "din_room_name").Create(&chargeData)
	if result.Error != nil {
		return fmt.Errorf("处理交易失败")
	} else {
		EmployeeFactory("").UpdateEmployee(employee.UserNO, blance, employee.CardSequ+1) //更新人事表
		return nil
	}
}
