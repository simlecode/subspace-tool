package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/simlecode/subspace-tool/config"
	"github.com/simlecode/subspace-tool/observer"
	"github.com/simlecode/subspace-tool/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "block-collect",
		Usage: "collect subspace chain data from node",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "mysql",
				Usage:    "mysql url, eg. username:password@localhost:3306/database_name",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "node-url",
				Usage: "node url",
				Value: "ws://127.0.0.1:9944",
			},
		},
		Action: run,
	}

	app.Version = version.Version
	app.Setup()

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Err: %v\n", err)
	}
}

func run(cctx *cli.Context) error {
	ctx, cancel := context.WithCancel(cctx.Context)
	defer cancel()

	cfg := config.DefaultConfig()
	cfg.MysqlDsn = cctx.String("mysql")
	cfg.NodeURL = cctx.String("node-url")

	sigs := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	}()

	_, err := observer.Run(ctx, cfg)
	if err != nil {
		return err
	}

	<-sigs

	return nil
}
