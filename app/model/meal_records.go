package model

func MealRecordsFactory(sqlType string) *MealRecords {
	return &MealRecords{BaseModel: BaseModel{DB: UseDbConn(sqlType)}}
}

/*
clock_id：机号
account_id：帐号
emp_id：工号
sign_time：消费日期
card_sequ：卡交易流水
pos_sequ：机器交易流水
card_consume：个人消费金额
card_balance：个人余额
subsidy_consume：补贴消费金额
subsidy_balance：补贴余额
kind：餐段标志：1早 2中 3晚 4夜
flag：刷卡标识
get_time：采集时间
*/
type MealRecords struct {
	BaseModel
	ID          int     `gorm:"column:ID"`
	Clockid     int     `gorm:"column:clock_id"`
	Empid       string  `gorm:"column:emp_id"`
	Opdate      string  `gorm:"column:sign_time"`
	CardSequ    int     `gorm:"column:card_sequ"`
	ChargeMoney float64 `gorm:"column:card_consume"`
	CardBalance float64 `gorm:"column:card_balance"`
	ChargeKind  string  `gorm:"column:kind"`
	Cardid      string  `gorm:"column:card_id"`
	Accountid   string  `gorm:"column:account_id"`
	Type        string
}

func (MealRecords) TableName() string {
	return "MealRecordsReal"
}

// // // 查询（根据关键词模糊查询）
func (m *MealRecords) List(empID string, Stime string, Etime string) (temp []MealRecords) {
	sql := `SELECT * FROM MealRecordsReal WHERE emp_id = ? AND sign_time between ? AND ? ORDER BY ID DESC`
	if res := m.Raw(sql, empID, Stime, Etime).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil

	// global.GVA_DB.Model(&model.MealRecords{}).Where("emp_id = ?  AND sign_time between ? AND ? ORDER BY ID DESC", p.Factor, p.Stime, p.Etime).Find(&meal)
}

// 根据用户ID查询一条信息
// func (u *MealRecords) List(userId int) (*MealRecords, error) {
// 	sql := "SELECT * FROM MealRecords WHERE ID = ?"
// 	result := u.Raw(sql, userId).First(u)
// 	if result.Error == nil {
// 		return u, nil
// 	} else {
// 		return nil, result.Error
// 	}
// meal := MealRecords{}
// u.Where("ID = ?", userId).First(&u)
// u.Model(u).Where("ID = ?", userId).First(&u)
// u.Model(u).Where("user_name=?", tmp.UserName).Count(&counts)
// var counts int64
// var tmp MealRecords
// u.Model(u).Where("ID = ?", userId).First(&tmp)
// return tmp, nil
// }

// // // 查询（根据关键词模糊查询）
// func (u *MealRecords) Show(userName string, limitStart, limitItems int) (counts int64, temp []MealRecords) {
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
