package utils

import (
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	// "golang.org/x/text/transform"
	"Hwgen/global"
	"errors"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// @Description  16进制转10进制
// @param_1 16进制字符
func Hex2Dec(val string) string {
	n, err := strconv.ParseUint(val, 16, 32)
	int64Str := strconv.FormatUint(n, 10)
	if err != nil {
		fmt.Println(err)
	}
	return int64Str
}

// @Description  拼接字符串达到规定长度
func JoiningString(s1 string, s2 string, le int) string {
	var p string
	for i := 0; i < le; i++ {
		p += s2
	}
	var build strings.Builder
	build.WriteString(s1)
	build.WriteString(p)
	return build.String()
}

// @Description  组成TCP包Head(长度4)
func PackageHead(str string) string {
	piece2 := fmt.Sprintf("%d", len(str))
	piece1 := JoiningString("", "0", 4-len(piece2)) //拼接字符
	return piece1 + piece2
}

// @Description  生成随机数
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

func RandStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func ConvertStr2GBK(str string) string {
	//将utf-8编码的字符串转换为GBK编码
	ret, _ := simplifiedchinese.GBK.NewEncoder().String(str)
	return ret //如果转换失败返回空字符串

	// //如果是[]byte格式的字符串，可以使用Bytes方法
	// b, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(str))
	// return string(b)
}

//@description: 文件目录是否存在
//@param: path string
//@return: bool, error
func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, errors.New("存在同名文件")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//@author: [SliverHorn](https://github.com/SliverHorn)
//@function: GetWriteSyncer
//@description: zap logger中加入file-rotatelogs
//@return: zapcore.WriteSyncer, error

func GetWriteSyncer(file string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file, // 日志文件的位置
		MaxSize:    10,   // 在进行切割之前，日志文件的最大大小（以MB为单位）
		MaxBackups: 200,  // 保留旧文件的最大个数
		MaxAge:     30,   // 保留旧文件的最大天数
		Compress:   true, // 是否压缩/归档旧文件
	}

	if global.H_CONFIG.Zap.LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	}
	return zapcore.AddSync(lumberJackLogger)
}
