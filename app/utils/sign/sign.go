package sign

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	// "strconv"
	"time"
)

func Create(order string, secret string) string {
	var orderMap map[string]interface{}
	_ = json.Unmarshal([]byte(order), &orderMap)
	//ksort
	var keys []string
	// orderMap["secret"] = secret
	for k := range orderMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var http_build_query string
	for _, k := range keys {
		ty := judgeTypeToString(orderMap[k])
		http_build_query = http_build_query + k + "=" + ty + "&"
	}
	ff := http_build_query[0 : len(http_build_query)-1]
	sign := MD5(ff + secret)
	return sign
}

func Verify(order string, secret string) error {
	var orderMap map[string]interface{}
	_ = json.Unmarshal([]byte(order), &orderMap)
	//ksort
	now := time.Now().Unix()
	timestamp := int64(orderMap["created_at"].(float64))
	if now-timestamp > 3600 {
		fmt.Println("sign 超时")
		// return false
	}
	originsign := orderMap["sign"]
	delete(orderMap, "sign")
	var keys []string
	for k := range orderMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	var http_build_query string
	for _, k := range keys {
		ty := judgeTypeToString(orderMap[k])
		http_build_query = http_build_query + k + "=" + ty + "&"
	}
	ff := http_build_query[0 : len(http_build_query)-1]
	sign := MD5(ff + secret)
	if sign != originsign {
		return fmt.Errorf("sign验证失败" + sign)
	}
	return nil
}

func MD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func judgeTypeToString(v interface{}) string {
	switch i := v.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%v", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case string:
		return fmt.Sprintf("%s", v)
	case float32:
		return fmt.Sprintf("%.0f", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		_ = i
		return "unknown"
	}
}
