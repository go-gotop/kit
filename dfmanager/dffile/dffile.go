package dffile

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/go-gotop/kit/dfmanager"
	"github.com/go-gotop/kit/dfmanager/dffile/csv"
	"github.com/go-gotop/kit/exchange"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

var (
	ErrCsvFileFinished = errors.New("csv file finished")
)

func NewFileDataFeed(opts ...Option) dfmanager.DataFeedManager {
	o := &options{
		logger: log.NewHelper(log.DefaultLogger),
	}

	for _, opt := range opts {
		opt(o)
	}

	return &df{
		name:    "File",
		opts:    o,
		streams: make(map[string]*stream),
	}
}

type stream struct {
	uuid       string
	symbol     string
	CancelFunc context.CancelFunc
}

type df struct {
	name    string
	opts    *options
	streams map[string]*stream
	mux     sync.Mutex
}

func (d *df) AddDataFeed(req *dfmanager.DataFeedRequest) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	if d.opts.path == "" {
		return errors.New("path is empty")
	}

	uuid := uuid.New().String()
	if req.ID != "" {
		uuid = req.ID
	}

	if _, ok := d.streams[uuid]; ok {
		return errors.New("stream already exists")
	}

	// 创建新的 context 和 cancel function
	ctx, cancel := context.WithCancel(context.Background())

	csvStream := csv.NewCSVDataFeed(d.opts.path, csv.WithStart(req.StartTime), csv.WithEnd(req.EndTime))

	tradeEventHandle := func(data *csv.TradeEvent) error {
		req.Event(&exchange.TradeEvent{
			TradeID:  fmt.Sprint(data.TradeID),
			Size:     data.Size,
			Price:    data.Price,
			Side:     exchange.SideType(data.Side),
			Symbol:   data.Symbol,
			TradedAt: data.TradedAt,
		})
		return nil
	}

	finishedEventHandle := func() error {
		req.ErrorHandler(ErrCsvFileFinished)
		return nil
	}

	streamReq := &csv.StreamRequest{
		Symbols:       []string{req.Symbol},
		Event:         tradeEventHandle,
		FinishedEvent: finishedEventHandle,
		Ctx:           ctx,
	}

	if err := csvStream.Trade(streamReq); err != nil {
		cancel()
		d.opts.logger.Errorf("csvFile.Trade error: %v", err)
		return err
	}

	d.streams[uuid] = &stream{
		uuid:       uuid,
		symbol:     req.Symbol,
		CancelFunc: cancel,
	}

	return nil
}

func (d *df) AddMarketPriceDataFeed(req *dfmanager.MarkPriceRequest) error {
	return nil
}

func (d *df) CloseDataFeed(id string) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	stream, ok := d.streams[id]
	if !ok {
		return errors.New("stream not found")
	}
	stream.CancelFunc()
	return nil
}

func (d *df) DataFeedList() []string {
	d.mux.Lock()
	defer d.mux.Unlock()

	var list []string
	for k := range d.streams {
		list = append(list, k)
	}
	return list
}

func (d *df) Name() string {
	return d.name
}

func (d *df) Shutdown() error {
	d.mux.Lock()
	defer d.mux.Unlock()

	for _, stream := range d.streams {
		stream.CancelFunc()
	}
	return nil
}
