package collection

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

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
