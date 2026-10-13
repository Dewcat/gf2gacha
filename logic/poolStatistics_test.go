package logic

import (
	"gf2gacha/model"
	"gf2gacha/pb"
	"reflect"
	"strconv"
	"testing"
)

func testPoolInfoData() poolInfoData {
	data := poolInfoData{
		items:           make(map[int64]*pb.ItemDataUnit),
		names:           make(map[int64]string),
		upItems:         map[int64]int64{3001: 1034, 5001: 10343, 90001: 1034, 91001: 10343},
		itemRanks:       map[int64]map[int64]int64{8001: {1034: 4, 300: 5}},
		standardWeapons: make(map[int64]bool),
	}
	addItem := func(id, typ, rank int64, name string) {
		data.items[id] = &pb.ItemDataUnit{Id: id, Type: typ, Rank: rank, Name: &pb.LanguageStringData{Id: id}}
		data.names[id] = name
	}
	for id, name := range map[int64]string{
		1033: "莫辛纳甘", 1025: "托洛洛", 1043: "绯", 1027: "琼玖",
		1049: "哈卜茜", 1029: "塞布丽娜", 1021: "佩里缇亚",
		1034: "玛绮朵", 1015: "维普蕾",
	} {
		addItem(id, 10, 5, name)
	}
	for id, name := range map[int64]string{
		11016: "猎心者", 11020: "光学幻境", 11038: "游星", 11044: "金石奏",
		11047: "梅扎露娜", 10333: "斯摩希克", 10433: "赫斯提亚", 10493: "二律背反",
	} {
		addItem(id, 20, 5, name)
		data.standardWeapons[id] = true
	}
	addItem(10343, 20, 5, "苦涩焦糖")
	addItem(400, 20, 4, "四星武器")
	addItem(300, 20, 3, "三星武器")
	return data
}

func poolRecords(poolId int64, itemIds ...int64) []model.LocalRecord {
	records := make([]model.LocalRecord, len(itemIds))
	for i, itemId := range itemIds {
		records[i] = model.LocalRecord{PoolId: poolId, ItemId: itemId}
	}
	return records
}

func TestSelectableCharacterLosses(t *testing.T) {
	data := testPoolInfoData()
	for _, id := range []int64{1033, 1025, 1043, 1027, 1049, 1029, 1021, 1034, 1015} {
		t.Run(data.names[id], func(t *testing.T) {
			pool := calculatePoolInfo(6, poolRecords(90001, id), data)
			wantLose := id != 1034 && id != 1015
			if pool.RecordList[0].Lose != wantLose || (pool.LoseCount == 1) != wantLose {
				t.Fatalf("incorrect loss classification: %+v", pool)
			}
		})
	}
}

func TestSelectableWeaponLosses(t *testing.T) {
	data := testPoolInfoData()
	for id := range data.standardWeapons {
		t.Run(data.names[id], func(t *testing.T) {
			pool := calculatePoolInfo(7, poolRecords(91001, id), data)
			if pool.LoseCount != 1 || !pool.RecordList[0].Lose {
				t.Fatalf("standard weapon was not counted as a loss: %+v", pool)
			}
		})
	}
	pool := calculatePoolInfo(7, poolRecords(91001, 10343, 400, 300), data)
	if pool.LoseCount != 0 || pool.Rank5Count != 1 || pool.Rank4Count != 1 || pool.Rank3Count != 1 {
		t.Fatalf("non-standard and lower-rank weapons must not count as losses: %+v", pool)
	}
}

func TestPoolLossStatistics(t *testing.T) {
	for _, tc := range []struct {
		poolType, poolId, loseItem, upItem int64
	}{
		{3, 3001, 1033, 1034},
		{4, 5001, 11016, 10343},
		{6, 90001, 1049, 1034},
		{7, 91001, 10493, 10343},
	} {
		t.Run(strconv.FormatInt(tc.poolType, 10), func(t *testing.T) {
			pool := calculatePoolInfo(tc.poolType, poolRecords(tc.poolId,
				300, tc.loseItem, 400, tc.upItem, tc.upItem, tc.loseItem, 300,
			), testPoolInfoData())
			if pool.GachaCount != 7 || pool.Rank5Count != 4 || pool.Rank4Count != 1 || pool.Rank3Count != 2 || pool.StoredCount != 1 {
				t.Fatalf("incorrect pull counts: %+v", pool)
			}
			if pool.LoseCount != 2 || pool.GuaranteesCount != 1 || pool.Rank5Count-pool.GuaranteesCount != 3 {
				t.Fatalf("expected loss rate 2/3 with one guaranteed gold excluded: %+v", pool)
			}
			var flags []bool
			var counts []int64
			for _, record := range pool.RecordList {
				flags = append(flags, record.Lose)
				counts = append(counts, record.Count)
			}
			if !reflect.DeepEqual(flags, []bool{true, false, false, true}) || !reflect.DeepEqual(counts, []int64{2, 2, 1, 1}) {
				t.Fatalf("incorrect displayed records: %+v", pool.RecordList)
			}
			consecutive := calculatePoolInfo(tc.poolType, poolRecords(tc.poolId,
				tc.loseItem, tc.loseItem, tc.upItem,
			), testPoolInfoData())
			if consecutive.LoseCount != 2 || consecutive.GuaranteesCount != 1 {
				t.Fatalf("consecutive losses must remain eligible for the loss rate: %+v", consecutive)
			}
		})
	}
}

func TestStandardAndBeginnerPoolsDoNotLose(t *testing.T) {
	data := testPoolInfoData()
	// 即使提供了 UP 配置，常驻池和新手池也不进行歪判定。
	data.upItems[1001] = 1034
	for _, poolType := range []int64{1, 5} {
		pool := calculatePoolInfo(poolType, poolRecords(1001, 1033, 1025, 1049, 11016, 10493), data)
		if pool.Rank5Count != 5 || pool.LoseCount != 0 || pool.GuaranteesCount != 0 {
			t.Fatalf("pool %d must not classify losses: %+v", poolType, pool)
		}
		for _, record := range pool.RecordList {
			if record.Lose {
				t.Fatalf("pool %d displayed a loss: %+v", poolType, record)
			}
		}
	}
}

func TestLimitedPoolUsesItsUpItem(t *testing.T) {
	data := testPoolInfoData()
	// 常驻武器作为某一期武器池的 UP 时，不应被误标为歪。
	data.upItems[5001] = 11016
	pool := calculatePoolInfo(4, poolRecords(5001, 11016, 10343), data)
	if pool.LoseCount != 1 || pool.RecordList[0].Lose || !pool.RecordList[1].Lose {
		t.Fatalf("limited weapon pool must use its actual UP item: %+v", pool)
	}
	pool = calculatePoolInfo(3, poolRecords(9999, 1033), data)
	if pool.LoseCount != 0 {
		t.Fatalf("unknown UP metadata must preserve existing behavior: %+v", pool)
	}
}

func TestMysteryBoxUsesPoolRanks(t *testing.T) {
	pool := calculatePoolInfo(8, poolRecords(8001, 1034, 300), testPoolInfoData())
	if pool.Rank5Count != 1 || pool.Rank4Count != 1 || pool.LoseCount != 0 || pool.RecordList[0].Id != 300 {
		t.Fatalf("mystery box must use pool ranks without loss classification: %+v", pool)
	}
}

func TestEmptyPool(t *testing.T) {
	pool := calculatePoolInfo(6, nil, testPoolInfoData())
	if !reflect.DeepEqual(pool, model.Pool{PoolType: 6}) {
		t.Fatalf("unexpected empty pool statistics: %+v", pool)
	}
}
