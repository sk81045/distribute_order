package controller

import (
	"fmt"
)

type Api struct {
}

func (rs *Api) Send(value string) {
	fmt.Println("api", value)
}
