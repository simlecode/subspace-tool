package collection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/simlecode/subspace-tool/models"
	"github.com/simlecode/subspace-tool/types"
)

const (
	interval         = time.Millisecond * 500
	lookBackInterval = time.Second * 1
)

type Collection struct {
	repo                models.Repo
	url                 string
	startHeight         int64
	lookBackStartHeight int64
}

func NewCollect(ctx context.Context, repo models.Repo, url string, startHeight int64, lookBackStartHeight int64) (*Collection, error) {
	ss := &Collection{
		repo:                repo,
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

	go s.trackEventDetail(ctx)

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
					ticker.Reset(8 * time.Second)
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

func (s *Collection) trackEventDetail(ctx context.Context) {
	eventDetails, err := s.repo.EventDetailRepo().List(ctx)
	if err != nil {
		log.Println("list event details failed:", err)
		return
	}
	if len(eventDetails) != 0 {
		return
	}

	events, err := s.repo.EventRepo().List(ctx, types.EventSubspaceFarmerVote)
	if err != nil {
		log.Println("list events failed:", err)
		return
	}

	receive := make(chan string, 100)

	go func() {
		control := make(chan struct{}, 10)
		for {
			select {
			case <-ctx.Done():
				return
			case id := <-receive:
				control <- struct{}{}

				eventDetail, err := s.QueryEventByID(ctx, id)
				if err != nil {
					log.Println("query event detail failed:", err)
					receive <- id
				} else {
					if err := s.repo.EventDetailRepo().SaveEventDetail(ctx, eventDetail); err != nil {
						log.Println("save event detail failed:", err)
						receive <- id
					}
					if eventDetail.EventArgs.Height%10 == 0 {
						log.Println("event height:", eventDetail.EventArgs.Height)
					}
				}
				<-control
			}
		}
	}()

	// var height int64
	for _, e := range events {
		// if e.Node.Name != types.EventSubspaceFarmerVote {
		// 	continue
		// }

		// i, _ := strconv.ParseInt(e.Node.Block.Height, 10, 64)
		// if i < 1114210 || i > 1120000 {
		// 	continue
		// }

		receive <- e.Node.ID

		// eventDetail, err := s.QueryEventByID(ctx, e.Node.ID)
		// if err != nil {
		// 	log.Printf("query event(%s) detail failed: %v\n", e.Node.ID, err)
		// 	continue
		// }

		// if err := s.repo.EventDetailRepo().SaveEventDetail(ctx, eventDetail); err != nil {
		// 	log.Println("save event detail failed:", err)
		// }
		// if height < eventDetail.EventArgs.Height {
		// 	height = eventDetail.EventArgs.Height
		// 	if height%10 == 0 {
		// 		log.Printf("query event detail at height: %d\n", height)
		// 	}
		// }
	}
}

type blkInfo struct {
	blk        *types.BlockInfo
	extrinsics []types.Event
	events     []types.Event
}

func (s *Collection) queryByBlockDetailHeight(ctx context.Context, blockHeight int64) (*blkInfo, error) {
	info, err := s.QueryBlock(ctx, blockHeight)
	if err != nil {
		return nil, fmt.Errorf("query block: %w", err)
	}

	extrinsics, err := s.QueryExtrinsic(ctx, blockHeight)
	if err != nil {
		return nil, fmt.Errorf("query extrinsic: %w", err)
	}

	events, err := s.QueryEvent(ctx, blockHeight)
	if err != nil {
		return nil, fmt.Errorf("query event: %w", err)
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
	resp, err := http.DefaultClient.Do(req)
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
	resp, err := http.DefaultClient.Do(req)
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
	resp, err := http.DefaultClient.Do(req)
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
	resp, err := http.DefaultClient.Do(req)
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
