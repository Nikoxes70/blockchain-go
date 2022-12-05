package neighbor_nodes_syncer

import (
	"context"
	"log"
	"time"
)

type blockchainer interface {
	SyncNeighbors() (int, error)
}

type syncer struct {
	tickerTime time.Duration
	blockchainer
}

func New(d time.Duration, b blockchainer) syncer {
	return syncer{
		tickerTime:   d,
		blockchainer: b,
	}
}

func (c *syncer) Start(ctx context.Context) {
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

func (c *syncer) do() {
	nodesCount, err := c.SyncNeighbors()
	if err != nil {
		log.Printf("failed to sync neighbors nodes with err: %s", err)
	}
	log.Printf("success sync neighbors nodes: %d", nodesCount)
}
