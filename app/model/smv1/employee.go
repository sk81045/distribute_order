package smv1

import (
	"encoding/json"
	"fmt"
	// "gorm.io/gorm"
	// "errors"
	"goskeleton/app/model"
	"time"
)

func EmployeeFactory(sqlType string) *Employee {
	return &Employee{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

type Employee struct {
	*model.BaseModel
	Name      string  `gorm:"column:emp_fname"`
	UserNO    int     `gorm:"column:emp_id"`
	Cardid    string  `gorm:"column:card_sn"`
	MacID     string  `gorm:"column:macID"`
	MacType   string  `gorm:"column:macType"`
	CardSequ  int     `gorm:"column:card_sequ"`
	AfterPay  float32 `gorm:"column:card_balance"`
	PayCount  float32 `gorm:"column:subsidy_balance"`
	CardState string  `gorm:"column:isleaved"`
	Accountid string  `gorm:"column:account_id"`
	Isleaved  string  `gorm:"column:isleaved"`
	Issued    string  `gorm:"column:issued"`
	Departid  string  `gorm:"column:Depart_id"`
	Jobid     string  `gorm:"column:Job_id"`
}

func (Employee) TableName() string {
	return "Employee"
}

// 根据用户ID查询一条信息
func (e *Employee) Fetch(empID int) (*Employee, error) {
	sql := "SELECT * FROM Employee WHERE emp_id = ?"
	result := e.Raw(sql, empID).First(&e)
	if result.RowsAffected != 0 {
		return e, nil
	} else {
		return nil, fmt.Errorf("未找到用户信息")
	}
}

func (rs *Employee) Rechage(order string) (ok bool) { //充值
	var payorder = model.Payorder{}
	if err := json.Unmarshal([]byte(order), &payorder); err != nil {
		fmt.Println("model.Payorder{}解析失败:", err)
		return
	}

	employee, err := rs.Fetch(payorder.Studentid)
	if err != nil {
		fmt.Println(err)
		return
	}

	var blance = employee.AfterPay + payorder.Price
	var chargeData = MChargeRecords{
		Clockid:     payorder.Macid,
		Empid:       payorder.Studentid,
		Opdate:      time.Now().Format("2006-01-02 15:04:05"),
		CardSequ:    employee.CardSequ + 1,
		ChargeMoney: payorder.Price,
		CardBalance: blance,
		ChargeKind:  "6",
		Cardid:      payorder.Orderid,
		Accountid:   employee.Accountid,
		Kind:        payorder.Macid,
		OpUser:      payorder.From,
		Remark:      payorder.From,
	}

	result := rs.Create(&chargeData)
	if result.Error != nil {
		fmt.Println("处理交易记录失败")
		return false
	} else {
		fmt.Println("增款 +$", payorder.Price)
		rs.UpdateEmployee(employee.UserNO, blance, employee.CardSequ+1) //更新人事表
		return true
	}
}

func (rs *Employee) UpdateEmployee(empID int, blance float32, sq int) {
	result := rs.Model(&rs).Where("emp_id = ?", empID).Updates(&Employee{
		AfterPay: blance,
		CardSequ: sq,
	})

	if result.Error != nil {
		panic("处理交易记录失败")
	} else {
		fmt.Println("更新余额$", blance)
	}
}

func (rs *Employee) List(empID int) (temp []Employee) {
	sql := `SELECT * FROM Employee WHERE emp_id = ?  ORDER BY ID DESC`
	if res := rs.Raw(sql, empID).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil

	// global.GVA_DB.Model(&model.MealRecords{}).Where("emp_id = ?  AND sign_time between ? AND ? ORDER BY ID DESC", p.Factor, p.Stime, p.Etime).Find(&meal)
}

// func (u *Employee) AfterCreate(tx *gorm.DB) {
// 	fmt.Println("回调函数")

// 	// if u.ID == 1 {
// 	//   tx.Model(u).Update("role", "admin")
// 	// }
// 	// return
// }
