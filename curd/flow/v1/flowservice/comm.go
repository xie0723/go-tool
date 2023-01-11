package flowservice

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/common"
	"gorm.io/gorm"
)

// 创建CRUD对象
func (c *CommFlow) Create(dbInst *gorm.DB, obj interface{}) (err error) {
	tx := dbInst.Begin()
	if err = tx.Create(obj).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

// 分页查询 CRUD 对象
// GetAll retrieves all oplogs matches certain condition. Returns empty list if
// no records exist
func (c *CommFlow) BaseGetAll(dbInst *gorm.DB, crudModel interface{}, ret interface{}, querys []*common.QueryConditon, fields []string, sortby []string, order []string,
	offset int, limit int) (ml interface{}, totalcount int64, err error) {
	var g *gorm.DB
	if g, err = c.BaseQuery(dbInst, crudModel, querys, fields, sortby, order); err != nil {
		return
	}
	if err = g.Count(&totalcount).Error; err != nil {
		logs.Error("Get CRUD obj Count error %s", err.Error())
		return
	}
	if limit < 0 || limit > 100 {
		limit = 100
	}
	g = g.Offset(offset).Limit(limit)

	if err = g.Find(ret).Error; err != nil {
		return
	}

	return ret, totalcount, nil
}

// 获取一个CRUD对象
func (c *CommFlow) BaseGetOne(dbInst *gorm.DB, crudModel interface{}, id int64) (ret interface{}, err error) {
	if err = dbInst.First(crudModel, id).Error; err != nil {
		logs.Error("query crud obj by id[%d] failed, %s", id, err.Error())
		return
	}
	return crudModel, nil
}

// 更新一个CRUD 对象
func (c *CommFlow) BaseUpdate(dbInst *gorm.DB, crudModel interface{}, fields []string, saveNil ...bool) (ret interface{}, err error) {
	var count int64
	var fieldsMap = map[string]struct{}{}
	if err = dbInst.Model(crudModel).Count(&count).Error; err != nil {
		return
	}
	if count == 0 {
		return nil, fmt.Errorf("record not found")
	}
	if len(fields) > 0 {
		for _, v := range fields {
			fieldsMap[v] = struct{}{}
		}
		if len(saveNil) <= 0 || !saveNil[0] {
			dbInst = dbInst.Select(fields)
		}
	}
	if len(saveNil) > 0 && saveNil[0] {
		var kvmap = map[string]interface{}{}
		// 需要存储空值
		obj := reflect.ValueOf(crudModel).Elem()
		for i := 0; i < obj.NumField(); i++ {
			if _, ok := fieldsMap[obj.Type().Field(i).Name]; len(fields) == 0 || ok {
				kvmap[obj.Type().Field(i).Name] = obj.Field(i).Interface()
			}
		}
		err = dbInst.Model(crudModel).Updates(kvmap).Error
	} else {
		err = dbInst.Model(crudModel).Updates(crudModel).Error
	}
	return crudModel, err
}

// 基础方法 ---
func (c *CommFlow) BaseQuery(dbInst *gorm.DB, crudModel interface{}, querys []*common.QueryConditon,
	fields []string, sortby []string, order []string) (o *gorm.DB, err error) {
	g := dbInst.Model(crudModel)
	for _, query := range querys {
		switch query.QueryType {
		case common.MultiSelect: // 多选
			g = g.Where(fmt.Sprintf("%s in (?)", query.QueryKey), query.QueryValues)
		case common.NumRange: // 数字范围
			if len(query.QueryValues) == 2 {
				g = g.Where(fmt.Sprintf("%s >= ? AND %s <= ?", query.QueryKey, query.QueryKey), query.QueryValues[0], query.QueryValues[1])
			} else {
				return nil, fmt.Errorf("query params err %+v", query)
			}
		case common.NotIn: //
			g = g.Not(query.QueryKey, query.QueryValues)
		case common.CommaMultiSelect:
			var whereCond []string
			var values []interface{}
			if len(query.QueryValues) > 0 {
				for _, v := range query.QueryValues {
					whereCond = append(whereCond, fmt.Sprintf("FIND_IN_SET(?,%s)", query.QueryKey))
					values = append(values, v)
				}
				g = g.Where(strings.Join(whereCond, " OR "), values...)
			}
		default:
			// case common.MultiText: // 模糊多值匹配
			sql := make([]string, len(query.QueryValues))
			val := make([]interface{}, len(query.QueryValues))
			for i := range query.QueryValues {
				sql[i] = fmt.Sprintf("%s LIKE ?", query.QueryKey)
				val[i] = fmt.Sprintf("%%%s%%", query.QueryValues[i])
			}
			g = g.Where(strings.Join(sql, " or "), val...)
		}
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
			return nil, errors.New("error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	}
	return g, nil
}

// 导入数据
// data 需要导入的数据
// head 属性表头
// dataRowIndex 从第几行开始导入
// db 数据连接
// model 实例
func (c *CommFlow) ImportData(data [][]string, head []string, db *gorm.DB,
	getModel func() interface{}, getValue func(string, string) (interface{}, error)) (err error) {
	for _, row := range data {
		model := getModel()
		s := reflect.ValueOf(model).Elem()
		for col, cell := range row {
			var v interface{}
			cell := strings.TrimSpace(cell)
			attr := strings.TrimSpace(head[col])
			if cell == "" {
				continue
			}
			if !s.FieldByName(attr).IsValid() { //不识别的属性，继续
				logs.Warning("unknow attr:", attr)
				continue
			}
			field := s.FieldByName(attr).Type().Kind()
			switch field {
			case reflect.Float64:
				v, err = strconv.ParseFloat(cell, 64)
			case reflect.Float32:
				v, err = strconv.ParseFloat(cell, 32)
			case reflect.String:
				v = cell
			case reflect.Int:
				v, err = strconv.Atoi(cell)
			case reflect.Bool:
				v, err = strconv.ParseBool(cell)
			case reflect.Int64:
				v, err = strconv.ParseInt(cell, 10, 64)
			case reflect.Ptr:
				switch s.FieldByName(attr).Type().String() {
				case "*time.Time":
					//处理时间对象
					var _format string
					if len(cell) == len("2006-01-02 15:04:05") {
						_format = "2006-01-02 15:04:05"
					} else if len(cell) == len("2006-01-02") {
						_format = "2006-01-02"
					} else {
						err = fmt.Errorf("value[%s] of %s is not time format", cell, attr)
						logs.Error(err.Error())
						return err
					}
					var vv time.Time
					if vv, err = time.Parse(_format, cell); err != nil {
						logs.Error("Parse Time.time[%s] data failed, %s", attr, err.Error())
						return err
					}
					v = &vv
				default:
					v, err = getValue(attr, cell)
				}
			case reflect.Struct:
				switch s.FieldByName(attr).Type().String() {
				case "time.Time":
					//处理时间对象
					var _format string
					if len(cell) == len("2006-01-02 15:04:05") {
						_format = "2006-01-02 15:04:05"
					} else if len(cell) == len("2006-01-02") {
						_format = "2006-01-02"
					} else {
						err = fmt.Errorf("value[%s] of %s is not time format", cell, attr)
						logs.Error(err.Error())
						return err
					}
					if v, err = time.Parse(_format, cell); err != nil {
						logs.Error("Parse Time.time[%s] data failed, %s", attr, err.Error())
						return err
					}
				default:
					v, err = getValue(attr, cell)
				}
			default:
				err = fmt.Errorf("unkown attr type: %s", field)
				logs.Error(err.Error())
				return err
			}
			if err != nil {
				logs.Error("parse cell value failed when import model[%+v], %s", model, err.Error())
				return err
			}
			refvalue := reflect.ValueOf(v)
			s.FieldByName(attr).Set(refvalue)
		}
		err = db.Create(model).Error
	}
	return
}
