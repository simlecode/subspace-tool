package service

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/pkg/recws"
	"github.com/itering/substrate-api-rpc/rpc"
	ws "github.com/itering/substrate-api-rpc/websocket"
)

const (
	runtimeVersion = iota + 1
	newHeader
	finalizeHeader
)

func (s *Service) Subscribe(ctx context.Context, conn ws.WsConn) {
	var err error
	subscribeSrv := s.initSubscribeService(ctx)
	go func() {
		var timeoutCount int
		for {
			if !conn.IsConnected() {
				continue
			}
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read: %s", err)
				if strings.Contains(err.Error(), "i/o timeout") {
					timeoutCount++
				}
				if timeoutCount > 100 {
					conn.Close()
					subscribeConn := &recws.RecConn{KeepAliveTimeout: 600 * time.Second, WriteTimeout: time.Second * 30, ReadTimeout: 30 * time.Second}
					subscribeConn.Dial(s.cfg.NodeURL, nil)
					conn = subscribeConn
					if err = conn.WriteMessage(websocket.TextMessage, rpc.ChainGetRuntimeVersion(runtimeVersion)); err != nil {
						log.Printf("write: %s", err)
					}
					if err = conn.WriteMessage(websocket.TextMessage, rpc.ChainSubscribeNewHead(newHeader)); err != nil {
						log.Printf("write: %s", err)
					}
					if err = conn.WriteMessage(websocket.TextMessage, rpc.ChainSubscribeFinalizedHeads(finalizeHeader)); err != nil {
						log.Printf("write: %s", err)
					}
					timeoutCount = 0
				}
				continue
			}
			_ = subscribeSrv.parser(message)
		}
	}()

	if err = conn.WriteMessage(websocket.TextMessage, rpc.ChainGetRuntimeVersion(runtimeVersion)); err != nil {
		log.Printf("write: %s", err)
	}
	if err = conn.WriteMessage(websocket.TextMessage, rpc.ChainSubscribeNewHead(newHeader)); err != nil {
		log.Printf("write: %s", err)
	}
	if err = conn.WriteMessage(websocket.TextMessage, rpc.ChainSubscribeFinalizedHeads(finalizeHeader)); err != nil {
		log.Printf("write: %s", err)
	}

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.TextMessage, rpc.SystemHealth(rand.Intn(100)+finalizeHeader)); err != nil {
				log.Printf("SystemHealth get error: %v", err)
			}
		case <-ctx.Done():
			err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("write close: %s", err)
				return
			}
			conn.Close()
			return
		}
	}

}
