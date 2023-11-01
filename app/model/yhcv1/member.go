package yhcv1

import (
	"fmt"
	"goskeleton/app/model"
)

func MermberFactory(sqlType string) *MemberInfo {
	return &MemberInfo{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

type MemberInfo struct {
	*model.BaseModel
	ID         string  `gorm:"column:Card_tid" json:"id"`
	Name       string  `gorm:"column:UserName" json:"name"`
	UserNO     int64   `gorm:"column:UserNO" json:"user_no"`
	Cardid     string  `gorm:"column:cardID" json:"ic"`
	Balance    float32 `gorm:"column:cash" json:"balance"`
	MerchantID string  `gorm:"column:MerchantID"`
	CardState  string  `gorm:"column:cardState" json:"card_state"`
	CardType
}
type CardType struct { //卡类型
	CardsName string `gorm:"column:LevelName" json:"card_tag"`
	CardsNo   string `gorm:"column:Level_id" json:"card_type"`
}

func (MemberInfo) TableName() string {
	return "CardInfo"
}

// 查询
func (e *MemberInfo) GetMemberInfo(key int64) (*MemberInfo, error) {
	sql := `SELECT CardInfo.*,LevelInfo.LevelName,LevelInfo.Level_id  
		FROM CardInfo JOIN LevelInfo
		ON CardInfo.cardLevel = LevelInfo.level_id
		WHERE UserNO = ? OR cardID = ?`
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
	sql := `SELECT TOP ` + fmt.Sprintf("%d", limit) + ` * FROM CardInfo WHERE 
		(Card_tid NOT IN (SELECT TOP ` + fmt.Sprintf("%d", page) + ` Card_tid FROM CardInfo ORDER BY Card_tid)) 
		ORDER BY Card_tid`
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
