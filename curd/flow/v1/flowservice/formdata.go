package flowservice

const (
	UISettingTypeTable   = "tableInit"
	UISettingTypeAddForm = "add"
	UISettingTypeModForm = "mod"

	UIGroupTypeSlot = "slot"
)

// 定义数据接口，给前端解析

// 分页
type ConfigPage struct {
	Type   string        `json:"type"`   // 定义使用对前端模版类型
	Label  string        `json:"label"`  // 名称，需要资源化对标签
	Groups []ConfigGroup `json:"groups"` // 分组
}

// 分组
type ConfigGroup struct {
	Label  string        `json:"label"`            // 名称，需要资源化对标签
	Key    string        `json:"key"`              // 唯一标示，用来对应slot插槽
	Type   string        `json:"type"`             // 分组类型
	Cols   int           `json:"col"`              // form中item一行显示几个
	Items  []ConfigItem  `json:"items,omitempty"`  // 配置项集合
	Tables []ConfigTable `json:"tables,omitempty"` // 表类型配置项
}

type ConfigTable struct {
	Columns    []ConfigItem             `json:"columns"`       // 列类型
	Value      []map[string]interface{} `json:"value"`         // 值
	RowModAble bool                     `json:"rowModAble"`    // 表格是否允许添加和删除行
	Key        string                   `json:"key,omitempty"` // 唯一标示，和itemkey等都不能重复
}

// 一个config item标示一个配置项
type ConfigItem struct {
	Label     string                   `json:"label"`              // 名称，需要资源化对标签
	Key       string                   `json:"key"`                // 需要传递给后台对form key
	Type      string                   `json:"type"`               // 类型：int，string，enum
	Value     interface{}              `json:"value"`              // 配置项的值
	ReadOnly  bool                     `json:"readOnly"`           // 是否只读
	Required  bool                     `json:"required"`           // 是否非必填
	Multiple  bool                     `json:"multiple"`           // 是否多选
	Clearable bool                     `json:"clearable"`          // 是否可以清空
	Columns   []map[string]interface{} `json:"columns"`            //
	Vkey      string                   `json:"vkey"`               //
	LabelKey  string                   `json:"labelkey"`           //
	Hidden    bool                     `json:"hidden"`             // 是否隐藏
	Lua       string                   `json:"lua"`                // lua脚本名，只针对enum类型有效
	RemoteUrl string                   `json:"remoteUrl"`          // 远程调用接口
	Validate  []map[string]interface{} `json:"validate,omitempty"` // 校验规则，正则表达式
	Options   []map[string]interface{} `json:"options,omitempty"`  // 可选结果
	Props     map[string]interface{}   `json:"props,omitempty"`    // 扩展属性（Key，value）
	Extra     Extra                    `json:"extra,omitempty"`    // 扩展属性，特定结构
	Control   []ItemControl            `json:"control,omitempty"`  // 联动控制
}

//
type ItemControl struct {
	Value   interface{}  `json:"value,omitempty"`
	Append  string       `json:"append,omitempty"`
	Prepend string       `json:"prepend,omitempty"`
	Rule    []ConfigItem `json:"rule,omitempty"`
}

// curd table init data 对象
type CurdTblData struct {
	DataUrl             string                 `json:"dataUrl"`
	ShowOperationColumn bool                   `json:"showOperationColumn"`
	Multidelete         bool                   `json:"multiDelete"`
	DeleteObject        map[string]interface{} `json:"deleteObject"`
	AddObject           ActionObject           `json:"addObject"`
	EditObject          ActionObject           `json:"editObject"`
	DetailObject        map[string]interface{} `json:"detailObject"`
	ImportAction        map[string]interface{} `json:"importAction"`
	ExportAction        string                 `json:"exportAction"`
	Columns             []TblColumn            `json:"columns"`
}

// 列属性（curd表的列）
type TblColumn struct {
	Label         string                   `json:"label"`                   // 列标题
	FixedPosition string                   `json:"fixedPosition,omitempty"` // 对齐方式
	Prop          string                   `json:"prop"`                    // 列属性名
	Sortable      bool                     `json:"sortable,omitempty"`      // 是否支持排序
	Queryable     bool                     `json:"queryable,omitempty"`     // 是否支持查询
	Vmap          map[string]interface{}   `json:"vmap,omitempty"`          // 显示替换值map
	Filter        []map[string]interface{} `json:"filter,omitempty"`        // 过滤
	Type          string                   `json:"type,omitempty"`          // 类型
	Width         int                      `json:"width,omitempty"`         // 列宽
}

type ActionObject struct {
	Showbtn   bool                   `json:"showbtn"`
	Formrules map[string]interface{} `json:"formrules,omitempty"`
}

type Extra struct {
	Columns []map[string]interface{} `json:"columns,omitempty"`
}
