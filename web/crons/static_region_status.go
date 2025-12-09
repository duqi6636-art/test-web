package crons

import (
	"api-360proxy/controller"
	"api-360proxy/models"
)

func StaticRegionStatusWarning() {
	ok, _, stock := controller.StaticZtRegionStockList()
	if ok {
		snList := make([]string, 0, len(stock))
		for sn := range stock {
			snList = append(snList, sn)
		}
		models.UpStaticRegionStatusAndIpNumberByStock(stock)
		models.SetStaticRegionStatusByNotInSnList(snList, 2)
	}
}
