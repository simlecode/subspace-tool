package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/itering/subscan/util/ss58"
	"github.com/simlecode/subspace-tool/models"
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

	timeStr := "2024-01-15T09:11:59.180000Z"

	t1, err := time.Parse("2006-01-02T15:04:05", strings.Split(timeStr, ".")[0])
	assert.NoError(t, err)
	fmt.Println(t1)
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
	for _, ed := range eds {
		pbs[ed.EventArgs.PublicKey] = struct{}{}
		rs[ed.EventArgs.RewardAddress] = struct{}{}
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

	fmt.Println("public keys:", len(pbs))
	fmt.Println("reward keys:", len(rs))
}
