package model

type DeductionParams struct {
	MerchantID      string  `json:"MerchantID"`
	MemberID        string  `json:"MemberID"`
	MemberNo        int     `json:"MemberNo"`
	ConsumptionTime string  `json:"ConsumptionTime"`
	TerminalNo      string  `json:"TerminalNo"`
	Amount          float32 `json:"Amount"`
	SubsidiesAmount float32 `json:"SubsidiesAmount"`
	GiftAmount      float32 `json:"GiftAmount"`
	Times           int     `json:"Times"`
	Remarks         string  `json:"Remarks"`
	AuthkeyType     int     `json:"AuthkeyType"`
	Authkey         string  `json:"Authkey"`
}
