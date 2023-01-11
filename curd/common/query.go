package common

// 查询条件常量
const (
	MultiSelect      = "multi-select"       // 多选
	MultiText        = "multi-text"         // 模糊多选
	NumRange         = "num-range"          // 数字范围或者数字查询
	NotIn            = "not-in"             // 不在范围内
	Range            = "range"              // 日期区间的查询
	CommaMultiSelect = "comma-multi-select" //数据库中存放是逗号分隔的值
)

// QueryConditon 表格查询条件对象
type QueryConditon struct {
	QueryKey    string
	QueryType   string // multi-select  multi-text  num-range  comma-multi-select
	QueryValues []string
}

func QueryKeyReplace(query []*QueryConditon, repMap map[string]string) (ret []*QueryConditon) {
	for i := range query {
		if v, ok := repMap[query[i].QueryKey]; ok {
			query[i].QueryKey = v
		}
	}
	return query
}
