package service

import (
	redis_method "Hwgen/app/controller/redis"
	"Hwgen/app/model"
	"Hwgen/global"
	helpers "Hwgen/utils"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	// "sync"
	"time"
)

var (
	redis redis_method.RedisStore
)

type Message struct {
	Type     int
	Describe string
	Content  string
	Pid      string
	Sid      int
}

func Run() {
	Server()
}

func Server() {
	key1 := "OrderDone"
	key2 := "BadOrder"
	go OrderDone(key1)
	go BadOrder(key2)
	go Alert()
	for {

	}
}

//// 返回从0开始到-1位置之间的数据，意思就是返回全部数据
// vals, err := rdb.LRange(ctx,"key",0,-1).Result()
// if err != nil {
// 	panic(err)
// }
// fmt.Println(vals)

func OrderDone(key string) {
	for {
		time.Sleep(5 * time.Second)

		rlen := redis.LLen(key)
		fmt.Printf("待更新订单 :%d\n", rlen)
		if rlen > 0 {
			payorder := model.Payorder{}
			list := redis.LRange(key, 0, -1)
			for _, i := range list {

				if err := json.Unmarshal([]byte(i), &payorder); err != nil {
					fmt.Println("序列化失败", err)
					continue
				}

				if err := SaveSync(payorder.Sid, payorder.Orderid); err != nil {
					redis.SetList("update_error", i)
					fmt.Println("更新订单状态失败", err)
					global.H_LOG.Warn("更新订单状态失败!:", zap.String("redis rlen:", string(i)))
				}

				l1 := redis.LRem(key, i)
				Backups := fmt.Sprintf("%d", payorder.Sid) + "Backups"
				l2 := redis.LRem(Backups, i) //删除备份列表
				fmt.Println("lrem", l1, l2)
				paytime := time.Unix(payorder.Dealtime, 0)
				fmt.Printf("\033[7;32;2m更新订单成功|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
			}
		}
	}
}

func BadOrder(key string) {
	for {
		time.Sleep(5 * time.Second)
		rlen := redis.LLen(key)
		fmt.Printf("获取错误订单 :%d\n", rlen)
		if rlen > 0 {
			payorder := model.Payorder{}
			list := redis.LRange(key, 0, -1)
			for _, i := range list {

				if err := json.Unmarshal([]byte(i), &payorder); err != nil {
					fmt.Println("错误订单序列化失败", err)
					continue
				}

				if err := PrcessBadOrder(payorder.Sid, payorder.Orderid, payorder.Error); err != nil {
					redis.SetList("update_error", i)
					global.H_LOG.Warn("更新错误订单状态失败!:", zap.String("redis rlen:", string(i)))
				}

				l1 := redis.LRem(key, i)
				fmt.Println("lrem", key, l1)
				paytime := time.Unix(payorder.Dealtime, 0)
				fmt.Printf("\033[7;32;2m更新错误订单成功|操作%d|$%s|订单号:%s|用户编号:%d|交易时间:%s\033[0m\n", payorder.Type, payorder.Price, payorder.Orderid, payorder.Studentid, paytime.Format("2006-01-02 15:04:05"))
			}
		}
	}
}

func Alert() {
	for {
		sc := School()
		for _, v := range sc {
			if v.Ping == "1" {
				continue
			}
			code, _ := helpers.HttpGetStatusCode(v.Hurl)
			if code != 200 {
				fmt.Println("NetWork error list", v.Name, v.Hurl)
				helpers.HttpGet("http://xb36.yoozhi.cn/api/outside/pubmsg/NotifyBalance?sid=34&scname=" + v.Name + "&code=" + fmt.Sprintf("%d", code) + "&url=" + v.Hurl)
			}
		}
		time.Sleep(7200 * time.Second)
	}
}

func SaveSync(sid int, oid string) error {
	result := global.H_DB.Model(model.Payorder{}).Where("sid = ? AND orderid = ? ", sid, oid).Updates(model.Payorder{
		Paystatus: 5,
	})
	if result.RowsAffected == 0 {
		return fmt.Errorf("处理订单状态失败")
	}
	return nil
}

func PrcessBadOrder(sid int, oid string, err string) error {
	result := global.H_DB.Model(model.Payorder{}).Where("sid = ? AND orderid = ? ", sid, oid).Updates(model.Payorder{
		Paystatus: 3,
		Mark:      err,
	})
	if result.RowsAffected == 0 {
		return fmt.Errorf("处理订单状态失败")
	}
	return nil
}

// func (rs *Employee) UpdateEmployee(empID int, blance float64, sq int) {
// 	result := rs.Model(&rs).Where("emp_id = ?", empID).Updates(&Employee{
// 		AfterPay: blance,
// 		CardSequ: sq,
// 	})

// }

func Payorder(sid int) []model.Payorder {
	var list = []model.Payorder{}
	_ = global.H_DB.Model(&model.Payorder{}).Where("`paystatus` = ? AND `sid` = ? AND `sync` = ? AND `price` > ?", 0, sid, 0, 0).Limit(50).Find(&list).Error
	return list
}

func School() []model.School {
	var school_list = []model.School{}
	_ = global.H_DB.Model(&model.School{}).Find(&school_list).Error
	return school_list
}
