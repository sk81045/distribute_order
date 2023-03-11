package model

import (
	"gorm.io/gorm"
)

type Payorder struct {
	ID         int
	Sid        int     `gorm:"column:sid"`
	Pid        int     `gorm:"column:pid"`
	Lid        int     `gorm:"column:lid"`
	Studentid  int     `gorm:"column:studentid"`
	Ic         string  `gorm:"column:ic"`
	Orderid    string  `gorm:"column:orderid"`
	Price      float32 `gorm:"column:price"`
	Macid      string  `gorm:"column:macid"`
	Type       int     `gorm:"column:type"`
	From       string  `gorm:"column:from"`
	Paystatus  bool    `gorm:"column:paystatus"`
	Category   string  `gorm:"column:category"`
	Sync       bool    `gorm:"column:sync"`
	Created_at int64   `gorm:"column:created_at"`
	// Students   Students `gorm:"foreignKey:id;references:pid"`
}

func (Payorder) TableName() string {
	return "payorder"
}

func (p *Payorder) BeforeUpdate(tx *gorm.DB) (err error) {
	return
	//   if u.readonly() {
	//   err = errors.New("read only user")
	// }
	// return
}
