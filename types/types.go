package types

const (
	DefURL = "https://squid.gemini-3g.subspace.network/graphql"

	EventQuery     = "query EventsByBlockId($blockId: BigInt!, $first: Int!, $after: String) {\n  eventsConnection(\n    orderBy: indexInBlock_ASC\n    first: $first\n    after: $after\n    where: {block: {height_eq: $blockId}}\n  ) {\n    edges {\n      node {\n        id\n        name\n        phase\n        indexInBlock\n        block {\n          height\n          id\n          __typename\n        }\n        extrinsic {\n          indexInBlock\n          block {\n            height\n            id\n            __typename\n          }\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    totalCount\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n      __typename\n    }\n    __typename\n  }\n}"
	ExtrinsicQuery = "query ExtrinsicsByBlockId($blockId: BigInt!, $first: Int!, $after: String) {\n  extrinsicsConnection(\n    orderBy: indexInBlock_ASC\n    first: $first\n    after: $after\n    where: {block: {height_eq: $blockId}}\n  ) {\n    edges {\n      node {\n        id\n        hash\n        name\n        success\n        block {\n          height\n          timestamp\n          __typename\n        }\n        indexInBlock\n        __typename\n      }\n      cursor\n      __typename\n    }\n    totalCount\n    pageInfo {\n      hasNextPage\n      endCursor\n      hasPreviousPage\n      startCursor\n      __typename\n    }\n    __typename\n  }\n}"
	BlockQuery     = "query BlockById($blockId: BigInt!) {\n  blocks(limit: 10, where: {height_eq: $blockId}) {\n    id\n    height\n    hash\n    stateRoot\n    timestamp\n    extrinsicsRoot\n    specId\n    parentHash\n    extrinsicsCount\n    eventsCount\n    logs(limit: 10, orderBy: block_height_DESC) {\n      block {\n        height\n        timestamp\n        __typename\n      }\n      kind\n      id\n      __typename\n    }\n    author {\n      id\n      __typename\n    }\n    __typename\n  }\n}"
	EventByIdQuery = "query EventById($eventId: String!) {\n  eventById(id: $eventId) {\n    args\n    id\n    indexInBlock\n    name\n    phase\n    timestamp\n    call {\n      args\n      name\n      success\n      timestamp\n      id\n      __typename\n    }\n    extrinsic {\n      args\n      success\n      tip\n      fee\n      id\n      signer {\n        id\n        __typename\n      }\n      __typename\n    }\n    block {\n      height\n      id\n      timestamp\n      specId\n      hash\n      __typename\n    }\n    __typename\n  }\n}"
)

const (
	OpExtrinsicsByBlockId = "ExtrinsicsByBlockId"
	OpEventsByBlockId     = "EventsByBlockId"
	OpBlockById           = "BlockById"
	OpEventById           = "EventById"
)

const (
	EventSubspaceFarmerVote = "Subspace.FarmerVote"
)

type Req struct {
	OperationName string    `json:"operationName"`
	Variables     Variables `json:"variables"`
	Query         string    `json:"query"`
}

type Variables struct {
	BlockID int64  `json:"blockId"`
	First   int    `json:"first"`
	EventId string `json:"eventId"`
}

type Resp struct {
	Data Data `json:"data"`
}

type Data struct {
	EventsConnection     EventsConnection     `json:"eventsConnection"`
	ExtrinsicsConnection ExtrinsicsConnection `json:"extrinsicsConnection"`
	Blocks               []BlockInfo          `json:"blocks"`
	EventDetail          EventDetail          `json:"eventById"`
}

/*
"id": "0001107843-614b9",
"height": "1107843",
"hash": "0x614b9af48696be5379051ac7c58d7afdaa1cf021d8222ce2634b0a6e961ca791",
"stateRoot": "0x4dba3b88c4c7bb8d7dbf1ab21d5c604fc00f5cd152ac7ae8906a48ecbdbe5e32",
"timestamp": "2024-01-15T09:11:59.180000Z",
"extrinsicsRoot": "0x104805c18e9d0f00ee0508adf29b8c89b4065723c53f01faa3fc8439710dc7f4",
"specId": "subspace@5",
"parentHash": "0x42a18b7bff96cf0d08dff2fb3f7f3a530eb798e748727d4e9705ad1a6023d441",
"extrinsicsCount": 8,
"eventsCount": 44,
"author": st8eJ9cuh4XsHyoqWNWr13o8e9SiqYvX2Yg7cSKVKQy6KeUCN
*/
type BlockInfo struct {
	ID              string `json:"id"`
	Author          Author `json:"author"`
	Height          string `json:"height"`
	Hash            string `json:"hash"`
	StateRoot       string `json:"stateRoot"`
	Timestamp       string `json:"timestamp"`
	ExtrinsicsRoot  string `json:"extrinsicsRoot"`
	SpecId          string `json:"specId"`
	ParentHash      string `json:"parentHash"`
	ExtrinsicsCount int    `json:"extrinsicsCount"`
	EventsCount     int    `json:"eventsCount"`
	TypeName        string `json:"__typename"`
}

type Author struct {
	ID       string `json:"id"`
	TypeName string `json:"__typename"`
}

type EventsConnection struct {
	Edges      []Event  `json:"edges"`
	TotalCount int      `json:"totalCount"`
	PageInfo   PageInfo `json:"pageInfo"`
	TypeName   string   `json:"__typename"`
}

type ExtrinsicsConnection struct {
	Edges      []Event  `json:"edges"`
	TotalCount int      `json:"totalCount"`
	PageInfo   PageInfo `json:"pageInfo"`
	TypeName   string   `json:"__typename"`
}

type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
	HasPrevious bool   `json:"hasPreviousPage"`
	StartCursor string `json:"startCursor"`
	TypeName    string `json:"__typename"`
}

type Event struct {
	Node     Node   `json:"node"`
	TypeName string `json:"__typename"`

	// Extrinsics
	Cursor string `json:"cursor"`
}

type Node struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IndexInBlock int    `json:"indexInBlock"`
	Block        Block  `json:"block"`
	TypeName     string `json:"__typename"`

	// event
	Phase     string    `json:"phase"`
	Extrinsic Extrinsic `json:"extrinsic"`

	// Extrinsics
	Hash    string `json:"hash"`
	Success bool   `json:"success"`
}

type Block struct {
	Height   string `json:"height"`
	TypeName string `json:"__typename"`

	// event
	ID string `json:"id"`

	// Extrinsics
	Timestamp string `json:"timestamp"`
}

type Extrinsic struct {
	IndexInBlock int    `json:"indexInBlock"`
	Block        Block  `json:"block"`
	TypeName     string `json:"__typename"`
}

type EventDetail struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	EventArgs EventArgs `json:"args"`
	Timestamp string    `json:"timestamp"`
	TypeName  string    `json:"__typename"`
}

//	"args": {
//		"height": 1120563,
//		"publicKey": "0x7483f122c69ed7ef3f8aad34a06de88381dc498b7a22f40732ff83cc0c25e40e",
//		"parentHash": "0x08422c0705c18a604849f2302e301566e3c3c34e9081e352b070f3a5247a12da",
//		"rewardAddress": "0x5c49626b1912124a5a83e174fc01e3f423d08a4c0a70fbb8c0e953ddfdaffd68"
//	}
type EventArgs struct {
	Height        int64  `json:"height"`
	PublicKey     string `json:"publicKey"`
	ParentHash    string `json:"parentHash"`
	RewardAddress string `json:"rewardAddress"`
}
