package model

type School struct {
	ID    int
	Name  string `gorm:"column:wxname"`
	Token string `gorm:"column:token"`
	Hurl  string `gorm:"column:hurl"`
	Ping  string `gorm:"column:ping"`
}

func (School) TableName() string {
	return "school"
}
