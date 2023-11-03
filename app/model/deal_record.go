package model

type DealRecord struct {
	ID           int64   `json:"id"`
	User         string  `json:"user"`
	Ic           string  `json:"ic"`
	Macid        string  `json:"macid"`
	Counterparty string  `json:"counterparty"`
	Kind         string  `json:"kind"`
	Cooperation  string  `json:"cooperation"`
	Operate      string  `json:"operate"`
	Orderid      string  `json:"orderid"`
	Money        float64 `json:"money"`
	Balance      float64 `json:"balance"`
	Createdat    string  `json:"createdat"`
	Dealtime     string  `json:"dealtime"`
	Remark       string  `json:"remark"`
}
