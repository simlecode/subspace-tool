package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/simlecode/subspace-tool/models"
	"github.com/simlecode/subspace-tool/ss58"
	"github.com/simlecode/subspace-tool/types"
	"github.com/stretchr/testify/assert"
)

func TestCollect(t *testing.T) {
	queryData := `{"operationName":"BlockById","variables":{"blockId":1107843},"query":"query BlockById($blockId: BigInt!) {\n  blocks(limit: 10, where: {height_eq: $blockId}) {\n    id\n    height\n    hash\n    stateRoot\n    timestamp\n    extrinsicsRoot\n    specId\n    parentHash\n    extrinsicsCount\n    eventsCount\n    logs(limit: 10, orderBy: block_height_DESC) {\n      block {\n        height\n        timestamp\n        __typename\n      }\n      kind\n      id\n      __typename\n    }\n    author {\n      id\n      __typename\n    }\n    __typename\n  }\n}"}`
	r := types.Req{}

	err := json.Unmarshal([]byte(queryData), &r)
	assert.NoError(t, err)

	d, err := json.Marshal(r)
	assert.NoError(t, err)
	fmt.Println(string(d))

	timeStr := "2024-02-15T09:11:59.180000Z"

	t1, err := time.Parse("2006-01-02T15:04:05", strings.Split(timeStr, ".")[0])
	assert.NoError(t, err)
	fmt.Println(t1)

	oneDay := 60 * 60 * 24

	days := time.Now().Unix() / int64(oneDay)
	fmt.Println("days:", days)

	fmt.Println("t1:", t1.Unix()/int64(oneDay))
	fmt.Println("now:", time.Unix(days*int64(oneDay), 0).String())
}

type SimpleBlock struct {
	Author    string
	Hash      string
	Height    int
	Timestamp time.Time
}

