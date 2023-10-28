package smv1

import (
	"fmt"
	"goskeleton/app/model"
)

func EmployeeFactory(sqlType string) *Employee {
	return &Employee{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

type Employee struct {
	*model.BaseModel
	ID        int     `gorm:"column:ID" json:"id"`
	Name      string  `gorm:"column:emp_fname" json:"name"`
	UserNO    int     `gorm:"column:emp_id" json:"user_no"`
	Cardid    string  `gorm:"column:card_sn" json:"ic"`
	CardSequ  int     `gorm:"column:card_sequ"`
	AfterPay  float32 `gorm:"column:card_balance" json:"balance"`
	PayCount  float32 `gorm:"column:subsidy_balance"`
	CardState string  `gorm:"column:isleaved" json:"card_state"`
	Accountid string  `gorm:"column:account_id"`
	Isleaved  string  `gorm:"column:isleaved"`
	Issued    string  `gorm:"column:issued"`
	Departid  string  `gorm:"column:Depart_id"`
	Jobid     string  `gorm:"column:Job_id"`
	State
}
type State struct { //IC卡状态
	Status string `json:"status"`
	Flag   string `json:"flag"`
}

func (Employee) TableName() string {
	return "Employee"
}

// 根据用户ID查询一条信息
func (rs *Employee) Fetch(empID int) (*Employee, error) {
	sql := `SELECT * FROM Employee WHERE emp_id = ?`
	if res := rs.Raw(sql, empID).First(&rs); res.RowsAffected > 0 {
		return rs, nil
	}
	return nil, fmt.Errorf("未找到用户信息")
}

// 模糊查询
func (e *Employee) Employee(key string) (*Employee, error) {
	sql := `SELECT *
			FROM Employee LEFT JOIN CardInfoTemp
			ON Employee.emp_id = CardInfoTemp.emp_id
			WHERE Employee.emp_id = ? OR  Employee.card_sn = ?`
	result := e.Raw(sql, key, key).First(&e)
	if result.RowsAffected != 0 {
		return e, nil
	} else {
		return nil, fmt.Errorf("未找到用户信息")
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
		fmt.Println("=====================start=============================")
		fmt.Println("更新余额$", blance)
	}
}
