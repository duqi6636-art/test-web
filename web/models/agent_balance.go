package models

import "time"

type AgentBalance struct {
	Id         int `json:"id"`
	Uid        int `json:"uid"`
	Balance    int `json:"balance"`
	Total      int `json:"total"`
	Status     int `json:"status"`
	CreateTime int `json:"create_time"`
}

var agentBalanceTable = "cm_agent_balance"

// 查询用户代理商数据
func GetAgentBalanceByUid(uid int) (data AgentBalance) {
	db.Table(agentBalanceTable).Where("uid =?", uid).Find(&data)
	return
}

// 创建代理商记录
func CreateAgentBalance(data AgentBalance) (err error, id int) {
	err = db.Table(agentBalanceTable).Create(&data).Error
	return err, data.Id
}

// 修改用户代理商记录
func EditAgentBalanceByUid(uid int, params interface{}) (err error) {
	err = db.Table(agentBalanceTable).Where("uid =?", uid).Update(params).Error
	return err
}
func EditAgentBalanceBy(where interface{}, params interface{}) (err error) {
	err = db.Table(agentBalanceTable).Where(where).Update(params).Error
	return err
}

type LogAgentExchange struct {
	Id           int    `json:"id"`            // ID
	Uid          int    `json:"uid"`           // uid
	Cate         string `json:"cate"`          // 类别：cdk：cdk兑换，exchange：直接转化
	Number       int    `json:"number"`        // 兑换ip数量
	Balance      int    `json:"balance"`       // 用户余额（增加前）
	AgentBalance int    `json:"agent_balance"` // 代理商余额（扣除前）
	CreateTime   int    `json:"create_time"`   // 创建时间
}

// 修改用户代理商余额、增加用户余额并写入直接转化记录
func AgentAndUser(usersInfo Users, agentInfo AgentBalance, number int) bool {
	var err1, err2, err3 error
	tx := db.Begin()
	err1 = tx.Table(CmUserTable).Where("id = ?", usersInfo.Id).Updates(map[string]interface{}{"balance": usersInfo.Balance + number}).Error
	err2 = tx.Table(agentBalanceTable).Where("id = ?", agentInfo.Id).Updates(map[string]interface{}{"balance": agentInfo.Balance - number}).Error

	err3 = tx.Table("log_agent_exchange").Create(&LogAgentExchange{
		Uid:          usersInfo.Id,
		Cate:         "exchange",
		Number:       number,
		Balance:      usersInfo.Balance,
		AgentBalance: agentInfo.Balance,
		CreateTime:   int(time.Now().Unix()),
	}).Error
	if err1 == nil && err2 == nil && err3 == nil {
		tx.Commit()
		return true
	}
	tx.Rollback()
	return false
}

// 写入对换记录
func AddAgentExchange(usersInfo Users, agentInfo AgentBalance, number int) error {
	return db.Table("log_agent_exchange").Create(&LogAgentExchange{
		Uid:          usersInfo.Id,
		Cate:         "exchange",
		Number:       number,
		Balance:      usersInfo.Balance,
		AgentBalance: agentInfo.Balance,
		CreateTime:   int(time.Now().Unix()),
	}).Error
}


// 用户代理商记录
type StUserAgentInfo struct {
	Id           int    `json:"id"`            // ID
	Uid          int    `json:"uid"`           // uid
	CreateTime   int    `json:"create_time"`   // 创建时间
}

// 查询用户代理商数据
func GetAgentByUid(uid int) (data StUserAgentInfo) {
	db.Table("st_agent_user").Where("uid =?", uid).Find(&data)
	return
}
