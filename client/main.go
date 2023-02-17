package main

import (
	"Hwgen/app/service"
	"Hwgen/core"
	"Hwgen/global"
)

func main() {
	global.H_VIPER = core.Viper() // Viper读取配置文件
	// global.H_DB = core.Gorm()     // gorm连接数据库
	global.H_LOG = core.Zaps()    // zap日志库
	global.H_REDIS = core.Redis() //Redis
	// db, _ := global.H_DB.DB()
	// defer db.Close()
	service.Run()
}
