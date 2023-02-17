package model

type Calllog struct {
	ID          int
	Sid         int     `gorm:"column:sid"`
	Pid         int     `gorm:"column:pid"`
	Key         string  `gorm:"column:key"`
	Ic          string  `gorm:"column:ic"`
	Describe    string  `gorm:"column:describe"`
	PhoneNumber string  `gorm:"column:phone_number"`
	CallTime    int     `gorm:"column:call_time"`
	Cost        float32 `gorm:"column:cost"`
	Stime       string  `gorm:"column:Stime"`
	Created_at  int64   `gorm:"column:created_at"`
}

func (Calllog) TableName() string {
	return "call_log"
}
