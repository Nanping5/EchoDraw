package model

// Lexicon 共享词表: 规则引擎和 LLM prompt 都用, 保持两端认知一致。
// 这里只放"硬事实" (颜色 hex / 形状类型), 不放语义规则。

// ColorHex 中文/英文颜色名 → hex。匹配时大小写不敏感。
var ColorHex = map[string]string{
	"红":   "#e53935", "红色": "#e53935", "red": "#e53935",
	"蓝":   "#1e88e5", "蓝色": "#1e88e5", "blue": "#1e88e5",
	"绿":   "#43a047", "绿色": "#43a047", "green": "#43a047",
	"黄":   "#fdd835", "黄色": "#fdd835", "yellow": "#fdd835",
	"黑":   "#212121", "黑色": "#212121", "black": "#212121",
	"白":   "#ffffff", "白色": "#ffffff", "white": "#ffffff",
	"灰":   "#9e9e9e", "灰色": "#9e9e9e", "gray": "#9e9e9e", "grey": "#9e9e9e",
	"紫":   "#8e24aa", "紫色": "#8e24aa", "purple": "#8e24aa",
	"橙":   "#fb8c00", "橙色": "#fb8c00", "orange": "#fb8c00",
	"粉":   "#ec407a", "粉色": "#ec407a", "pink": "#ec407a",
	"棕":   "#6d4c41", "棕色": "#6d4c41", "brown": "#6d4c41",
	"天蓝": "#4fc3f7", "skyblue": "#4fc3f7",
	"深蓝": "#0d1b2a", "darkblue": "#0d1b2a",
}

// ShapeWords 中文形状词 → (类型, 默认尺寸)
// 默认尺寸作为"中等"基准, 大/小修饰词会按比例调整。
var ShapeWords = map[string]struct {
	Type   ShapeType
	Size   float64
}{
	"圆":     {ShapeCircle, 60}, "圆形": {ShapeCircle, 60},
	"矩形":   {ShapeRect, 100}, "方块": {ShapeRect, 100}, "方形": {ShapeRect, 100}, "长方形": {ShapeRect, 120},
	"线":     {ShapeLine, 0}, "直线": {ShapeLine, 0},
	"椭圆":   {ShapeEllipse, 80},
	"三角":   {ShapeTriangle, 70}, "三角形": {ShapeTriangle, 70},
	"星":     {ShapeStar, 60}, "星星": {ShapeStar, 60}, "五角星": {ShapeStar, 60},
	"箭头":   {ShapeArrow, 100},
}

// Position 位置词 → 画布坐标 (相对 0-1 比例, 由调用方乘以 canvasW/canvasH)
var Position = map[string]struct{ X, Y float64 }{
	"中间":  {0.5, 0.5}, "中心":  {0.5, 0.5}, "中央":  {0.5, 0.5},
	"左上":  {0.2, 0.2}, "右上":  {0.8, 0.2},
	"左下":  {0.2, 0.8}, "右下":  {0.8, 0.8},
	"左边":  {0.2, 0.5}, "左侧":  {0.2, 0.5},
	"右边":  {0.8, 0.5}, "右侧":  {0.8, 0.5},
	"上边":  {0.5, 0.2}, "上面":  {0.5, 0.2}, "上方":  {0.5, 0.2},
	"下边":  {0.5, 0.8}, "下面":  {0.5, 0.8}, "下方":  {0.5, 0.8},
}

// CreateTriggers 创建触发词 (画/加/放/绘制 等)
var CreateTriggers = []string{"画", "绘制", "加", "放", "添加", "新增", "来个", "来一个", "给我", "再来一个"}

// DeltaTriggers 增量触发词 (再/又/加一个)
var DeltaTriggers = []string{"再", "又", "再来", "额外", "继续"}

// RedrawTriggers 重画触发词 (改成/变成 + 形状词)
var RedrawTriggers = []string{"改", "改成", "换成", "变成", "变为"}

// SystemCommands 系统级命令
var SystemCommands = map[string]CommandType{
	"撤销": CmdUndo, "撤回": CmdUndo, "回退": CmdUndo,
	"重做": CmdRedo, "恢复": CmdRedo,
	"清空": CmdClear, "清除": CmdClear, "清空画布": CmdClear,
	"保存": CmdExport, "导出": CmdExport, "下载": CmdExport,
}

// RefWords 代词 → 解析规则
var RefWords = map[string]string{
	"它":       "last",
	"这个":     "last",
	"那个":     "last",
	"刚才":     "last",
	"刚才的":   "last",
	"最近":     "last",
	"最新":     "last",
	"最后的":   "last",
	"最后那个": "last",
	"首个":     "first",
	"第一个":   "first",
	"所有":     "all",
	"全部":     "all",
	"每一个":   "all",
}
