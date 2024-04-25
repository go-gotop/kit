package limiter

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-gotop/kit/rate"
	"github.com/redis/go-redis/v9"
)

// 解析 period 字符串，返回 time.Duration 和 int
func ParsePeriod(period string) (time.Duration, int, error) {
	var unit time.Duration

	// 去除字符串中的空格
	period = strings.TrimSpace(period)

	// 获取数字部分
	var numStr string
	var unitStr string
	for i, char := range period {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		} else {
			unitStr = period[i:]
			break
		}
	}
	// 解析数字部分
	num, err := strconv.Atoi(numStr)
	log.Printf("numStr: %v, unitStr %v", numStr, unitStr)
	if err != nil {
		return 0, 0, err
	}
	// 解析时间单位部分
	switch strings.ToLower(unitStr) {
	case "ms":
		unit = time.Millisecond
	case "s":
		unit = time.Second
	case "m":
		unit = time.Minute
	case "h":
		unit = time.Hour
	default:
		return 0, 0, fmt.Errorf("unsupported time unit: %s", unitStr)
	}
	return unit, num, nil
}

// 动态添加所有限流器
func SetAllLimiters(accountId string, redis redis.Client, periodLimitArray []PeriodLimit) map[string][]*rate.Limiter {
	ip, err := GetOutBoundIP()
	if err != nil {
		log.Printf("get out bound ip error: %v", err)
	}
	limiterMap := make(map[string][]*rate.Limiter)
	// 限流器唯一标识用于 redis key，对于websocket只对ip限制，其他请求对accountId限制
	limiterMap[WsConnectLimit] = SetLimiterMap("BINANCE_WSCONNECT_"+ip, redis, periodLimitArray, "WsConnectPeriod", "WsConnectTimes")
	limiterMap[SpotCreateOrderLimit] = SetLimiterMap("BINANCE_SPOTCREATEORDER_"+accountId, redis, periodLimitArray, "SpotCreateOrderPeriod", "SpotCreateOrderTimes")
	limiterMap[FutureCreateOrderLimit] = SetLimiterMap("BINANCE_FUTURECREATEORDER_"+accountId, redis, periodLimitArray, "FutureCreateOrderPeriod", "FutureCreateOrderTimes")
	limiterMap[SpotNormalRequestLimit] = SetLimiterMap("BINANCE_NORMALREQUEST_"+accountId, redis, periodLimitArray, "SpotNormalRequestPeriod", "SpotNormalRequestTimes")
	return limiterMap
}

// 动态添加限流器通用函数
func SetLimiterMap(uniq string, redis redis.Client, periodLimitArray []PeriodLimit, periodField string, timesField string) []*rate.Limiter {
	if periodField == "" || timesField == "" {
		return nil
	}
	limiterGroup := make([]*rate.Limiter, 0)
	for _, pl := range periodLimitArray {
		plValue := reflect.ValueOf(pl)
		periodFieldValue := plValue.FieldByName(periodField).String()
		timesFieldValue := plValue.FieldByName(timesField).Int()
		if periodFieldValue != "" && timesFieldValue != 0 {
			timeUnit, duration, err := ParsePeriod(periodFieldValue)
			if err != nil {
				log.Printf("parse period error: %v", err)
				continue
			}
			every := timeUnit * time.Duration(duration) / time.Duration(timesFieldValue)
			// 如果every超过1ms，则every设置为1ms
			if every > time.Millisecond {
				every = time.Millisecond
			}
			// 每种请求的限流可能不同周期限制不一样，所以唯一标识需要再拼接上周期
			limiterGroup = append(limiterGroup, rate.NewLimiterWithPeriod(uniq+"_"+periodFieldValue, redis, rate.Every(every), int(timesFieldValue), timeUnit*time.Duration(duration)))
		}
	}
	return limiterGroup
}

func LimiterAllow(l []*rate.Limiter) bool {
	for _, limiter := range l {
		if !limiter.AllowC() {
			return false
		}
	}
	return true
}

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}
