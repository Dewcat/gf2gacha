package logic

import (
	"gf2gacha/model"
	"gf2gacha/preload"
	"github.com/pkg/errors"
)

func GetPoolInfo(uid string, poolType int64) (model.Pool, error) {
	localRecordList, err := GetLocalRecord(uid, poolType, 0)
	if err != nil {
		return model.Pool{}, errors.WithStack(err)
	}

	return calculatePoolInfo(poolType, localRecordList, poolInfoData{
		items:           preload.ItemMap,
		names:           preload.LangMap,
		upItems:         preload.UpItemMap,
		itemRanks:       preload.ItemRankMap,
		standardWeapons: preload.StandardWeaponMap,
	}), nil
}
