package yhcv2

import (
	"fmt"
	"goskeleton/app/model"
)

func MermberFactory(sqlType string) *MemberInfo {
	return &MemberInfo{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

type MemberInfo struct {
	*model.BaseModel
	ID         string  `gorm:"column:ID" json:"id"`
	Name       string  `gorm:"column:MemberName" json:"name"`
	UserNO     int     `gorm:"column:MemberNo" json:"user_no"`
	Cardid     string  `gorm:"column:MemberCardNo" json:"ic"`
	AfterPay   float32 `gorm:"column:Balance" json:"balance"`
	MerchantID string  `gorm:"column:MerchantID"`
	CardState  string  `gorm:"column:MemberState" json:"card_state"`
	CardType
}
type CardType struct { //卡类型
	CardsName string `gorm:"column:CardsName" json:"card_tag"`
	CardsNo   string `gorm:"column:CardsNo" json:"card_type"`
}

func (MemberInfo) TableName() string {
	return "Member_Info"
}

// 模糊查询
func (e *MemberInfo) GetMemberInfo(key string) (*MemberInfo, error) {
	sql := `SELECT Member_Info.*,CardType_Info.CardsName,CardType_Info.CardsNo FROM Member_Info JOIN CardType_Info
		ON Member_Info.CardType = CardType_Info.ID
		WHERE  MemberNo = ? OR MemberCardNo = ?`
	result := e.Raw(sql, key, key).First(&e)
	if result.RowsAffected != 0 {
		return e, nil
	} else {
		return nil, fmt.Errorf("查询失败未找到用户信息")
	}
}

// func (rs *MemberInfo) UpdateEmployee(empID int, blance float32, sq int) {
// 	result := rs.Model(&rs).Where("emp_id = ?", empID).Updates(&MemberInfo{
// 		AfterPay: blance,
// 		CardSequ: sq,
// 	})

// 	if result.Error != nil {
// 		panic("处理交易记录失败")
// 	} else {
// 		fmt.Println("=====================start=============================")
// 		fmt.Println("更新余额$", blance)
// 	}
// }
