/*
clock_id：机号
account_id：帐号
emp_id：工号
op_date：充值日期
card_sequ：卡交易流水
charge_money：个人充值金额
card_balance：个人余额
subsidy_consume：补贴充值金额
subsidy_balance：补贴余额
charge_kind：类型 0现金,1补贴,2支付宝,3微信,4现金取款,5补贴取款,6银行充值,7银
*/
package model

type MChargeRecords struct {
	Clockid     string  `gorm:"column:clock_id"`
	Empid       string  `gorm:"column:emp_id"`
	Opdate      string  `gorm:"column:op_date"`
	CardSequ    int     `gorm:"column:card_sequ"`
	ChargeMoney float32 `gorm:"column:charge_money"`
	CardBalance float32 `gorm:"column:card_balance"`
	ChargeKind  string  `gorm:"column:charge_kind"`
	Cardid      string  `gorm:"column:card_id"`
	Accountid   string  `gorm:"column:account_id"`
}

func (MChargeRecords) TableName() string {
	return "MChargeRecords"
}
