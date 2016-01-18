package main

import (
	"net/http"
	"io/ioutil"
	"net"
	"strings"
	"time"
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"fmt"
	"strconv"
	"sort"
)

var getMatchDetails = "https://api.steampowered.com/IDOTA2Match_570/GetMatchDetails/v001/?"
var getMatchHistory = "https://api.steampowered.com/IDOTA2Match_570/GetMatchHistory/V001/?"

var HeroNameMap = map[string]int {
		"亚巴顿": 102,
		"炼金术士": 73,
		"远古冰魂": 68,
		"敌法师": 1,
		"天穹守望者": 113,
		"斧王": 2,
		"痛苦之源": 3,
		"蝙蝠骑士": 65,
		"兽王": 38,
		"血魔": 4,
		"赏金猎人": 62,
		"酒仙": 78,
		"钢背兽": 99,
		"育母蜘蛛": 61,
		"半人马站行者": 96,
		"混沌骑士": 81,
		"陈": 66,
		"克林克玆": 56,
		"发条地精": 51,
		"水晶室女": 5,
		"黑暗贤者": 55,
		"戴泽": 50,
		"死亡先知": 43,
		"萨尔": 87,
		"末日使者": 69,
		"龙骑士": 49,
		"卓尔游侠": 6,
		"大地之灵": 107,
		"撼地者": 7,
		"上古巨神": 103,
		"灰烬之灵": 106,
		"魅惑魔女": 58,
		"谜团": 33,
		"虚空假面": 41,
		"矮人直升机": 72,
		"哈斯卡": 59,
		"祈求者": 74,
		"艾欧": 91,
		"杰奇洛": 64,
		"主宰": 8,
		"光之守卫": 90,
		"昆卡": 23,
		"军团指挥官": 104,
		"拉席克": 52,
		"巫妖": 31,
		"噬魂鬼": 54,
		"莉娜": 25,
		"莱恩": 26,
		"德鲁伊": 80,
		"露娜": 48,
		"狼人": 77,
		"马格纳斯": 97,
		"美杜莎": 94,
		"米波": 82,
		"米拉娜": 9,
		"变体精灵": 10,
		"娜迦海妖": 89,
		"先知": 53,
		"瘟疫法师": 36,
		"暗夜魔王": 60,
		"司夜刺客": 88,
		"食人魔魔法师": 84,
		"全能骑士": 57,
		"神谕者": 111,
		"殁境神蚀者": 76,
		"幻影刺客": 44,
		"幻影长矛手": 12,
		"凤凰": 110,
		"帕克": 13,
		"帕吉": 14,
		"帕格纳": 45,
		"痛苦女王": 39,
		"剃刀": 15,
		"力丸": 32,
		"拉比克": 86,
		"沙王": 16,
		"暗影恶魔": 79,
		"影魔": 11,
		"暗影萨满": 27,
		"沉默术士": 75,
		"天怒法师": 101,
		"斯拉达": 28,
		"斯拉克": 93,
		"狙击手": 35,
		"幽鬼": 67,
		"裂魂人": 71,
		"风暴之灵": 17,
		"斯温": 18,
		"工程师": 105,
		"圣堂刺客": 46,
		"恐怖利刃": 109,
		"潮汐猎人": 29,
		"伐木机": 98,
		"修补匠": 34,
		"小小": 19,
		"树精卫士": 83,
		"巨魔战将": 95,
		"巨牙海民": 100,
		"不朽尸王": 85,
		"熊战士": 70,
		"复仇之魂": 20,
		"剧毒术士": 40,
		"冥界亚龙": 47,
		"维萨吉": 92,
		"术士": 37,
		"编织者": 63,
		"风行者": 21,
		"寒冬飞龙": 112,
		"巫医": 30,
		"冥魂大帝": 42,
		"宙斯": 22,
}

var HeroIdMap map[int]string

func initHeroIdMap() {
	HeroIdMap = make(map[int]string)
	for heroName, heroId := range HeroNameMap {
		HeroIdMap[heroId] = heroName
    }
}

