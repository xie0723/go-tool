package oplog

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/common"
	"gorm.io/gorm"
)

type OpLog struct {
	ID        uint      `gorm:"primary_key"  json:"id"`                     // 自增主键
	User      string    `gorm:"size:100;column:user;index"  json:"user"`    // 操作者的账户id
	Action    string    `gorm:"size:20;column:action;index"  json:"action"` // 行为
	FlowId    uint      `gorm:"column:flow_id;index"  json:"flowId"`        // FlowId
	Flow      string    `gorm:"size:50;column:flow;index"  json:"flow"`     // FlowType
	Remark    string    `gorm:"type:text;column:remark"  json:"remark"`     // 操作详情
	CreatedAt time.Time `json:"created_at"`                                 // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                 // 最后更新时间
}

func (OpLog) TableName() string {
	if beego.AppConfig.String("oplogTblName") != "" {
		return beego.AppConfig.String("oplogTblName")
	}
	return "op_log"
}

// 行为
const (
	DEFAULT_REMARK = "NA" // 默认的Remark值

	OP_ACTION_ADD    = "add"    // 增
	OP_ACTION_DELETE = "delete" // 删
	OP_ACTION_ALTER  = "alter"  // 改
	OP_ACTION_QUERY  = "query"  // 查
	OP_ACTION_IMPORT = "import" // 导入
)

// 记录操作日志
// AddOperationLog insert a new operation log into database and returns
// last inserted Id on success.
func AddOperationLog(dbInst *gorm.DB, oplogModel interface{}) {
	if dbInst == nil {
		panic("oplog db[gorm instance] is nil")
	}
	go func() {
		//tx := dbInst.Begin()
		err := dbInst.Create(oplogModel).Error
		if err != nil {
			logs.Error("oprate log Create error %s", err.Error())
			//tx.Rollback()
		}
		//tx.Commit()
	}()
}

// 查询操作日志
// GetAllOpLog retrieves all oplogs matches certain condition. Returns empty list if
// no records exist
func GetAllOpLog(dbInst *gorm.DB, opModel interface{}, ret interface{}, querys []*common.QueryConditon, fields []string, sortby []string, order []string,
	offset int, limit int) (ml interface{}, totalcount int64, err error) {
	g := dbInst.Model(opModel)
	for _, query := range querys {
		switch query.QueryType {
		case common.MultiSelect: // 多选
			g = g.Where(fmt.Sprintf("%s in (?)", query.QueryKey), query.QueryValues)
		case common.NumRange: // 数字范围
			if len(query.QueryValues) == 2 {
				g = g.Where(fmt.Sprintf("%s >= ? AND %s <= ?", query.QueryKey, query.QueryKey), query.QueryValues[0], query.QueryValues[1])
			} else {
				return nil, 0, fmt.Errorf("query params err %+v", query)
			}
		default:
			// case common.MultiText: // 模糊多值匹配
			sql := make([]string, len(query.QueryValues))
			val := make([]interface{}, len(query.QueryValues))
			for i := range query.QueryValues {
				sql[i] = "name LIKE ?"
				val[i] = query.QueryValues[i]
			}
			g = g.Where(strings.Join(sql, "or"), val...)
		}
	}
	if err = g.Count(&totalcount).Error; err != nil {
		logs.Error("GetAllOpLog Count error %s\n", err.Error())
		return
	}
	if len(fields) > 0 {
		g = g.Select(fields)
	}
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i := range sortby {
				g = g.Order(fmt.Sprintf("%s %s", sortby[i], order[i]))
			}

		} else if len(sortby) != len(order) && len(order) == 1 {
			for i := range sortby {
				g = g.Order(fmt.Sprintf("%s %s", sortby[i], order[0]))
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, 0, errors.New("error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, 0, errors.New("error: unused 'order' fields")
		}
	}
	if limit < 0 || limit > 100 {
		limit = 100
	}
	if err = g.Offset(offset).Limit(limit).Find(ret).Error; err != nil {
		return
	}
	return ret, totalcount, nil
}
