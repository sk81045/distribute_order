package model

type Employee struct {
	Name      string  `gorm:"column:emp_fname"`
	UserNO    string  `gorm:"column:emp_id"`
	Cardid    string  `gorm:"column:card_sn"`
	MacID     string  `gorm:"column:macID"`
	MacType   string  `gorm:"column:macType"`
	CardSequ  int     `gorm:"column:card_sequ"`
	AfterPay  float32 `gorm:"column:card_balance"`
	PayCount  float32 `gorm:"column:subsidy_balance"`
	CardState string  `gorm:"column:isleaved"`
	// CardStatus string `gorm:"column:CardInfoTemp"`
	// CardStatus CardInfoTemp `gorm:"foreignKey:card_sn"`
}

func (Employee) TableName() string {
	return "Employee"
}