type MatchDetails struct {
	Result struct {
		Players [10] struct {
			AccountId int	`json:"account_id"`                                                              
			PlayerSlot int `json:"player_slot"`                                        
			HeroId int `json:"hero_id"`                                                                    
		}`json:"players"`                                                            
		RadiantWin bool `json:"radiant_win"`                                                             
	} `json:"result"`
}

type MatchHistory struct {
	Result struct {
		Status int
		NumResults int `json:"num_results"`
		Matches []struct {
			MatchId int `json:"match_id"`
		} `json:"matches"`
	} `json:"result"`
}

type PlayerInfo struct {
	MatchCount int
	MaxMatchId int
	HeroWins map[string]int
	HeroCounts map[string]int
	HeroBeatWins map[string]int
	HeroBeatCounts map[string]int
}

func (p *PlayerInfo) Init() {
	p.HeroWins = make(map[string]int)
	p.HeroCounts = make(map[string]int)
	p.HeroBeatWins = make(map[string]int)
	p.HeroBeatCounts = make(map[string]int)
}

type apiServer struct {
	Version string
	Compile string
	Players map[string]*PlayerInfo
	overview []PlayerOverview
}

type PlayerOverview struct {
	AccountId string
	Players PlayerInfo
}

func (s *apiServer) Save() {
	if len(s.overview) != 0 {
		s.overview = make([]PlayerOverview,0)
    }
	for name, playerInfo := range s.Players {
		var tmp PlayerOverview
		tmp.AccountId = name
		tmp.Players = *playerInfo
		s.overview = append(s.overview, tmp)
    }
	data, _ := json.MarshalIndent(s.overview,"", "    ")
	//os.Remove("overview.data")
	ioutil.WriteFile("overview.data", data, 0666)
	//var overview []PlayerOverview
	//json.Unmarshal(data, &overview)
	//fmt.Println(overview)
}

func (s *apiServer) Load() {
	s.Players = make(map[string]*PlayerInfo)

	data, err := ioutil.ReadFile("overview.data")
	if err != nil {
		fmt.Println(err)
		return
    }
	json.Unmarshal(data, &s.overview)
	for _, overview := range s.overview {
		p := overview.Players
		s.Players[overview.AccountId] = &p
    }
}


