package core

import (
	"gorm.io/gorm"
	// "gin/config"
	"Hwgen/global"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
)

// Gorm 初始化数据库并产生数据库全局变量
func Gorm() *gorm.DB {
	switch global.H_CONFIG.System.DbType {
	case "mysql":
		return GormMysql()
	case "mssql":
		return GormMssql()
	default:
		return GormMysql()
	}
}

// GormMysql 初始化Mysql数据库
func GormMysql() *gorm.DB {
	m := global.H_CONFIG.Mysql
	if m.Dbname == "" {
		return nil
	}
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}

	if db, err := gorm.Open(mysql.New(mysqlConfig)); err != nil {
		fmt.Println("数据库连接失败", mysqlConfig)
		return nil
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		fmt.Println("数据库已连接")
		return db
	}
}

//  初始化sqlServer数据库
func GormMssql() *gorm.DB {
	m := global.H_CONFIG.Mssql

	if m.Dbname == "" {
		return nil
	}
	fmt.Println("Dbname", m.Dbname)
	mssqlConfig := sqlserver.Config{
		DSN: m.Dsn(), // DSN data source name
	}
	fmt.Println("Dsn", mssqlConfig.DSN)
	if db, err := gorm.Open(sqlserver.Open(mssqlConfig.DSN), &gorm.Config{}); err != nil {
		fmt.Println("数据库连接失败", err)
		return nil
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		fmt.Println("数据库已连接")
		return db
	}
}
