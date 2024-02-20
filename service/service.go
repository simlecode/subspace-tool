package service

import (
	"context"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/itering/substrate-api-rpc"
	"github.com/itering/substrate-api-rpc/metadata"
	"github.com/itering/substrate-api-rpc/websocket"
	"github.com/simlecode/subspace-tool/collection"
	"github.com/simlecode/subspace-tool/config"
	"github.com/simlecode/subspace-tool/models/dao"
	"github.com/simlecode/subspace-tool/types"
)

type Service struct {
	dao dao.IDao
	cfg *config.Config
	c   *collection.Collection
}

func New(ctx context.Context, cfg *config.Config) (*Service, error) {
	websocket.SetEndpoint(cfg.NodeURL)
	d, dbStorage, err := dao.New(ctx, cfg.MysqlDsn)
	if err != nil {
		return nil, err
	}
	s := &Service{dao: d, cfg: cfg, c: collection.NewSimpleCollect(ctx, types.DefURL)}
	s.initSubRuntimeLatest()
	pluginRegister(dbStorage)
	GlobalEventDetail = newEventDetailWatcher(ctx, d, s.c)

	if err := s.c.TrackSpacePledged(ctx, s.dao); err != nil {
		return nil, fmt.Errorf("track space pledged failed: %v", err)
	}
	return s, nil
}

type SubscribeService struct {
	ctx context.Context
	*Service
	newHead    chan bool
	newFinHead chan bool
}

func (s *Service) initSubscribeService(ctx context.Context) *SubscribeService {
	return &SubscribeService{
		ctx:        ctx,
		Service:    s,
		newHead:    make(chan bool, 1),
		newFinHead: make(chan bool, 1),
	}
}

func (s *Service) initSubRuntimeLatest() {
	// reg network custom type
	defer func() {
		go s.unknownToken()
		if c, err := readTypeRegistry(s.cfg.NetworkNode); err == nil {
			substrate.RegCustomTypes(c)
			// if unknown := metadata.Decoder.CheckRegistry(); len(unknown) > 0 {
			// 	log.Printf("Found unknown type %s", strings.Join(unknown, ", "))
			// }
		} else {
			if os.Getenv("TEST_MOD") != "true" {
				panic(err)
			}
		}
	}()

	// find db
	recent := s.dao.RuntimeVersionRecent()
	if recent != nil && strings.HasPrefix(recent.RawData, "0x") {
		metadata.Latest(&metadata.RuntimeRaw{Spec: recent.SpecVersion, Raw: recent.RawData})
		return
	}
	fmt.Println("recent: ", recent)
	// find metadata for blockChain
	raw := s.regCodecMetadata()
	fmt.Println("raw: ", raw)
	if strings.HasPrefix(raw, "0x") {
		metadata.Latest(&metadata.RuntimeRaw{Spec: 1, Raw: raw})
		return
	}

	panic("can not find chain metadata, please check network")
}

func readTypeRegistry(networkNode string) ([]byte, error) {
	return typeFiles.ReadFile("source/" + networkNode + ".json")
}

//go:embed source
var typeFiles embed.FS