func httpGet(url string) ([]byte){
	fmt.Println("DEBUG-URL" + url)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func MapSort(p map[string]float32) PairList{
  pl := make(PairList, len(p))
  i := 0
  for k, v := range p {
    pl[i] = Pair{k, v}
    i++
  }
  sort.Sort(sort.Reverse(pl))
  return pl
}

type Pair struct {
  Key string
  Value float32
}

type PairList []Pair

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

func (s *apiServer) teamPick(params martini.Params) (int ,string ){
	heroListStr := params["herolist"]
	var show string
	heroList := strings.Split(heroListStr,"-")
	if len(heroList) == 0 {
		return 200 ,"NoHero"
	}
	
	for _, overview := range s.overview {
		heroBeatWinRate := make(map[string]map[string]float32)
		heroWinRate := make(map[string]float32) 
		for name, counts := range overview.Players.HeroCounts{			
			winRate := float32(overview.Players.HeroWins[name]) / float32(counts)
			heroWinRate[name] = winRate
        }
		for name, counts := range overview.Players.HeroBeatCounts {
			if counts >= 2 {
				originHeroName, enemyHeroName := SplitHeroName(name)
				if heroBeatWinRate[enemyHeroName] == nil {
					heroBeatWinRate[enemyHeroName] = make(map[string]float32)
                }
				winRate := float32(overview.Players.HeroBeatWins[name]) / float32(counts)
				//data := fmt.Sprintf("%s%3d -%3d，胜率%4.4g%%，克制指数%4.4g%%\n",
				//name, overview.Players.HeroBeatWins[name],counts, winRate*100,  (winRate - heroWinRate[originHeroName]) * 100)
				heroBeatWinRate[enemyHeroName][originHeroName] =  winRate - heroWinRate[originHeroName]
			}
        }
		var HeroMap map[string]float32
		var maxLen int
		var haveTarget bool
		//遍历 敌方英雄-自己的英雄-克制指数列表
		//选出最大的列表
		for enemyHeroName, enemyHeroMap := range heroBeatWinRate {
			//如果没有目标英雄就跳过
			haveTarget = false
			for _, targetHeroName := range heroList {
				if targetHeroName == enemyHeroName {
					//fmt.Println(targetHeroName,enemyHeroName)
					haveTarget = true
                }
            }
			if !haveTarget {
				continue
            }
			if len(enemyHeroMap) > maxLen {
				//fmt.Println(len(enemyHeroMap),maxLen)
				maxLen = len(enemyHeroMap)
				//fmt.Println(enemyHeroName, enemyHeroMap)
				HeroMap = enemyHeroMap
            }
			//pList := MapSort(enemyHeroMap)
			//for _, p := range pList {
			//	show += p.Key
            //}
        }
		choiceHeroRateMap := make(map[string]float32)
		
		for _, targetHeroName := range heroList {
			delete(HeroMap,targetHeroName)
		}
		//遍历可供选择的英雄
		for choiceHeroName, _ := range HeroMap {
			//遍历目标英雄
			//fmt.Println("\n\n\n",HeroMap)
			for _, targetHeroName := range heroList {
				//fmt.Println(targetHeroName,choiceHeroName,heroBeatWinRate[targetHeroName][choiceHeroName])
				if winRate, ok := heroBeatWinRate[targetHeroName][choiceHeroName];ok {
					choiceHeroRateMap[choiceHeroName] += winRate
                }
			}
			choiceHeroRateMap[choiceHeroName]/=float32(len(heroList))
        }
		data, _ := json.Marshal(choiceHeroRateMap)
		show += ("用户ID" + overview.AccountId + string(data) + "\n")
    }

	return 200, show
}

func (s *apiServer) teamPickWinRate(params martini.Params) (int ,string ){
	heroListStr := params["herolist"]
	var show string
	heroList := strings.Split(heroListStr,"-")
	if len(heroList) == 0 {
		return 200 ,"NoHero"
	}
	
	for _, overview := range s.overview {
		heroBeatWinRate := make(map[string]map[string]float32)
		heroWinRate := make(map[string]float32) 
		for name, counts := range overview.Players.HeroCounts{			
			winRate := float32(overview.Players.HeroWins[name]) / float32(counts)
			heroWinRate[name] = winRate
        }
		for name, counts := range overview.Players.HeroBeatCounts {
			if counts >= 2 {
				originHeroName, enemyHeroName := SplitHeroName(name)
				if heroBeatWinRate[enemyHeroName] == nil {
					heroBeatWinRate[enemyHeroName] = make(map[string]float32)
                }
				winRate := float32(overview.Players.HeroBeatWins[name]) / float32(counts)
				//data := fmt.Sprintf("%s%3d -%3d，胜率%4.4g%%，克制指数%4.4g%%\n",
				//name, overview.Players.HeroBeatWins[name],counts, winRate*100,  (winRate - heroWinRate[originHeroName]) * 100)
				heroBeatWinRate[enemyHeroName][originHeroName] =  winRate
			}
        }
		var HeroMap map[string]float32
		var maxLen int
		var haveTarget bool
		//遍历 敌方英雄-自己的英雄-克制指数列表
		//选出最大的列表
		for enemyHeroName, enemyHeroMap := range heroBeatWinRate {
			//如果没有目标英雄就跳过
			haveTarget = false
			for _, targetHeroName := range heroList {
				if targetHeroName == enemyHeroName {
					//fmt.Println(targetHeroName,enemyHeroName)
					haveTarget = true
                }
            }
			if !haveTarget {
				continue
            }
			if len(enemyHeroMap) > maxLen {
				//fmt.Println(len(enemyHeroMap),maxLen)
				maxLen = len(enemyHeroMap)
				//fmt.Println(enemyHeroName, enemyHeroMap)
				HeroMap = enemyHeroMap
            }
			//pList := MapSort(enemyHeroMap)
			//for _, p := range pList {
			//	show += p.Key
            //}
        }
		choiceHeroRateMap := make(map[string]float32)
		
		for _, targetHeroName := range heroList {
			delete(HeroMap,targetHeroName)
        }
		//遍历可供选择的英雄
		for choiceHeroName, _ := range HeroMap {

			//遍历目标英雄
			//fmt.Println("\n\n\n",HeroMap)
			for _, targetHeroName := range heroList {
				//fmt.Println(targetHeroName,choiceHeroName,heroBeatWinRate[targetHeroName][choiceHeroName])
				if winRate, ok := heroBeatWinRate[targetHeroName][choiceHeroName];ok {
					choiceHeroRateMap[choiceHeroName] += winRate
				}else {
					choiceHeroRateMap[choiceHeroName] += 0.4
                }
			}
			choiceHeroRateMap[choiceHeroName] /= float32(len(heroList))
        }
		data, _ := json.Marshal(choiceHeroRateMap)
		show += ("用户ID" + overview.AccountId + string(data) + "\n")
    }

	return 200, show
}

func (s *apiServer) showOverview() (int, string) {
	var show string
	for _, overview := range s.overview {
		show += "玩家ID:"
		show += overview.AccountId
		show += "\n"
		heroBeatWinRate := make(map[string]map[string]float32)
		heroWinRate := make(map[string]float32) 
		for name, counts := range overview.Players.HeroCounts{			
			winRate := float32(overview.Players.HeroWins[name]) / float32(counts)
			heroWinRate[name] = winRate
        }
		for name, counts := range overview.Players.HeroBeatCounts {
			if counts >= 3 {
				originHeroName, _ := SplitHeroName(name)
				if heroBeatWinRate[originHeroName] == nil {
					heroBeatWinRate[originHeroName] = make(map[string]float32)
                }
				winRate := float32(overview.Players.HeroBeatWins[name]) / float32(counts)
				data := fmt.Sprintf("%s%3d -%3d，胜率%4.4g%%，克制指数%4.4g%%\n",
				name, overview.Players.HeroBeatWins[name],counts, winRate*100,  (winRate - heroWinRate[originHeroName]) * 100)
				heroBeatWinRate[originHeroName][data] =  winRate - heroWinRate[originHeroName]
			}
        }
		for _, originHeroMap := range heroBeatWinRate {
			pList := MapSort(originHeroMap)
			for _, p := range pList {
				show += p.Key
            }
        }
    }
	return 200, string(show)
}

func MergeHeroName(heroName string, enemyHeroName string) string{
	return heroName + "-" + enemyHeroName
}

func SplitHeroName(name string) (string, string) {
	s := strings.Split(name, "-")
	return s[0], s[1]
}

func (p *PlayerInfo) updatePlayerInfo(accountId string, data []byte) {
	if(p.HeroWins == nil) {
		p.Init()
    }

	var matchDetails MatchDetails
	json.Unmarshal(data,&matchDetails)
	for _, player := range matchDetails.Result.Players {
		//数据分析的玩家
		if strconv.Itoa(player.AccountId) == accountId {
			HeroName := HeroIdMap[player.HeroId]
			//fmt.Println(strconv.Itoa(player.HeroId) + " " + HeroName)
			//英雄出场数+1
			p.HeroCounts[HeroName]++
			//天辉
			if player.PlayerSlot < 5 {
				var enemyHeroName string
				for _, enemyPlayer := range matchDetails.Result.Players[5:] {
					//夜宴英雄名字
					enemyHeroName = HeroIdMap[enemyPlayer.HeroId]
					//对抗次数+1
					p.HeroBeatCounts[MergeHeroName(HeroName,enemyHeroName)]++
					if matchDetails.Result.RadiantWin == true {
						//对抗获胜次数+1
						p.HeroBeatWins[MergeHeroName(HeroName,enemyHeroName)]++
					}
				}
				if matchDetails.Result.RadiantWin == true {
					//获胜次数+1
					p.HeroWins[HeroName]++
				}
			} 
			//夜宴
			if player.PlayerSlot > 5 {
				var enemyHeroName string
				for _, enemyPlayer := range matchDetails.Result.Players[:5] {
					//天辉英雄名字
					enemyHeroName = HeroIdMap[enemyPlayer.HeroId]
					//对抗次数+1
					p.HeroBeatCounts[MergeHeroName(HeroName,enemyHeroName)]++
					if matchDetails.Result.RadiantWin == true {
						//对抗获胜次数+1
						p.HeroBeatWins[MergeHeroName(HeroName,enemyHeroName)]++
					}
				}
				if matchDetails.Result.RadiantWin == false{
					//获胜次数+1
					p.HeroWins[HeroName]++
				}
            }
        }
	}
}

func (s *apiServer) fetchId(params martini.Params) (int, string) {
	accountId := params["account_id"]
	key := params["key"]
	reqUrl := getMatchHistory + "&account_id=" + accountId + "&key=" + key
	data := httpGet(reqUrl + "&matches_requested=1")
	//fmt.Printf("%s\n",data);
	var matchHistory MatchHistory
	json.Unmarshal(data,&matchHistory)
	//fmt.Println(matchHistory);

	if(s.Players[accountId] == nil) {
		s.Players[accountId] = new(PlayerInfo)
	}
	OldMaxMatchId := s.Players[accountId].MaxMatchId
	if matchHistory.Result.NumResults != 0 {
		if matchHistory.Result.Matches[0].MatchId <= s.Players[accountId].MaxMatchId {
			return 200, "NoNewData"
		}
		s.Players[accountId].MaxMatchId = matchHistory.Result.Matches[0].MatchId
    }
	defer s.Save()
	for {
		fmt.Printf("matchHistory.Result.NumResults %d\n",matchHistory.Result.NumResults)
		//一轮解析开始
		var curMatchId int
		for _, match := range matchHistory.Result.Matches {
			//获取数据
			curMatchId = match.MatchId
			if curMatchId == OldMaxMatchId {
				return 200, "OK"
            }
			data = httpGet(getMatchDetails + "&match_id=" + strconv.Itoa(curMatchId) + "&key=" + key)
			s.Players[accountId].updatePlayerInfo(accountId, data)
		}
		s.Players[accountId].MatchCount += matchHistory.Result.NumResults
		fmt.Printf("MatchCount %d\n",s.Players[accountId].MatchCount)
		//一轮解析结束

		data = httpGet(reqUrl + "&matches_requested=100" + "&start_at_match_id=" + strconv.Itoa(curMatchId - 1))
		json.Unmarshal(data,&matchHistory)
		if matchHistory.Result.NumResults == 0 {
			break;
		}
		if curMatchId == matchHistory.Result.Matches[0].MatchId {
			return 200, "OK"
        }
    }
	return 200, "OK"
}

func newApiServer() http.Handler{
	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(render.Renderer())
	m.Use(func(w http.ResponseWriter, req *http.Request, c martini.Context) {
		path := req.URL.Path
		if req.Method == "GET" && strings.HasPrefix(path, "/"){
			var remoteAddr = req.RemoteAddr
			var headerAddr string
			for _, key := range []string{"X-Real-IP", "X-Forwarded-For"} {
				if val := req.Header.Get(key); val != "" {
					headerAddr = val
					break
				}
            }
			fmt.Printf("API call %s from %s [%s]\n", path, remoteAddr, headerAddr)
			if ip := strings.Split(remoteAddr,":");ip[0] != "172.17.140.52" {
				w.WriteHeader(404)
				return 
            }
        }
		c.Next()
    })
	api := &apiServer{Version: "1.00", Compile: "go"}
	api.Load()
	m.Use(func(c martini.Context, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
    })
	r := martini.NewRouter()
	r.Get("/", func(r render.Render) {
		r.Redirect("/overview")
    })
	r.Get("/overview", api.showOverview)
	r.Get("/fetch/:account_id/:key", api.fetchId)
	r.Get("/teampick/:herolist",api.teamPick)
	r.Get("/teampickwr/:herolist",api.teamPickWinRate)
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return m
}

func serve() error{
	l, err := net.Listen("tcp", "172.17.140.76:8081")
	//l, err := net.Listen("tcp", "192.168.52.128:8081")
	if err != nil {
		return err
    }
	eh := make(chan error, 1)
	go func(l net.Listener) {
		h := http.NewServeMux()
		h.Handle("/", newApiServer())
		hs := &http.Server{Handler: h}
		eh <- hs.Serve(l)
	}(l)
	err = <-eh
	return err
}

func main() {
	initHeroIdMap()
	err := serve()
	fmt.Printf("%v\n",err)
	for {
		time.Sleep(time.Second)
    }
}
