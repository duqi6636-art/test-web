package crons

import (
	"api-360proxy/web/controller"
	"api-360proxy/web/models"
)

func StaticRegionStatusWarning() {
	ok, _, stock := controller.StaticZtRegionStockList()
	if ok {
		snList := make([]string, 0, len(stock))
		for sn := range stock {
			snList = append(snList, sn)
		}
		_ = models.SetStaticRegionStatusBySnList(snList, 1)
	}
}
