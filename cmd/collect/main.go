package main

import (
	"context"
	"fmt"
	"os"

	"github.com/simlecode/subspace-tool/collection"
	"github.com/simlecode/subspace-tool/models"
	"github.com/simlecode/subspace-tool/types"
	"github.com/simlecode/subspace-tool/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "collect",
		Usage: "collect subspace chain data",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "mysql",
				Usage:    "mysql url, eg. username:password@localhost:3306/database_name",
				Required: true,
			},
			&cli.Int64Flag{
				Name:  "start-height",
				Usage: "start height",
				Value: 0,
			},
			&cli.Int64Flag{
				Name:   "look-back-start-height",
				Usage:  "start height",
				Value:  0,
				Hidden: true,
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

	mysqlURL := cctx.String("mysql")
	repo, err := models.OpenMysql(mysqlURL, false)
	if err != nil {
		return err
	}

	s, err := collection.NewCollect(ctx, repo, types.DefURL, cctx.Int64("start-height"), cctx.Int64("look-back-start-height"))
	if err != nil {
		return err
	}
	s.Start(ctx)

	return nil
}
