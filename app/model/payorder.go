package model

import (
	"gorm.io/gorm"
)

type Payorder struct {
	ID         int64  `json:"id"`
	Sid        int    `gorm:"column:sid" json:"sid"`
	Pid        int    `gorm:"column:pid" json:"pid"`
	Lid        int    `gorm:"column:lid" json:"lid"`
	Studentid  int    `gorm:"column:student_id" json:"student_id"`
	Ic         string `gorm:"column:ic" json:"ic"`
	Orderid    string `gorm:"column:orderid" json:"orderid"`
	Price      string `gorm:"column:price" json:"price"`
	Macid      string `gorm:"column:macid" json:"macid"`
	Type       int    `gorm:"column:type" json:"type"`
	From       string `gorm:"column:from" json:"from"`
	Paystatus  int    `gorm:"column:paystatus" json:"paystatus"`
	Category   string `gorm:"column:category" json:"category"`
	Sync       int    `gorm:"column:sync" json:"sync"`
	Created_at int64  `gorm:"column:created_at" json:"created_at"`
	Dealtime   int64  `gorm:"column:deal_time" json:"dealtime"`
	Mark       string `gorm:"column:mark" json:"mark"`
	Error      string `gorm:"column:error" json:"error"`
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
