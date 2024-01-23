package observer

import (
	"context"
	"time"

	"github.com/itering/substrate-api-rpc/pkg/recws"
	"github.com/simlecode/subspace-tool/config"
	"github.com/simlecode/subspace-tool/service"
)

// var (
// 	srv  *service.Service
// 	stop = make(chan struct{}, 2)
// )

func Run(ctx context.Context, cfg *config.Config) (*service.Service, error) {
	srv, err := service.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	subscribeConn := &recws.RecConn{KeepAliveTimeout: 600 * time.Second, WriteTimeout: time.Second * 30, ReadTimeout: 30 * time.Second}
	subscribeConn.Dial(cfg.NodeURL, nil)
	go srv.Subscribe(ctx, subscribeConn)

	return srv, nil

	// for {
	// 	switch dt {
	// 	case "substrate":
	// 		subscribeConn := &recws.RecConn{KeepAliveTimeout: 10 * time.Second, WriteTimeout: time.Second * 5, ReadTimeout: 10 * time.Second}
	// 		subscribeConn.Dial(util.WSEndPoint, nil)
	// 		go srv.Subscribe(ctx, subscribeConn)
	// 	default:
	// 		log.Fatalf("no such daemon component: %s", dt)
	// 	}
	// 	enableTermSignalHandler()
	// 	if _, ok := <-stop; ok {
	// 		time.Sleep(3 * time.Second)
	// 		break
	// 	}
	// }
}

// func enableTermSignalHandler() {
// 	sigs := make(chan os.Signal, 1)
// 	signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
// 	go func() {
// 		log.Printf("Received signal %s, exiting...\n", <-sigs)
// 		stop <- struct{}{}
// 	}()
// }
