package yhcv1

type RechargeParams struct { //通过webserver接口充值时，需引用的结构体
	AccountID int     `json:"AccountID"`
	CardID    string  `json:"CardID"`
	PayMoney  float32 `json:"PayMoney"`
	PayTime   string  `json:"PayTime"`
	MacID     string  `json:"MacID"`
	MacType   string  `json:"MacType"`
	PayKind   string  `json:"PayKind"`
	OrderNO   string  `json:"OrderNO"`
}
