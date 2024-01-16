# subspace-tool

subspace tool

## collection

从 [subspace 浏览器](https://explorer.subspace.network/#/gemini-3g/consensus) 获取链数据，再存储到数据库，主要包含 `block`，`event` 和 `extrinsic` 三种数据。
`block` 数据存储在 `blocks` 表，`event` 数据存储在 `events` 表，`extrinsic` 数据存储在 `extrinsics` 表

### build

```
make collect
```

### run

> --start-height 用于设置从哪个高度开始查询链数据

```
./collect --mysql "username:password@(127.0.0.1:3306)/database?parseTime=true&loc=Local" --start-height 1100043
```

### 统计数据

**奖励包含区块奖励和 vote 奖励。**

1. 查询某段时间的出块情况
```
SELECT * FROM blocks WHERE timestamp >= '2024-01-08 00:00:00' && timestamp <= '2024-01-09 00:00:00'
```

1. 查询某段时间内 vote reward 奖励数量

```
SELECT * FROM events WHERE block_height >= 1110135 AND block_height <= 1110335 AND name = 'Rewards.VoteReward';
```
