package models

const (
	VALID   int = 0
	INVALID int = 1

	SUCCESS int = 0
	ERROR   int = 1
	INFO    int = 3
	WARNING int = 20
	EXISTED int = 21
	WAITING int = 22
	ENDING  int = 23

	MODULE_DEFAULT = 1
	MODULE_VIDEO   = 2
	MODULE_PICTURE = 4
	MODULE_AUDIO   = 8
	MODULE_NOVEL   = 16

	DATA_INIT  = 0 // 初始化数据
	DATA_TURE  = 1
	DATA_FALSE = 0
)

type TypeValue struct {
	Name  string `json:"name"`
	Value int    `jon:"value"`
}

var (
	// 类型MAP
	ModuleMap = map[string]int{
		"视频": MODULE_VIDEO,
		// "图片": MODULE_PICTURE,
		// "音频": MODULE_AUDIO,
		// "小说": MODULE_NOVEL,
	}
	// 所有类型数组
	ModuleValueArr = []TypeValue{
		{"视频", MODULE_VIDEO},
		// {"图片", MODULE_PICTURE},
		// {"音频", MODULE_AUDIO},
		// {"小说", MODULE_NOVEL},
	}
)
