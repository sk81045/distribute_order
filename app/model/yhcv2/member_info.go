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
	AfterPay   float64 `gorm:"column:Balance" json:"balance"`
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

// 模糊查询
func (e *MemberInfo) GetMembers(page int, limit int) (temp []MemberInfo, er error) {
	page = page * limit
	sql := `SELECT TOP ` + fmt.Sprintf("%d", limit) + ` * FROM Member_Info WHERE 
	(MemberNo NOT IN (SELECT TOP ` + fmt.Sprintf("%d", page) + ` MemberNo FROM Member_Info ORDER BY MemberNo)) 
	ORDER BY MemberNo`
	result := e.Raw(sql).Find(&temp)
	if result.RowsAffected != 0 {
		return temp, nil
	} else {
		return nil, fmt.Errorf("查询失败未找到用户信息")
	}
}

// // 查询（根据关键词模糊查询）
// func (u *UsersModel) Show(userName string, limitStart, limitItems int) (counts int64, temp []UsersModel) {
// 	if counts = u.counts(userName); counts > 0 {
// 		sql := `
// 		SELECT  id, user_name, real_name, phone,last_login_ip, status, CONVERT(varchar(20), created_at, 120 ) as created_at, CONVERT(varchar(20), updated_at, 120 ) as updated_at
// 		FROM  tb_users  WHERE status=1 and   user_name like ? order  by id  desc OFFSET ? ROW FETCH NEXT ? ROWS ONLY
// 		`
// 		if res := u.Raw(sql, "%"+userName+"%", limitStart, limitItems).Find(&temp); res.RowsAffected > 0 {
// 			return counts, temp
// 		}
// 	}
// 	return 0, nil
// }
