package controller

import (
	"Hwgen/app/model"
	"Hwgen/global"
	"encoding/json"
	"fmt"
	// "io/ioutil"
	"time"
)

type Song struct {
}

func (rs *Song) Mission(value string) (ok bool) {
	var order = model.Payorder{}
	_ = json.Unmarshal([]byte(value), &order)
	switch order.Type {
	case 1:
		ok = rs.Rechage(order)
	case 2:
		ok = rs.Deduction(order)
	default:
		fmt.Println("未识别业务类型->", order.Type)
	}
	return ok
}

func (rs *Song) Rechage(order model.Payorder) (ok bool) { //充值
	fmt.Println("Rechage->", order)
	var employee = model.Employee{}
	global.H_DB.Model(&model.Employee{}).Where("Emp_id = ? ", order.Studentid).Find(&employee)
	fmt.Println("employee->", employee)

	now := time.Now()
	var blance = employee.AfterPay + order.Price
	var chargeData = model.MChargeRecords{
		Clockid:     0,
		Empid:       employee.UserNO,
		Opdate:      now.Format("2006-01-02 15:03:04"),
		CardSequ:    employee.CardSequ + 1,
		ChargeMoney: order.Price,
		CardBalance: blance,
		ChargeKind:  "0",
		Cardid:      order.Orderid,
		Accountid:   employee.UserNO,
	}

	dd := AddRecords(chargeData) //增加记录

	UpdateEmployee(employee.UserNO, blance, employee.CardSequ+1) //更新人事表
	return dd
}

func (rs *Song) Deduction(order model.Payorder) bool { //扣费
	fmt.Println("Deduction->", order)
	var employee = model.Employee{}
	global.H_DB.Model(&model.Employee{}).Where("Emp_id = ? ", order.Studentid).Find(&employee)
	fmt.Println("employee->", employee)

	now := time.Now()
	var blance = employee.AfterPay - order.Price
	var chargeData = model.MChargeRecords{
		Clockid:     0,
		Empid:       employee.UserNO,
		Opdate:      now.Format("2006-01-02 15:03:04"),
		CardSequ:    employee.CardSequ + 1,
		ChargeMoney: order.Price,
		CardBalance: blance,
		ChargeKind:  "0",
		Cardid:      order.Orderid,
		Accountid:   employee.UserNO,
	}

	dd := AddRecords(chargeData) //增加记录

	UpdateEmployee(employee.UserNO, blance, employee.CardSequ+1) //更新人事表
	return dd
}

func AddRecords(data model.MChargeRecords) bool {
	result := global.H_DB.Create(&data)
	if result.Error != nil {
		panic("func AddRecords():处理交易记录失败")
	}
	fmt.Println("result->", result)

	return true
}

func UpdateEmployee(id string, blance float32, sq int) {

	global.H_DB.Model(&model.Employee{}).Where("emp_id = ?", id).Updates(model.Employee{
		AfterPay: blance,
		CardSequ: sq,
	})
	fmt.Println("更新人事表")
}
