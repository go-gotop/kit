package bnlimiter

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-gotop/kit/rate"
)

// 解析 period 字符串，返回 time.Duration 和 int
func parsePeriod(period string) (time.Duration, int, error) {
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
func SetAllLimiters(periodLimitArray []PeriodLimit) map[string][]*rate.Limiter {
	limiterMap := make(map[string][]*rate.Limiter)
	limiterMap[spotCreateOrderLimit] = SetLimiterMap(periodLimitArray, "spotCreateOrderPeriod", "spotCreateOrderTimes")
	limiterMap[futureCreateOrderLimit] = SetLimiterMap(periodLimitArray, "futureCreateOrderPeriod", "futureCreateOrderTimes")
	limiterMap[spotNormalRequestLimit] = SetLimiterMap(periodLimitArray, "spotNormalRequestPeriod", "spotNormalRequestTimes")
	return limiterMap
}

// 动态添加限流器通用函数
func SetLimiterMap(periodLimitArray []PeriodLimit, periodField string, timesField string) []*rate.Limiter {
	if periodField == "" || timesField == "" {
		return nil
	}
	limiterGroup := make([]*rate.Limiter, 0)
	for _, pl := range periodLimitArray {
		plValue := reflect.ValueOf(pl)
		periodFieldValue := plValue.FieldByName(periodField).String()
		timesFieldValue := plValue.FieldByName(timesField).Int()
		if periodFieldValue != "" && timesFieldValue != 0 {
			timeUnit, duration, err := parsePeriod(periodFieldValue)
			if err != nil {
				log.Printf("parse period error: %v", err)
				continue
			}
			every := timeUnit * time.Duration(duration) / time.Duration(timesFieldValue)
			// 如果every超过1ms，则every设置为1ms
			if every > time.Millisecond {
				every = time.Millisecond
			}
			limiterGroup = append(limiterGroup, rate.NewLimiterWithPeriod(rate.Every(every), int(timesFieldValue), timeUnit*time.Duration(duration)))
		}
	}
	return limiterGroup
}

func limiterAllow(l []*rate.Limiter) bool {
	for _, limiter := range l {
		if !limiter.AllowC() {
			return false
		}
	}
	return true
}
