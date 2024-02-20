package collection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/simlecode/subspace-tool/models"
	"github.com/simlecode/subspace-tool/ss58"
	"github.com/simlecode/subspace-tool/types"
)

const (
	interval         = time.Millisecond * 500
	lookBackInterval = time.Second * 1
)

type Collection struct {
	repo                models.Repo
	client              *http.Client
	url                 string
	startHeight         int64
	lookBackStartHeight int64
}

func NewSimpleCollect(ctx context.Context, url string) *Collection {
	return &Collection{
		client: http.DefaultClient,
		url:    url,
	}
}

func NewCollect(ctx context.Context, repo models.Repo, url string, startHeight int64, lookBackStartHeight int64) (*Collection, error) {
	ss := &Collection{
		repo:                repo,
		client:              http.DefaultClient,
		url:                 url,
		startHeight:         startHeight,
		lookBackStartHeight: lookBackStartHeight,
	}
	es, err := ss.repo.ExtrinsicRepo().List(ctx, 10)
	if err != nil {
		return nil, err
	}
	if len(es) > 0 {
		blockHeight, err := strconv.ParseInt(es[0].Node.Block.Height, 10, 64)
		if err != nil {
			return nil, err
		}
		if blockHeight > ss.startHeight {
			ss.startHeight = blockHeight
		}
	}

	return ss, nil
}

func (s *Collection) Start(ctx context.Context) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	lookBackTicker := time.NewTicker(lookBackInterval)
	defer ticker.Stop()

	// go s.trackEventDetail(ctx)
	go s.TrackSpacePledged(ctx, s.repo.SpaceRepo())

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			blockDetailStart := time.Now()
			blkInfo, err := s.queryByBlockDetailHeight(ctx, s.startHeight)
			if err != nil {
				log.Println("query block detail failed:", err)
				if strings.Contains(err.Error(), "not found") {
					ticker.Reset(2 * time.Second)
					time.Sleep(2 * time.Second)
				}
				continue
			}
			blockDetailTook := time.Since(blockDetailStart)

			if err := s.repo.BlockRepo().SaveBlock(ctx, blkInfo.blk); err != nil {
				log.Println("save block failed: ", err)
				continue
			}
			for _, e := range blkInfo.extrinsics {
				if err := s.repo.ExtrinsicRepo().SaveExtrinsic(ctx, &e); err != nil {
					log.Println("save extrinsic failed:", err)
					continue
				}
			}

			for _, e := range blkInfo.events {
				if err := s.repo.EventRepo().SaveEvent(ctx, &e); err != nil {
					log.Println("save event failed:", err)
					continue
				}
			}

			eventDetailStart := time.Now()
			var wg sync.WaitGroup
			var err2 error
			control := make(chan struct{}, 10)
			for _, e := range blkInfo.events {
				if e.Node.Name == types.EventSubspaceFarmerVote {
					wg.Add(1)
					id := e.Node.ID
					control <- struct{}{}

					go func(id string) {
						defer func() {
							wg.Done()
							<-control
						}()

						eventDetail, err := s.QueryEventByID(ctx, id)
						if err != nil {
							log.Printf("query event detail failed, id: %v, err: %v\n", id, err)
							err2 = fmt.Errorf("query event detail failed, id: %v, err: %v", id, err)
						} else {
							if err := s.repo.EventDetailRepo().SaveEventDetail(ctx, eventDetail); err != nil {
								log.Println("save event detail failed:", err)
								err2 = fmt.Errorf("save event detail failed: %s", err)
							}
						}
					}(id)
				}
				if e.Node.Name == types.EventSubspaceBlockReward {
					id := e.Node.ID
					eventDetail, err := s.QueryEventByID(ctx, id)
					if err != nil {
						log.Printf("query event detail failed, id: %v, err: %v\n", id, err)
						err2 = fmt.Errorf("query event detail failed, id: %v, err: %v", id, err)
					} else {
						if strings.HasPrefix(blkInfo.blk.Author.ID, "st") {
							eventDetail.EventArgs.PublicKey = "0x" + ss58.Decode(blkInfo.blk.Author.ID, ss58.SubspaceAddressType)
						}
						eventDetail.EventArgs.RewardAddress = eventDetail.EventArgs.BlockAuthor
						eventDetail.EventArgs.Height, _ = strconv.ParseInt(blkInfo.blk.Height, 10, 64)
						eventDetail.EventArgs.ParentHash = blkInfo.blk.ParentHash

						if err := s.repo.EventDetailRepo().SaveEventDetail(ctx, eventDetail); err != nil {
							log.Println("save event detail failed:", err)
							err2 = fmt.Errorf("save event detail failed: %s", err)
						}
					}
				}
			}
			wg.Wait()
			close(control)
			eventDetailTook := time.Since(eventDetailStart)

			if err2 != nil {
				continue
			}

			s.startHeight++

			log.Printf("current block height: %d, block took: %v, event detail: %v\n", s.startHeight, blockDetailTook, eventDetailTook)

		case <-lookBackTicker.C:

		}
	}
}

