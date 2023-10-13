package model

type MealRecords struct {
	Clockid        string  `gorm:"column:clock_id"`
	Empid          string  `gorm:"column:emp_id"`
	Opdate         string  `gorm:"column:sign_time"`
	GetTime        string  `gorm:"column:get_time"`
	CardSequ       int     `gorm:"column:card_sequ"`
	ChargeMoney    float32 `gorm:"column:card_consume"`
	CardBalance    float32 `gorm:"column:card_balance"`
	ChargeKind     string  `gorm:"column:kind"`
	Cardid         string  `gorm:"column:card_id"`
	Accountid      string  `gorm:"column:account_id"`
	SubsidyConsume float32 `gorm:"column:subsidy_consume"`
	SubsidyBalance float32 `gorm:"column:subsidy_balance"`
}

func (MealRecords) TableName() string {
	return "MealRecordsReal"
}
