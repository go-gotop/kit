package center

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

type LogEntry struct {
	Service   string `json:"service"`
	Level     string `json:"level"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

// customLogger 是一个自定义的日志记录器，用于格式化日志时间戳。
type customLogger struct {
	logger log.Logger
}

func (c *customLogger) Log(level log.Level, keyvals ...interface{}) error {
	var timestamp string
	for i := 0; i < len(keyvals); i += 2 {
		if keyvals[i] == "ts" {
			if ts, ok := keyvals[i+1].(time.Time); ok {
				// 转换时间为上海时间
				loc, _ := time.LoadLocation("Asia/Shanghai")
				timestamp = ts.In(loc).Format("2006-01-02T15:04:05Z07:00")
				keyvals[i+1] = timestamp
			}
		}
	}
	err := c.logger.Log(level, keyvals...)
	if err != nil {
		return err
	}
	return nil
}

func newCustomLogger() log.Logger {
	return &customLogger{
		logger: log.NewStdLogger(os.Stdout),
	}
}

// RedisHandler 是一个log.Logger，将日志存储到Redis。
type RedisHandler struct {
	client      *redis.Client
	serviceName string // 日志json格式中的服务名 用做检索
}

type MultiLogger struct {
	loggers []log.Logger
}

func newMultiLogger(loggers ...log.Logger) *MultiLogger {
	return &MultiLogger{
		loggers: loggers,
	}
}

func (m *MultiLogger) Log(level log.Level, keyvals ...interface{}) error {
	for _, logger := range m.loggers {
		if err := logger.Log(level, keyvals...); err != nil {
			return err
		}
	}
	return nil
}

// Log 实现了log.Logger接口。
func (h *RedisHandler) Log(level log.Level, keyvals ...interface{}) error {
	if level == log.LevelInfo {
		return nil
	}
	levelField := levelToString(level)
	// 开始构建日志字符串，包含日志级别
	logStr := fmt.Sprintf("level=%s ", levelToString(level))
	// 遍历键值对，构造日志内容
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			if keyvals[i] == "ts" {
				if ts, ok := keyvals[i+1].(time.Time); ok {
					loc, _ := time.LoadLocation("Asia/Shanghai")
					keyvals[i+1] = ts.In(loc).Format("2006-01-02T15:04:05Z07:00")
				}
			}
			logStr += fmt.Sprintf("%s=%v ", keyvals[i], keyvals[i+1])
		} else {
			logStr += fmt.Sprintf("%s=MISSING_VALUE ", keyvals[i]) // 处理键没有值的情况
		}
	}
	nano := time.Now().UnixNano()
	key := fmt.Sprintf("log:%d", nano)
	entry := &LogEntry{
		Service:   h.serviceName,
		Level:     levelField,
		Timestamp: nano,
		Message:   logStr,
	}
	// 将日志条目序列化为 JSON
	jsonData, _ := json.Marshal(entry)

	// 使用 JSON.SET 存储到 Redis
	_, err := h.client.Do(context.Background(), "JSON.SET", key, ".", string(jsonData)).Result()
	if err != nil {
		log.Error(err)
	}
	// 设置过期时间
	err = h.client.Expire(context.Background(), key, 10*24*time.Hour).Err()
	if err != nil {
		log.Error(err)
	}
	return nil
}

// NewRedisHandler 创建一个新的RedisHandler实例。
func newRedisHandler(client *redis.Client, name string) *RedisHandler {
	return &RedisHandler{
		client:      client,
		serviceName: name,
	}
}

// NewRedisClient

func newRedisClient(addr, passwd string, db int32) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,  // no password set
		DB:       int(db), // use default DB
	})
	return rdb
}

func NewLogger(env, svcName, addr, passwd string, db int32) *MultiLogger {
	var multi *MultiLogger
	if env == "PRD" {
		stdout := log.NewStdLogger(os.Stdout)
		handler := newRedisHandler(newRedisClient(addr, passwd, db), svcName)
		multi = newMultiLogger(stdout, handler)
	} else {
		//stdout := newCustomLogger()
		multi = newMultiLogger(log.NewStdLogger(os.Stdout))
	}
	return multi
}

// levelToString 将日志级别转换为字符串
func levelToString(level log.Level) string {
	switch level {
	case log.LevelDebug:
		return "DEBUG"
	case log.LevelInfo:
		return "INFO"
	case log.LevelWarn:
		return "WARN"
	case log.LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
