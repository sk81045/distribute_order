package yhcv1

import (
	"fmt"
	"goskeleton/app/model"
)

func RecordInfoFactory(sqlType string) *FlowRecord {
	return &FlowRecord{BaseModel: (&model.BaseModel{DB: model.UseDbConn(sqlType)})}
}

type FlowRecord struct {
	*model.BaseModel
	ID           int64   `gorm:"column:FID" json:"ids"`
	UserNO       string  `gorm:"column:userNO" json:"user_no"`
	Clockid      string  `gorm:"column:macID" json:"macID"`
	Ic           string  `gorm:"column:cardID" json:"ic"`
	Opdate       string  `gorm:"column:payTime" json:"dealtime"`     //实际消费时间
	Createdat    string  `gorm:"column:CreateDate" json:"createdat"` //创建时间
	Cooperate    string  `gorm:"column:addMode" json:"cooperate"`
	CardSequ     int     `gorm:"column:card_sequ" json:"count"`
	Money        float64 `gorm:"column:payMoney" json:"price"`
	Balance      float64 `gorm:"column:afterPay" json:"balance"`
	Orderid      string  `gorm:"column:PayListNO" json:"orderid"`
	Remark       string  `gorm:"column:Remarks" json:"remark"`
	Terminal     string  `gorm:"column:devGrpName" json:"macType"`
	Terminaltype string  `gorm:"column:devType"`
	Cooperation  string  `gorm:"column:note"`
}

func (FlowRecord) TableName() string {
	return "FlowRecord"
}

// 查询
func (r *FlowRecord) List(empID int64, Stime string, Etime string) (temp []FlowRecord) {
	sql := `SELECT top 300 * FROM 
			FlowRecord JOIN DevGrpAndDev
			ON FlowRecord.macID = DevGrpAndDev.devID
			JOIN DevGrp
			ON DevGrpAndDev.devgrpID = DevGrp.devgrpID
			JOIN DevInfo
			ON DevGrpAndDev.devID = DevInfo.devID
			WHERE FlowRecord.userNO = ? 
			AND FlowRecord.payTime 
			BETWEEN ? AND ?
			ORDER BY FlowRecord.FID DESC`
	if res := r.Raw(sql, empID, Stime, Etime).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil
}

// 查询
func (f *FlowRecord) GetOrder(empID string, Oid string) (temp []FlowRecord) {
	sql := `SELECT top 300 * FROM 
			FlowRecord JOIN DevGrpAndDev
			ON FlowRecord.macID = DevGrpAndDev.devID
			JOIN DevGrp
			ON DevGrpAndDev.devgrpID = DevGrp.devgrpID
			JOIN DevInfo
			ON DevGrpAndDev.devID = DevInfo.devID
			WHERE FlowRecord.userNO = ? 
			AND FlowRecord.PayListNO = ? 
			ORDER BY FlowRecord.FID DESC`
	if res := f.Raw(sql, empID, Oid).Find(&temp); res.RowsAffected > 0 {
		return temp
	}
	return nil
}

type AggregateTemp struct {
	DevgrpID   int     `gorm:"column:devgrpID"`
	Devgrpname string  `gorm:"column:devgrpname"`
	Cnt        int     `gorm:"column:cnt"`
	Totalmoney float64 `gorm:"column:totalmeony"`
}

// 消费汇总
func (r *FlowRecord) Aggregate(Stime string, Etime string) (temp []AggregateTemp, er error) {
	sql := `select DevGrp.devgrpID,devgrpname,count(*) cnt ,sum(payMoney) totalmeony from FlowRecord,CardInfo,UserInfo,devgrpAndDev,dbo.DevGrp where FlowRecord.userNO=CardInfo.UserNO and CardInfo.UserNO=UserInfo.UserNO  and FlowRecord.payTime >= ? and FlowRecord.payTime < ? and macID=devID and devgrpAndDev.devgrpID=DevGrp.devgrpID group by DevGrp.devgrpID,devgrpname`
	if res := r.Raw(sql, Stime, Etime).Find(&temp); res.RowsAffected > 0 {
		return temp, nil
	}
	return nil, fmt.Errorf("查询失败未找到信息")
}
