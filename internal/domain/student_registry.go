package domain

// MaxStudentRegistryCSVImportRows 学员总表 CSV 数据行上限（不含表头）
const MaxStudentRegistryCSVImportRows = 5000

// StudentRegistryCSVHeaders 与「学员信息总表」导出列顺序一致（UTF-8 CSV，首行为表头）
var StudentRegistryCSVHeaders = []string{
	"学员编号",
	"姓名",
	"责任教员",
	"港澳外",
	"服务状态",
	"中断原因",
	"转场",
	"两周",
	"三个月结束",
	"工作日飞行",
	"上午飞行",
	"体验次数",
	"体验剩余",
	"视频记录",
	"定制拍摄",
	"拍摄剩余",
	"总合同小时",
	"证件号码",
	"出生日期",
	"出生地",
	"性别",
	"执照档案号",
	"手机号",
	"学历",
	"永久地址",
	"通讯地址",
}
