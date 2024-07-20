package csv

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

type CsvFile struct {
	symbol string
	dir    string
	opts   *options
}

func NewCSVDataFeed(dir string, ops ...Options) *CsvFile {
	opts := &options{}
	for _, op := range ops {
		op(opts)
	}
	symbol := filepath.Base(dir)
	return &CsvFile{
		symbol: symbol,
		dir:    dir,
		opts:   opts,
	}
}

func (c *CsvFile) Close() error {
	return nil
}

func (c *CsvFile) CloseSymbol(id string) error {
	return nil
}

func (c *CsvFile) Trade(req *StreamRequest) error {
	doneC := make(chan struct{})
	defer close(doneC)

	files, err := readCSVFileNames(c.dir, c.opts.start, c.opts.end)
	if err != nil {
		return err
	}

	go c.readCSVFiles(req, files, c.opts.start, c.opts.end)

	return nil
}

func (c *CsvFile) ReadCSVFile(filePath string) ([]*TradeData, error) {
	return readCSVFile(filePath)
}

func (c *CsvFile) readCSVFiles(req *StreamRequest, files []string, start int64, end int64) {
	path := c.dir
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	eventChan := make(chan *TradeEvent, 1)
	finishedEventChan := make(chan struct{}, 1)

	go func() {
		defer close(eventChan)
		defer close(finishedEventChan)

		for _, f := range files {
			select {
			case <-req.Ctx.Done():
				return
			default:
				if err := c.processFile(req.Ctx, path+f, eventChan, start, end); err != nil {
					return
				}
			}
		}
		finishedEventChan <- struct{}{}
	}()

	for event := range eventChan {
		req.Event(event)
	}

	<-finishedEventChan
	req.FinishedEvent()
}

func (c *CsvFile) toTradeEvent(t *TradeData) (*TradeEvent, error) {
	price, err := decimal.NewFromString(t.Price)
	if err != nil {
		return nil, err
	}
	size, err := decimal.NewFromString(t.Size)
	if err != nil {
		return nil, err
	}
	// TODO side 如何取
	return &TradeEvent{
		TradeID:  t.TradeID,
		Size:     size,
		Price:    price,
		Side:     t.Side,
		Symbol:   c.symbol,
		TradedAt: t.TradedAt,
	}, nil
}

func (c *CsvFile) processFile(ctx context.Context, filePath string, eventChan chan<- *TradeEvent, start int64, end int64) error {
	data, err := readCSVFile(filePath)
	if err != nil {
		return err
	}
	for _, v := range data {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 检查时间戳是否在所需范围内
			if (start == 0 || v.TradedAt >= start) && (end == 0 || v.TradedAt <= end) {
				v1, err := c.toTradeEvent(v)
				if err != nil {
					return err
				}
				eventChan <- v1
			} else if v.TradedAt > end {
				log.Infof("逐笔数据读取完成，退出时间戳：tradedAt: %v", v.TradedAt)
				return nil
			}
		}
	}
	return nil
}

func readCSVFile(f string) ([]*TradeData, error) {
	file, err := os.Open(f)
	if err != nil {
		log.Errorf("failed to open file: %v", err)
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)

	// 读取 csv 文件中的表头
	headers, err := r.Read()
	if err != nil {
		return nil, err
	}

	rows := make([]*TradeData, 0, 3000)
	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		row, err := toTradeData(headers, record)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func readCSVFileNames(path string, start int64, end int64) ([]string, error) {
	// 确保路径以斜杠结尾
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// 打开目录
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	// 获取目录下所有文件
	files, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	// 过滤CSV文件并存储文件名和时间戳
	var fileNames []string
	var fileTimestamps []int64
	re := regexp.MustCompile(`^(\d+)\.csv$`)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".csv") {
			match := re.FindStringSubmatch(file.Name())
			if len(match) != 2 {
				// 文件名不是时间戳类型，跳过
				continue
			}
			// 从文件名解析时间戳
			timestamp, err := strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				return nil, err
			}

			// 检查时间戳是否在指定范围内
			if (start == 0 || timestamp >= normalizeTimestamp(start)) && (end == 0 || timestamp <= normalizeTimestamp(end)) {
				fileTimestamps = append(fileTimestamps, timestamp)
			}
		}
	}

	// 按照文件名排序
	sort.Slice(fileTimestamps, func(i, j int) bool {
		return fileTimestamps[i] < fileTimestamps[j]
	})
	for _, ts := range fileTimestamps {
		fileNames = append(fileNames, strconv.FormatInt(ts, 10)+".csv")
	}
	return fileNames, nil
}

func normalizeTimestamp(timestamp int64) int64 {
	return timestamp - timestamp%(3600*1000)
}

func toTradeData(headers []string, record []string) (*TradeData, error) {
	row := &TradeData{}
	for i, value := range record {
		switch headers[i] {
		case "trade_id":
			tradeID, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}
			row.TradeID = tradeID
		case "size":
			row.Size = value
		case "price":
			row.Price = value
		case "side":
			row.Side = value
		case "symbol":
			row.Symbol = value
		case "quote":
			row.Quote = value
		case "traded_at":
			tradedAt, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			row.TradedAt = tradedAt
		}
	}
	return row, nil
}
