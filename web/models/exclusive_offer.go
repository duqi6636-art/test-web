package models
// 专属优惠配置

// 专属
type ExclusiveOffer struct {
	Id          	int       `json:"id"`
	Uid          	int       `json:"uid"`
	Code      		string    `json:"code"`			//专属邀请码
	Ratio 			float64   `json:"ratio"`		//专属赠送比例
	Img 			string    `json:"img"`			//专属图片
	Money 			float64   `json:"money"`		//折扣金额点
	DiscountLt 		float64   `json:"discount_lt"`		// < money 的折扣
	Discount 		float64   `json:"discount"`		// >= money 时的折扣
}
// 专属返回
type ResExclusiveOffer struct {
	Code      		string    `json:"code"`			//专属邀请码
	Ratio 			float64   `json:"ratio"`		//专属赠送比例
	Percent 		string    `json:"percent"`		//专属赠送比例
	Img 			string    `json:"img"`			//专属图片
	Money 			float64   `json:"money"`		//折扣金额点
	Discount 		string   `json:"discount"`		//
}
// 获取信息
func GetExclusiveByCode(code string)(data ExclusiveOffer)  {
	dbRead.Table("cm_exclusive_offer").Where("code = ?",code).Where("status = ?",1).Order("id desc").First(&data)
	return
}
// 获取信息
func GetExclusiveOfferBy()(data[] ExclusiveOffer)  {
	dbRead.Table("cm_exclusive_offer").Where("status =?",1).Order("id desc").Find(&data)
	return
}
