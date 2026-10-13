package logic

import (
	"gf2gacha/logger"
	"gf2gacha/model"
	"gf2gacha/pb"
)

type poolInfoData struct {
	items           map[int64]*pb.ItemDataUnit
	names           map[int64]string
	upItems         map[int64]int64
	itemRanks       map[int64]map[int64]int64
	standardWeapons map[int64]bool
}

func (data poolInfoData) isLose(poolType, poolId int64, item *pb.ItemDataUnit) bool {
	switch poolType {
	case 3, 4:
		upItemId, hasUp := data.upItems[poolId]
		return hasUp && upItemId != item.GetId()
	case 6:
		if item.GetType() != 10 {
			return false
		}
		switch data.names[item.GetName().GetId()] {
		case "莫辛纳甘", "托洛洛", "绯", "琼玖", "哈卜茜", "塞布丽娜", "佩里缇亚":
			return true
		}
	case 7:
		return data.standardWeapons[item.GetId()]
	}
	return false
}

func calculatePoolInfo(poolType int64, records []model.LocalRecord, data poolInfoData) model.Pool {
	pool := model.Pool{PoolType: poolType}
	var isPreviousLose bool
	for _, storedRecord := range records {
		pool.GachaCount++
		pool.StoredCount++
		item := data.items[storedRecord.ItemId]
		itemRank := item.GetRank()
		if poolType == 8 {
			itemRank = data.itemRanks[storedRecord.PoolId][storedRecord.ItemId]
		}
		switch itemRank {
		case 5:
			lose := data.isLose(poolType, storedRecord.PoolId, item)
			// 只有歪后获得非歪五星才计作大保底，连续歪仍参与歪率统计。
			if isPreviousLose && !lose {
				pool.GuaranteesCount++
			}
			if lose {
				pool.LoseCount++
			}
			isPreviousLose = lose
			pool.RecordList = append(pool.RecordList, model.DisplayRecord{
				Id:    item.GetId(),
				Name:  data.names[item.GetName().GetId()],
				Lose:  lose,
				Count: pool.StoredCount,
			})
			pool.StoredCount = 0
			pool.Rank5Count++
		case 4:
			pool.Rank4Count++
		case 3:
			pool.Rank3Count++
		default:
			logger.Logger.Warnf("未知的物品Rank poolType:%d poolId:%d itemId:%d", poolType, storedRecord.PoolId, storedRecord.ItemId)
		}
	}
	return pool
}