func (s *Collection) queryAndSaveEventDeatail(ctx context.Context, eventID string, farmer string, blockHeight int64, parentHash string) error {
	eventDetail, err := s.QueryEventByID(ctx, eventID)
	if err != nil {
		log.Printf("query event detail failed, id: %v, err: %v\n", eventID, err)
		return fmt.Errorf("query event detail failed, id: %v, err: %v", eventID, err)
	}
	if eventDetail.Name == types.EventSubspaceBlockReward {
		if strings.HasPrefix(farmer, "st") {
			eventDetail.EventArgs.PublicKey = "0x" + ss58.Decode(farmer, ss58.SubspaceAddressType)
		}
		eventDetail.EventArgs.RewardAddress = eventDetail.EventArgs.BlockAuthor
		eventDetail.EventArgs.Height = blockHeight
		eventDetail.EventArgs.ParentHash = parentHash

		if err := s.repo.EventDetailRepo().SaveEventDetail(ctx, eventDetail); err != nil {
			log.Println("save event detail failed:", err)
			return fmt.Errorf("save event detail failed: %s", err)
		}
	} else if eventDetail.Name == types.EventSubspaceFarmerVote {
		if err := s.repo.EventDetailRepo().SaveEventDetail(ctx, eventDetail); err != nil {
			log.Println("save event detail failed:", err)
			return fmt.Errorf("save event detail failed: %s", err)
		}
	}

	return nil
}

func (s *Collection) trackEventDetail(ctx context.Context) {
	// eventDetails, err := s.repo.EventDetailRepo().List(ctx)
	// if err != nil {
	// 	log.Println("list event details failed:", err)
	// 	return
	// }
	// if len(eventDetails) != 0 {
	// 	return
	// }

	events, err := s.repo.EventRepo().List(ctx, types.EventSubspaceFarmerVote)
	if err != nil {
		log.Println("list events failed:", err)
		return
	}

	events2, err := s.repo.EventRepo().List(ctx, types.EventSubspaceBlockReward)
	if err != nil {
		log.Println("list events failed:", err)
		return
	}

	sort.Slice(events2, func(i, j int) bool {
		return events2[i].Node.Block.Height < events2[j].Node.Block.Height
	})

	sort.Slice(events, func(i, j int) bool {
		return events[i].Node.Block.Height < events[j].Node.Block.Height
	})

	fmt.Println("events:", len(events), len(events2))

	type e struct {
		EventType   string
		BlockHeight string
		ID          string
	}
	es := make([]e, 0, len(events)+len(events2))
	for _, one := range events {
		es = append(es, e{
			EventType:   one.Node.Name,
			ID:          one.Node.ID,
			BlockHeight: one.Node.Block.Height,
		})
	}
	for _, one := range events2 {
		es = append(es, e{
			ID:          one.Node.ID,
			BlockHeight: one.Node.Block.Height,
			EventType:   one.Node.Name,
		})
	}

	var wg sync.WaitGroup
	receive := make(chan e, 100)

	go func() {
		control := make(chan struct{}, 50)
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-receive:
				control <- struct{}{}

				func() {
					defer func() {
						wg.Done()
						<-control
					}()

					blockHeight, err := strconv.ParseInt(e.BlockHeight, 10, 64)
					if err != nil {
						log.Println("parse block height failed:", err)
						receive <- e
					}
					if blockHeight%10 == 0 {
						log.Println("block height:", blockHeight)
					}
					blkInfo, err := s.repo.BlockRepo().ByBlockHeight(ctx, int(blockHeight))
					if err != nil {
						log.Println("get block info failed:", err)
						receive <- e
					}
					if err := s.queryAndSaveEventDeatail(ctx, e.ID, blkInfo.Author.ID, blockHeight, blkInfo.ParentHash); err != nil {
						log.Println("query and save event detail failed:", err)
						receive <- e
					}
				}()
			}
		}
	}()

	for _, e := range es {
		// if e.EventType != types.EventSubspaceFarmerVote && e.EventType != types.EventSubspaceBlockReward {
		// 	continue
		// }

		if e.EventType != types.EventSubspaceBlockReward {
			continue
		}

		wg.Add(1)

		i, _ := strconv.ParseInt(e.BlockHeight, 10, 64)
		if i < 230600 {
			continue
		}

		receive <- e
	}

	wg.Wait()
}