func TestStat(t *testing.T) {
	// st7ctEPDYyzydLQaEWXZpr1jYHxsHFW3QVm5vpkWCdRtyhdb8
	in := "0x3c04cb0139a5eae6994fc406c864d825b6e6a2d487205cbb4ff459954441dfae"
	addr := ss58.Encode(in, 2254)
	fmt.Println("addr:", addr)
	// return

	mysqlURL := "admin:_Admin123@(127.0.0.1:3306)/subspace_3h_collect?parseTime=true&loc=Local"
	repo, err := models.OpenMysql(mysqlURL, false)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blks, err := repo.BlockRepo().ListBlock(ctx)
	assert.NoError(t, err)
	fmt.Println("blocks:", len(blks))

	var sBlks []*SimpleBlock
	dayBlks := make(map[string][]*SimpleBlock)
	for _, blk := range blks {
		sb := &SimpleBlock{
			Author: blk.Author.ID,
			Hash:   blk.Hash,
		}

		sb.Height, err = strconv.Atoi(blk.Height)
		assert.NoError(t, err)
		sb.Timestamp, err = time.Parse("2006-01-02T15:04:05", blk.Timestamp)
		assert.NoError(t, err)
		sBlks = append(sBlks, sb)
	}

	sort.Slice(sBlks, func(i, j int) bool {
		return sBlks[i].Height > sBlks[j].Height
	})

	for _, blk := range sBlks {
		str := blk.Timestamp.Format("2006-01-02")
		dayBlks[str] = append(dayBlks[str], blk)
	}

	authors := make(map[string]struct{})
	for _, b := range blks {
		authors[b.Author.ID] = struct{}{}
	}
	fmt.Println("authors:", len(authors))

	for k := range authors {
		fmt.Println("author:", k)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()

	eds, err := repo.EventDetailRepo().List(ctx)
	assert.NoError(t, err)
	fmt.Println("event details:", len(eds))

	pbs := make(map[string]struct{})
	rs := make(map[string]struct{})
	// rewardPublckKeys := make(map[string]map[string]int)
	for _, ed := range eds {
		pbs[ed.EventArgs.PublicKey] = struct{}{}
		rs[ed.EventArgs.RewardAddress] = struct{}{}

		// if _, ok := rewardPublckKeys[ed.EventArgs.RewardAddress]; !ok {
		// 	rewardPublckKeys[ed.EventArgs.RewardAddress] = make(map[string]int)
		// }
		// rewardPublckKeys[ed.EventArgs.RewardAddress][ed.EventArgs.PublicKey]++
	}

	for k := range pbs {
		fmt.Println("public key:", k)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()

	for k := range rs {
		fmt.Println("reward key:", k)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()

	fmt.Println("所有 farmer 数量:", len(pbs))
	fmt.Println("所有 reward 数量:", len(rs))

	fmt.Println()
	fmt.Println()
	fmt.Println()

	days := 3
	oneDayHeight := 14400 * days
	var oneDayEds []*types.EventDetail

	sort.Slice(eds, func(i, j int) bool {
		return eds[i].EventArgs.Height > eds[j].EventArgs.Height
	})
	maxHeight := eds[0].EventArgs.Height
	for _, ed := range eds {
		if ed.EventArgs.Height >= maxHeight-int64(oneDayHeight) {
			oneDayEds = append(oneDayEds, ed)
		}
	}
	// fmt.Println("one day event details:", len(oneDayEds), "min height", maxHeight-int64(oneDayHeight))

	rewardPublckKeys := make(map[string]map[string]int)
	for _, ed := range oneDayEds {
		pbs[ed.EventArgs.PublicKey] = struct{}{}
		rs[ed.EventArgs.RewardAddress] = struct{}{}

		if _, ok := rewardPublckKeys[ed.EventArgs.RewardAddress]; !ok {
			rewardPublckKeys[ed.EventArgs.RewardAddress] = make(map[string]int)
		}
		rewardPublckKeys[ed.EventArgs.RewardAddress][ed.EventArgs.PublicKey]++
	}

	rpks := []RewardPublicKey{}
	for reward, pks := range rewardPublckKeys {
		rpk := RewardPublicKey{RewardAddress: reward}
		for pk, count := range pks {
			rpk.PublicKeyNums = append(rpk.PublicKeyNums, PublicKeyNum{PublicKey: pk, Count: count})
		}
		rpks = append(rpks, rpk)
	}

	sort.Slice(rpks, func(i, j int) bool {
		return len(rpks[i].PublicKeyNums) > len(rpks[j].PublicKeyNums)
	})

	var topRewardFarmers int
	var topRewards int
	for i, rpk := range rpks {
		if i >= 100 {
			break
		}
		fmt.Println("reward address:", rpk.RewardAddress, "farmer count:", len(rpk.PublicKeyNums))
		for _, pkn := range rpk.PublicKeyNums {
			// fmt.Println("public key:", pkn.PublicKey, "count:", pkn.Count)
			topRewards += pkn.Count
		}
		topRewardFarmers += len(rpk.PublicKeyNums)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println(days, "天", "vote奖励数量：", len(oneDayEds))
	fmt.Println(days, "天", "前100奖励地址总奖励数量：", topRewards, "奖励数量占比：", float64(topRewards)/float64(len(oneDayEds)))
	fmt.Println(days, "天", "前100奖励地址包含farmer数量：", topRewardFarmers)
	fmt.Println()
	fmt.Println()
	fmt.Println()
}

type RewardPublicKey struct {
	RewardAddress string
	TotalReward   int
	PublicKeyNums []PublicKeyNum
}

type PublicKeyNum struct {
	PublicKey string
	Count     int
}

func TestStat2(t *testing.T) {
	mysqlURL := "admin:_Admin123@(127.0.0.1:3306)/subspace_3h_collect?parseTime=true&loc=Local"
	repo, err := models.OpenMysql(mysqlURL, false)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blks, err := repo.BlockRepo().ListBlock(ctx)
	assert.NoError(t, err)
	fmt.Println("blocks:", len(blks))

	authors := make(map[string]struct{})
	for _, b := range blks {
		authors[b.Author.ID] = struct{}{}
	}
	fmt.Println("authors:", len(authors))

	// for k := range authors {
	// 	fmt.Println("author:", k)
	// }
	fmt.Println()
	fmt.Println()
	fmt.Println()

	eds, err := repo.EventDetailRepo().List(ctx)
	assert.NoError(t, err)
	fmt.Println("event details:", len(eds))

	pbs := make(map[string]struct{})
	rs := make(map[string]struct{})
	for _, ed := range eds {
		pbs[ed.EventArgs.PublicKey] = struct{}{}
		rs[ed.EventArgs.RewardAddress] = struct{}{}
	}

	// for k := range pbs {
	// 	fmt.Println("public key:", k)
	// }
	fmt.Println()
	fmt.Println()
	fmt.Println()

	// for k := range rs {
	// 	fmt.Println("reward key:", k)
	// }
	fmt.Println()
	fmt.Println()
	fmt.Println()

	totalFarmer := len(pbs)
	totalReward := len(rs)
	fmt.Println("所有 farmer 数量:", totalFarmer)
	fmt.Println("所有 reward 数量:", totalReward)

	sort.Slice(eds, func(i, j int) bool {
		return eds[i].EventArgs.Height > eds[j].EventArgs.Height
	})
	maxHeight := eds[0].EventArgs.Height
	days := 7
	oldPbs := make(map[string]int)
	// oldRs := make(map[string]struct{})
	for i := days; i >= 0; i-- {
		height := 14400 * i
		// var tmpEds []*types.EventDetail

		pbs := make(map[string]int)
		rs := make(map[string]struct{})
		for _, ed := range eds {
			if ed.EventArgs.Height <= maxHeight-int64(height) {
				// tmpEds = append(tmpEds, ed)
				pbs[ed.EventArgs.PublicKey]++
				rs[ed.EventArgs.RewardAddress] = struct{}{}
			}
		}
		newFarmers := make(map[string]int)
		for k, v := range pbs {
			newFarmers[k] = v
		}

		newFarmerRewardTotal := 0
		if len(oldPbs) > 0 {
			for k := range pbs {
				if _, ok := oldPbs[k]; ok {
					delete(newFarmers, k)
				} else {
					newFarmerRewardTotal += pbs[k]
				}
			}
			newFarmerAvgRewards := float64(newFarmerRewardTotal) / float64(len(newFarmers))
			fmt.Println(i, "天前", "farmer 数量:", len(pbs), ", reward 数量:", len(rs),
				", 新farmer数量", len(newFarmers), ", 新farmer平均奖励：", newFarmerAvgRewards)
		} else {
			fmt.Println(i, "天前", "farmer 数量:", len(pbs), ", reward 数量:", len(rs))
		}

		oldPbs = pbs
		// oldRs = rs
	}
}

func TestTrackEventDetails(t *testing.T) {
	mysqlURL := "admin:_Admin123@(127.0.0.1:3306)/subspace_3h_collect?parseTime=true&loc=Local"
	repo, err := models.OpenMysql(mysqlURL, false)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := NewCollect(ctx, repo, types.DefURL, 0, 0)
	assert.NoError(t, err)
	c.trackEventDetail(ctx)
}

func TestTrackSpacePledged(t *testing.T) {
	mysqlURL := "admin:_Admin123@(127.0.0.1:3306)/subspace_3h_collect?parseTime=true&loc=Local"
	repo, err := models.OpenMysql(mysqlURL, false)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := NewCollect(ctx, repo, types.DefURL, 0, 0)
	assert.NoError(t, err)
	c.TrackSpacePledged(ctx, repo.SpaceRepo())

	time.Sleep(time.Minute * 3)
}
