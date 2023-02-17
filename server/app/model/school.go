package model

type School struct {
	ID    int
	Name  string `gorm:"column:wxname"`
	Token string `gorm:"column:token"`
}

func (School) TableName() string {
	return "school"
}