func (s *Collection) TrackSpacePledged(ctx context.Context, r models.SpaceRepo) error {
	spaces, err := r.ListSapce()
	if err != nil {
		return err
	}
	sort.Slice(spaces, func(i, j int) bool {
		return spaces[i].Timestamp > spaces[j].Timestamp
	})

	var maxSpace *models.Space
	if len(spaces) > 0 {
		maxSpace = &spaces[0]
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		adjustIntervalSecs := int64(2016 * 6)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var changed bool
				space, err := s.querySpacePledged(ctx)
				if err != nil {
					log.Println("query and save space pledged failed:", err)
				} else {
					if maxSpace == nil {
						changed = true
					} else if space.Timestamp-maxSpace.Timestamp >= adjustIntervalSecs {
						changed = true
					} else if space.Pledged != maxSpace.Pledged {
						changed = true
					}
				}
				if changed {
					maxSpace = space
					if err := r.SaveSpace(maxSpace); err != nil {
						log.Println("save space pledged failed:", err)
					} else {
						log.Println("save space pledged:", maxSpace)
					}
				}
			}
		}
	}()

	return nil
}

type blkInfo struct {
	blk        *types.BlockInfo
	extrinsics []types.Event
	events     []types.Event
}

func (s *Collection) queryByBlockDetailHeight(ctx context.Context, blockHeight int64) (*blkInfo, error) {
	var wg sync.WaitGroup
	wg.Add(3)

	var info *types.BlockInfo
	var extrinsics *types.ExtrinsicsConnection
	var events *types.EventsConnection
	var blkErr, extrinsicErr, eventErr error

	go func() {
		defer wg.Done()
		info, blkErr = s.QueryBlock(ctx, blockHeight)
	}()

	go func() {
		defer wg.Done()
		extrinsics, extrinsicErr = s.QueryExtrinsic(ctx, blockHeight)
	}()

	go func() {
		defer wg.Done()
		events, eventErr = s.QueryEvent(ctx, blockHeight)
	}()

	wg.Wait()

	if blkErr != nil || extrinsicErr != nil || eventErr != nil {
		return nil, fmt.Errorf("query block: %w, extrinsic: %w, event: %w", blkErr, extrinsicErr, eventErr)
	}

	return &blkInfo{info, extrinsics.Edges, events.Edges}, nil
}

func (s *Collection) QueryBlock(ctx context.Context, blockID int64) (*types.BlockInfo, error) {
	reqParams := &types.Req{
		OperationName: types.OpBlockById,
		Variables: types.Variables{
			BlockID: blockID,
		},
		Query: types.BlockQuery,
	}

	data, err := json.Marshal(reqParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r types.Resp
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}

	if len(r.Data.Blocks) == 0 {
		return nil, fmt.Errorf("block not found")
	}

	return &r.Data.Blocks[0], nil
}

func (s *Collection) QueryEvent(ctx context.Context, blockID int64) (*types.EventsConnection, error) {
	reqParams := &types.Req{
		OperationName: types.OpEventsByBlockId,
		Variables: types.Variables{
			BlockID: blockID,
			First:   100,
		},
		Query: types.EventQuery,
	}

	data, err := json.Marshal(reqParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r types.Resp
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}

	return &r.Data.EventsConnection, nil
}

func (s *Collection) QueryExtrinsic(ctx context.Context, blockID int64) (*types.ExtrinsicsConnection, error) {
	reqParams := &types.Req{
		OperationName: types.OpExtrinsicsByBlockId,
		Variables: types.Variables{
			BlockID: blockID, // block height
			First:   100,
		},
		Query: types.ExtrinsicQuery,
	}

	data, err := json.Marshal(reqParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r types.Resp
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}

	return &r.Data.ExtrinsicsConnection, nil
}

func (s *Collection) QueryEventByID(ctx context.Context, eventID string) (*types.EventDetail, error) {
	reqParams := &types.Req{
		OperationName: types.OpEventById,
		Variables: types.Variables{
			EventId: eventID,
		},
		Query: types.EventByIdQuery,
	}

	data, err := json.Marshal(reqParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r types.Resp
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}

	return &r.Data.EventDetail, nil
}

func (s *Collection) querySpacePledged(ctx context.Context) (*models.Space, error) {
	reqParams := &types.Req{
		OperationName: types.OpHomeQuery,
		Variables: types.Variables{
			Limit:        10,
			Offset:       0,
			AccountTotal: "00",
		},
		Query: types.HomeQuery,
	}

	data, err := json.Marshal(reqParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r types.Resp
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}
	space := &models.Space{}
	if len(r.Data.Blocks) > 0 {
		space.Pledged, err = strconv.ParseInt(r.Data.Blocks[0].SpacePledged, 10, 64)
		if err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02T15:04:05", strings.Split(r.Data.Blocks[0].Timestamp, ".")[0])
		if err != nil {
			return nil, err
		}
		space.Timestamp = t.Unix()
	}

	return space, nil
}
