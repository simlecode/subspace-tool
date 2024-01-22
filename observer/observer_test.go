package observer

import (
	"context"
	"testing"
	"time"

	"github.com/simlecode/subspace-tool/config"
	"github.com/stretchr/testify/assert"
)

func TestObserver(t *testing.T) {
	cfg := &config.Config{
		MysqlDsn:    "admin:_Admin123@(127.0.0.1:3306)/subspace?parseTime=true&loc=Local",
		NodeURL:     "ws://127.0.0.1:9944",
		NetworkNode: "polkadot",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := Run(ctx, cfg)
	assert.NoError(t, err)

	time.Sleep(time.Minute * 10)
}
