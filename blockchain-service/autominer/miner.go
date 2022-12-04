package autominer

import (
	"context"
	"log"
	"time"
)

type blockchainer interface {
	Mine() (int64, error)
}

type miner struct {
	tickerTime time.Duration
	blockchainer
}

func New(d time.Duration, b blockchainer) miner {
	return miner{
		tickerTime:   d,
		blockchainer: b,
	}
}

func (c *miner) Start(ctx context.Context) {
	t := time.NewTicker(c.tickerTime)
	run := true
	go func(c context.Context) {
		for {
			select {
			case <-c.Done():
				run = false
				return
			}
		}
	}(ctx)
	for {
		select {
		case <-t.C:
			if run {
				c.do()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *miner) do() {
	timestamp, err := c.Mine()
	if err != nil {
		log.Printf("failed to autoMine with err: %s", err)
	}
	log.Printf("automine sucess block timestamp: %d", timestamp)
}
