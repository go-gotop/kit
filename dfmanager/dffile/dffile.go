package dffile

import (
	"errors"
	"sync"

	"github.com/go-gotop/kit/dfmanager"
	"github.com/go-kratos/kratos/v2/log"
)

func NewFileDataFeed(opts ...Option) dfmanager.DataFeedManager {
	o := &options{
		logger: log.NewHelper(log.DefaultLogger),
	}

	for _, opt := range opts {
		opt(o)
	}

	return &df{
		name: "file",
		opts: o,
	}
}

type df struct {
	name string
	path string
	opts *options
	mux  sync.Mutex
}

func (d *df) AddDataFeed(req *dfmanager.DataFeedRequest) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	if d.path == "" {
		return errors.New("path is empty")
	}

	
	return nil
}

func (d *df) CloseDataFeed(id string) error {
	return nil
}

func (d *df) DataFeedList() []string {
	return nil
}

func (d *df) Name() string {
	return d.name
}

func (d *df) Shutdown() error {
	return nil
}
