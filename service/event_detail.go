package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/simlecode/subspace-tool/collection"
	"github.com/simlecode/subspace-tool/models/dao"
	"github.com/simlecode/subspace-tool/ss58"
	"github.com/simlecode/subspace-tool/types"
)

var GlobalEventDetail *eventDetailWatcher

type eventDetailWatcher struct {
	dao      dao.IDao
	receiver chan int
	c        *collection.Collection
}

func newEventDetailWatcher(ctx context.Context, dao dao.IDao, c *collection.Collection) *eventDetailWatcher {
	w := &eventDetailWatcher{
		receiver: make(chan int, 20),
		dao:      dao,
		c:        c,
	}

	go w.Start(ctx)
	return w
}

func (w *eventDetailWatcher) Add(blkNum int) {
	w.receiver <- blkNum
}

func (w *eventDetailWatcher) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case blkNum := <-w.receiver:
			if err := w.createEventDetail(blkNum); err != nil {
				fmt.Printf("create event detail at %d failed: %v \n", blkNum, err)
				w.Add(blkNum)
			} else {
				fmt.Println("create event detail success:", blkNum)
			}
		}
	}
}

func (w *eventDetailWatcher) createEventDetail(blkNum int) error {
	blk := w.dao.GetBlockByNum(blkNum)
	if blk == nil {
		return fmt.Errorf("blk is nil")
	}
	events := w.dao.GetEventByBlockNum(blkNum)
	if events == nil {
		return fmt.Errorf("events is nil")
	}

	var eds []*dao.EventDetail
	blkRewardEventDetail := &dao.EventDetail{
		ID:          fmt.Sprintf("%d-1", blkNum),
		Name:        types.EventSubspaceBlockReward,
		BlockHeight: blkNum,
		ParentHash:  blk.ParentHash,
	}

	blkInfo2, err := w.c.QueryBlock(context.Background(), int64(blkNum))
	if err != nil {
		return fmt.Errorf("query block failed: %v", err)
	}

	if strings.HasPrefix(blkInfo2.Author.ID, "st") {
		blkRewardEventDetail.PublicKey = "0x" + ss58.Decode(blkInfo2.Author.ID, ss58.SubspaceAddressType)
	}

	start := 3
	for idx, e := range events {
		if e.EventId == "BlockReward" {
			var params []EventJSONData
			err := json.Unmarshal([]byte(e.Params), &params)
			if err != nil {
				return fmt.Errorf("unmarshal event block reward params error: %v", err)
			}
			for _, p := range params {
				if p.Name == "block_author" {
					blkRewardEventDetail.RewardAddress = p.Value.(string)
					break
				}
			}
			eds = append(eds, blkRewardEventDetail)
		}
		if e.EventId == "FarmerVote" {
			var params []EventJSONData
			err := json.Unmarshal([]byte(e.Params), &params)
			if err != nil {
				return fmt.Errorf("unmarshal event(%d) farmer vote params error: %v", idx, err)
			}
			ed := dao.EventDetail{
				ID:          fmt.Sprintf("%d-%d", blkNum, start),
				Name:        types.EventSubspaceFarmerVote,
				BlockHeight: blkNum,
			}
			for _, p := range params {
				if p.Name == "public_key" {
					ed.PublicKey = p.Value.(string)
				}
				if p.Name == "reward_address" {
					ed.RewardAddress = p.Value.(string)
				}
				if p.Name == "parent_hash" {
					ed.ParentHash = p.Value.(string)
				}
			}
			eds = append(eds, &ed)
			start += 2
		}
	}

	for _, ed := range eds {
		err := w.dao.CreateEventDetail(nil, ed)
		if err != nil {
			return err
		}
	}

	return nil
}

//// event

// vote
// [{
//   "name": "public_key",
//   "type": "[U8; 32]",
//   "type_name": "FarmerPublicKey",
//   "value": "0xda57fd931741b19590359c867fa3d122f66e22649e987ecdef1c523654adcf55"
// }, {
//   "name": "reward_address",
//   "type": "[U8; 32]",
//   "type_name": "AccountId",
//   "value": "0x4ecc0ee03bcca0cea9f7f2180bae5964eb80b29d38b6fa010e0fe45ba7e1a264"
// }, {
//   "name": "height",
//   "type": "U32",
//   "type_name": "BlockNumberFor",
//   "value": 1160591
// }, {
//   "name": "parent_hash",
//   "type": "H256",
//   "type_name": "Hash",
//   "value": "0xa14e31c39d0869bcfa6032ae45596ca54266d504cccbe99f416231c323a287f0"
// }]

// block reward
// [{
//   "name": "block_author",
//   "type": "[U8; 32]",
//   "type_name": "AccountId",
//   "value": "0x005ed3cb9967d03e49430b302c8fc37540748e161e90fde908083b418759b732"
// }, {
//   "name": "reward",
//   "type": "U128",
//   "type_name": "BalanceOf",
//   "value": "100000000000000000"
// }]

type EventJSONData struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	TypeName string `json:"type_name"`
	Value    any    `json:"value"`
}

type JSONData struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	TypeName string `json:"type_name"`
	Value    Value  `json:"value"`
}
type Solution struct {
	Chunk            string `json:"chunk"`
	ChunkWitness     string `json:"chunk_witness"`
	HistorySize      int    `json:"history_size"`
	PieceOffset      int    `json:"piece_offset"`
	ProofOfSpace     string `json:"proof_of_space"`
	PublicKey        string `json:"public_key"`
	RecordCommitment string `json:"record_commitment"`
	RecordWitness    string `json:"record_witness"`
	RewardAddress    string `json:"reward_address"`
	SectorIndex      int    `json:"sector_index"`
}
type V0 struct {
	FutureProofOfTime string   `json:"future_proof_of_time"`
	Height            int      `json:"height"`
	ParentHash        string   `json:"parent_hash"`
	ProofOfTime       string   `json:"proof_of_time"`
	Slot              int      `json:"slot"`
	Solution          Solution `json:"solution"`
}
type Vote struct {
	V0 V0 `json:"V0"`
}
type Value struct {
	Signature string `json:"signature"`
	Vote      Vote   `json:"vote"`
}
