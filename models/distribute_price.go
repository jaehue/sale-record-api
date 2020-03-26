package models

import (
	"sort"

	"github.com/pangpanglabs/goutils/number"
)

type DistributeData struct {
	DistributeItemsTotalAmt float64
	DistributeItems         []DistributeItem
}

type DistributeItem struct {
	Id                 int64
	ItemCode           string
	ItemTotalSalePrice float64
	DistributeAmt      float64
}

func CalculateDistributeAmt(distributeData *DistributeData, totalDistributeAmt float64, setting *number.Setting) {
	var totalDistributeAmtForChange = totalDistributeAmt
	var distributeItems = distributeData.DistributeItems

	for i, distributeItem := range distributeItems {
		distributeItems[i].DistributeAmt = number.ToFixed(totalDistributeAmt*distributeItem.ItemTotalSalePrice/distributeData.DistributeItemsTotalAmt, setting)
		totalDistributeAmtForChange = number.ToFixed(totalDistributeAmtForChange-distributeItems[i].DistributeAmt, nil)
	}
	if totalDistributeAmtForChange > 0 {
		sort.Slice(distributeItems, func(i, j int) bool {
			return distributeItems[i].ItemTotalSalePrice > distributeItems[j].ItemTotalSalePrice
		})

		maxSalePriceItem := &distributeItems[0]
		maxSalePriceItem.DistributeAmt = number.ToFixed(maxSalePriceItem.DistributeAmt+totalDistributeAmtForChange, nil)
	} else if totalDistributeAmtForChange < 0 {
		for i, distributeItem := range distributeItems {
			if totalDistributeAmtForChange+distributeItem.ItemTotalSalePrice >= 0 {
				distributeItems[i].DistributeAmt = number.ToFixed(distributeItems[i].DistributeAmt+totalDistributeAmtForChange, nil)
				break
			}
		}
	}
}
