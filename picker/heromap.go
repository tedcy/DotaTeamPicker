package picker

var HeroNameMap = map[string]int{
	"亚巴顿":    102,
	"炼金术士":   73,
	"远古冰魂":   68,
	"敌法师":    1,
	"天穹守望者":  113,
	"斧王":     2,
	"痛苦之源":   3,
	"蝙蝠骑士":   65,
	"兽王":     38,
	"血魔":     4,
	"赏金猎人":   62,
	"酒仙":     78,
	"钢背兽":    99,
	"育母蜘蛛":   61,
	"半人马站行者": 96,
	"混沌骑士":   81,
	"陈":      66,
	"克林克玆":   56,
	"发条地精":   51,
	"水晶室女":   5,
	"黑暗贤者":   55,
	"戴泽":     50,
	"死亡先知":   43,
	"萨尔":     87,
	"末日使者":   69,
	"龙骑士":    49,
	"卓尔游侠":   6,
	"大地之灵":   107,
	"撼地者":    7,
	"上古巨神":   103,
	"灰烬之灵":   106,
	"魅惑魔女":   58,
	"谜团":     33,
	"虚空假面":   41,
	"矮人直升机":  72,
	"哈斯卡":    59,
	"祈求者":    74,
	"艾欧":     91,
	"杰奇洛":    64,
	"主宰":     8,
	"光之守卫":   90,
	"昆卡":     23,
	"军团指挥官":  104,
	"拉席克":    52,
	"巫妖":     31,
	"噬魂鬼":    54,
	"莉娜":     25,
	"莱恩":     26,
	"德鲁伊":    80,
	"露娜":     48,
	"狼人":     77,
	"马格纳斯":   97,
	"美杜莎":    94,
	"米波":     82,
	"米拉娜":    9,
	"变体精灵":   10,
	"娜迦海妖":   89,
	"先知":     53,
	"瘟疫法师":   36,
	"暗夜魔王":   60,
	"司夜刺客":   88,
	"食人魔魔法师": 84,
	"全能骑士":   57,
	"神谕者":    111,
	"殁境神蚀者":  76,
	"幻影刺客":   44,
	"幻影长矛手":  12,
	"凤凰":     110,
	"帕克":     13,
	"帕吉":     14,
	"帕格纳":    45,
	"痛苦女王":   39,
	"剃刀":     15,
	"力丸":     32,
	"拉比克":    86,
	"沙王":     16,
	"暗影恶魔":   79,
	"影魔":     11,
	"暗影萨满":   27,
	"沉默术士":   75,
	"天怒法师":   101,
	"斯拉达":    28,
	"斯拉克":    93,
	"狙击手":    35,
	"幽鬼":     67,
	"裂魂人":    71,
	"风暴之灵":   17,
	"斯温":     18,
	"工程师":    105,
	"圣堂刺客":   46,
	"恐怖利刃":   109,
	"潮汐猎人":   29,
	"伐木机":    98,
	"修补匠":    34,
	"小小":     19,
	"树精卫士":   83,
	"巨魔战将":   95,
	"巨牙海民":   100,
	"不朽尸王":   85,
	"熊战士":    70,
	"复仇之魂":   20,
	"剧毒术士":   40,
	"冥界亚龙":   47,
	"维萨吉":    92,
	"术士":     37,
	"编织者":    63,
	"风行者":    21,
	"寒冬飞龙":   112,
	"巫医":     30,
	"冥魂大帝":   42,
	"宙斯":     22,
}

var HeroIdMap map[int]string

func InitHeroIdMap() {
	HeroIdMap = make(map[int]string)
	for heroName, heroId := range HeroNameMap {
		HeroIdMap[heroId] = heroName
	}
}

