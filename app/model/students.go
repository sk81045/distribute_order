package model

type Students struct {
	ID         int
	Sid        int       `gorm:"column:sid"`
	Name       string    `gorm:"column:name"`
	Studentid  string    `gorm:"column:studentid"`
	Balance    float32   `gorm:"column:balance"`
	Cardid     string    `gorm:"column:cardid"`
	Grade      string    `gorm:"column:grade"`
	Class      string    `gorm:"column:class"`
	Created_at string    `gorm:"column:created_at"`
	Parents    []Parents `gorm:"foreignKey:lid"`
}

type Parents struct {
	ID     int
	Lid    int    `gorm:"column:lid"`
	Parent string `gorm:"column:parent"`
	Phone  string `gorm:"column:phone"`
	Guanxi string `gorm:"column:guanxi"`
}
