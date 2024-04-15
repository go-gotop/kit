package dfbinance

import (
	"github.com/go-gotop/kit/dfmanager"
)

var _ dfmanager.DataFeedManager = (*df)(nil)

func NewBinanceDataFeed() dfmanager.DataFeedManager {
	return &df{}
}

type df struct {
	id string
}