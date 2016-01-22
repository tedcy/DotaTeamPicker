package picker

import (
	"net/http"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"strings"
	"fmt"
	"log"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

type apiServer struct {
	Version  string
	Compile  string
	Players  map[string]*PlayerInfo
	overview []PlayerOverview
	fetchChan chan string
}

func NewApiServer() http.Handler {
	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(render.Renderer())
	m.Use(func(w http.ResponseWriter, req *http.Request, c martini.Context) {
		path := req.URL.Path
		if req.Method == "GET" && strings.HasPrefix(path, "/") {
			var remoteAddr = req.RemoteAddr
			var headerAddr string
			for _, key := range []string{"X-Real-IP", "X-Forwarded-For"} {
				if val := req.Header.Get(key); val != "" {
					headerAddr = val
					break
				}
			}
			fmt.Printf("API call %s from %s [%s]\n", path, remoteAddr, headerAddr)
			//if ip := strings.Split(remoteAddr,":");ip[0] != "172.17.140.52" {
			//	w.WriteHeader(404)
			//	return
			//}
		}
		c.Next()
	})
	api := &apiServer{Version: "1.00", Compile: "go"}
	api.Load()
	api.StartDaemonRoutines()
	m.Use(func(req *http.Request, c martini.Context, w http.ResponseWriter) {
		if req.Method == "GET" && strings.HasPrefix(req.URL.Path,"/teampickwrwithoutjson") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}else {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
        }
	})
	r := martini.NewRouter()
	r.Get("/", func(r render.Render) {
		r.Redirect("/overview")
	})
	r.Get("/overview", api.showOverview)
	r.Get("/fetch/:account_id", api.fetchId)
	r.Get("/teampick/:herolist", api.teamPick)
	r.Get("/teampickwr/:herolist", api.teamPickWinRate)
	r.Get("/teampickwrwithoutjson/:herolist", api.teamPickWinRateWithoutJSON)
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return m
}

func (s *apiServer) StartDaemonRoutines(){
	s.fetchChan = make(chan string,4096)
	go func() {
		for {
			accountId := <-s.fetchChan
			if len(accountId) == 0 {
				fmt.Println("fetchChan broken")
				return
            }
			fmt.Println("fetchId",accountId)
			s.fetchOneId(accountId)
        }
    }()
}

func searchSubStrToInt(dataStr string,subStr string,appendLen int) (string,int){
	findIndex := strings.Index(dataStr,subStr)
	if findIndex < 0 {
		return "", -1
    }
	countsIndex := findIndex + len(subStr) + appendLen
	var checkRun bool
	var i = 0
	for ;i != 20;i++ {
		if dataStr[countsIndex + i] < '0' || dataStr[countsIndex + i] > '9' {
			checkRun = true
			break
        }
    }
	if checkRun == false {
		return "", -1
    }
	return dataStr[countsIndex:countsIndex + i],countsIndex + i
}

func GetMatchCounts(accountId string) int{
	data := httpGet("http://dotamax.com/player/detail/" + accountId + "/")
	if data == nil {
		return -1
    }
	dataStr := string(data)
	subStr := "</div><div style=\"font-size: 11px;color:#777;\">"
	countsStr,index := searchSubStrToInt(dataStr,subStr,9)
	if index < 0 {
		return -1
    }
	counts,err := strconv.Atoi(countsStr)
	if err != nil {
		return -1	
    }
	return counts
}

func (s *apiServer) fetchIdAll(accountId string) {
	key := ConfigData.key
	/*matchCount := GetMatchCounts(accountId)
	if(matchCount < 0) {
		return 200,"no find id"
    }*/

	reqUrl := getMatchHistoryFromDotamax + accountId + "/?skill=&ladder=&hero=-1&p="
	index := 1
	subStr := "sorttable_customkey=\""
	//记录且仅记录一次maxMatchId的控制器
	var saveMaxMatchId bool
	lastMatchId := "999999999999"
	var wg sync.WaitGroup
	for ;;index++ {
		data := httpGet(reqUrl + strconv.Itoa(index))
		if data == nil {
			fmt.Println("req error")
			return
        }
		dataStr := string(data)
		var countsIndex int
		var curMatchId string
		for i:=0;;i++{
			curMatchId,countsIndex = searchSubStrToInt(dataStr,subStr,0)
			if countsIndex < 0 {
				break
			}
			//fmt.Println(curMatchId)
			if saveMaxMatchId == false {
				s.Players[accountId].MaxMatchId,_ = strconv.Atoi(curMatchId)
				saveMaxMatchId = true
            }

            {
				//如果当前的ID比上一个大，说明抓完了
				id, _ := strconv.Atoi(curMatchId)
				lastId, _ := strconv.Atoi(lastMatchId)
				if id > lastId {
					wg.Wait()
					s.Save()
					fmt.Println("OK")
					return
                }
				//每几个协程就等一下，防止抓太快被forbid
				if i % 4 == 0 {
					wg.Wait()
					time.Sleep(time.Second)
                }
			}

			go func(matchId string) {
				wg.Add(1)
				defer wg.Done()
				data := httpGet(getMatchDetails + "&match_id=" + matchId + "&key=" + key)
				if data == nil {
					fmt.Println("req error")
					return
				}
				if s.Players[accountId].updatePlayerInfo(accountId, data) == 0 {
					log.Println("MATCHID",matchId)
                }else {
					log.Println("ERRID",matchId)
                }

            }(curMatchId)
			lastMatchId = curMatchId
			dataStr = dataStr[countsIndex:]
        }
    }
	wg.Wait()
	s.Save()
	fmt.Println("OK")
	return
}

func (s *apiServer) fetchId(params martini.Params) (int, string) {
	accountId := params["account_id"]

	s.fetchChan <- accountId
	return 200, "send fetch request success"
}

func (s *apiServer) fetchOneId(accountId string) {
	if s.Players[accountId] == nil {
		s.Players[accountId] = new(PlayerInfo)
		//从dotamax抓
		s.Players[accountId].Init()
		s.fetchIdAll(accountId)
		return
	}
	//从steam抓
	key := ConfigData.key
	reqUrl := getMatchHistory + "&account_id=" + accountId + "&key=" + key
	data := httpGet(reqUrl + "&matches_requested=1")
	//fmt.Printf("%s\n",data);
	var matchHistory MatchHistory
	json.Unmarshal(data, &matchHistory)
	//fmt.Println(matchHistory);
	
	OldMaxMatchId := s.Players[accountId].MaxMatchId
	if matchHistory.Result.NumResults != 0 {
		if matchHistory.Result.Matches[0].MatchId <= s.Players[accountId].MaxMatchId {
			fmt.Println("NoNewData")
			return
		}
		s.Players[accountId].MaxMatchId = matchHistory.Result.Matches[0].MatchId
	}
	defer s.Save()
	for {
		fmt.Printf("matchHistory.Result.NumResults %d\n", matchHistory.Result.NumResults)
		//一轮解析开始
		var curMatchId int
		for _, match := range matchHistory.Result.Matches {
			//获取数据
			curMatchId = match.MatchId
			if curMatchId == OldMaxMatchId {
				fmt.Println("OK")
				return
			}
			data = httpGet(getMatchDetails + "&match_id=" + strconv.Itoa(curMatchId) + "&key=" + key)

			if s.Players[accountId].updatePlayerInfo(accountId, data) == 0 {
				log.Println("MATCHID",curMatchId)
            }else {
				log.Println("ERRID",curMatchId)
            }
		}
		s.Players[accountId].MatchCount += matchHistory.Result.NumResults
		fmt.Printf("MatchCount %d\n", s.Players[accountId].MatchCount)
		//一轮解析结束

		data = httpGet(reqUrl + "&matches_requested=100" + "&start_at_match_id=" + strconv.Itoa(curMatchId-1))
		json.Unmarshal(data, &matchHistory)
		if matchHistory.Result.NumResults == 0 {
			break
		}
		//重复解析，说明解析完毕
		if curMatchId == matchHistory.Result.Matches[0].MatchId {
			fmt.Println("OK")
			return
		}
	}
	fmt.Println("OK")
	return
}

func (s *apiServer) Save() {
	if len(s.overview) != 0 {
		s.overview = make([]PlayerOverview, 0)
	}
	for name, playerInfo := range s.Players {
		var tmp PlayerOverview
		tmp.AccountId = name
		tmp.Players = *playerInfo
		s.overview = append(s.overview, tmp)
	}
	data, _ := json.MarshalIndent(s.overview, "", "    ")
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

func (s *apiServer) teamPick(params martini.Params) (int, string) {
	heroListStr := params["herolist"]
	var show string
	heroList := strings.Split(heroListStr, "-")
	if len(heroList) == 0 {
		return 200, "NoHero"
	}

	for _, overview := range s.overview {
		heroBeatWinRate := make(map[string]map[string]float32)
		heroWinRate := make(map[string]float32)
		for name, counts := range overview.Players.HeroCounts {
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
				heroBeatWinRate[enemyHeroName][originHeroName] = winRate - heroWinRate[originHeroName]
			}
		}
		choiceHeroMap := make(map[string]int)
		for _, targetHeroName := range heroList {
			for originHeroName, _ := range heroBeatWinRate[targetHeroName] {
				choiceHeroMap[originHeroName] = 0
			}
		}
		
		choiceHeroRateMap := make(map[string]float32)

		for _, targetHeroName := range heroList {
			delete(choiceHeroMap, targetHeroName)
		}
		//遍历可供选择的英雄
		for choiceHeroName, _ := range choiceHeroMap {
			//遍历目标英雄
			//fmt.Println("\n\n\n",HeroMap)
			for _, targetHeroName := range heroList {
				//fmt.Println(targetHeroName,choiceHeroName,heroBeatWinRate[targetHeroName][choiceHeroName])
				if winRate, ok := heroBeatWinRate[targetHeroName][choiceHeroName]; ok {
					choiceHeroRateMap[choiceHeroName] += winRate
				}
			}
			choiceHeroRateMap[choiceHeroName] /= float32(len(heroList))
		}
		data, _ := json.Marshal(choiceHeroRateMap)
		show += ("用户ID" + overview.AccountId + string(data) + "\n")
	}

	return 200, show
}

func (s *apiServer) teamPickWinRate(params martini.Params) (int, string) {
	heroListStr := params["herolist"]
	var show string
	heroList := strings.Split(heroListStr, "-")
	if len(heroList) == 0 {
		return 200, "NoHero"
	}

	for _, overview := range s.overview {
		heroBeatWinRate := make(map[string]map[string]float32)
		heroWinRate := make(map[string]float32)
		for name, counts := range overview.Players.HeroCounts {
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
				heroBeatWinRate[enemyHeroName][originHeroName] = winRate
			}
		}
		choiceHeroMap := make(map[string]int)
		for _, targetHeroName := range heroList {
			for originHeroName, _ := range heroBeatWinRate[targetHeroName] {
				choiceHeroMap[originHeroName] = 0
			}
		}

		choiceHeroRateMap := make(map[string]float32)

		for _, targetHeroName := range heroList {
			delete(choiceHeroMap, targetHeroName)
		}
		//遍历可供选择的英雄
		for choiceHeroName, _ := range choiceHeroMap {

			//遍历目标英雄
			//fmt.Println("\n\n\n",HeroMap)
			for _, targetHeroName := range heroList {
				//fmt.Println(targetHeroName,choiceHeroName,heroBeatWinRate[targetHeroName][choiceHeroName])
				if winRate, ok := heroBeatWinRate[targetHeroName][choiceHeroName]; ok {
					choiceHeroRateMap[choiceHeroName] += winRate
				} else {
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

func (s *apiServer) teamPickWinRateWithoutJSON(params martini.Params) (int, string) {
	heroListStr := params["herolist"]
	show := "<html>"
	heroList := strings.Split(heroListStr, "-")
	if len(heroList) == 0 {
		return 200, "NoHero"
	}

	for _, overview := range s.overview {
		heroBeatWinRate := make(map[string]map[string]float32)
		heroWinRate := make(map[string]float32)
		for name, counts := range overview.Players.HeroCounts {
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
				heroBeatWinRate[enemyHeroName][originHeroName] = winRate
			}
		}
		choiceHeroMap := make(map[string]int)
		for _, targetHeroName := range heroList {
			for originHeroName, _ := range heroBeatWinRate[targetHeroName] {
				choiceHeroMap[originHeroName] = 0
			}
		}

		choiceHeroRateMap := make(map[string]float32)

		for _, targetHeroName := range heroList {
			delete(choiceHeroMap, targetHeroName)
		}
		//遍历可供选择的英雄
		for choiceHeroName, _ := range choiceHeroMap {

			//遍历目标英雄
			//fmt.Println("\n\n\n",HeroMap)
			for _, targetHeroName := range heroList {
				//fmt.Println(targetHeroName,choiceHeroName,heroBeatWinRate[targetHeroName][choiceHeroName])
				if winRate, ok := heroBeatWinRate[targetHeroName][choiceHeroName]; ok {
					choiceHeroRateMap[choiceHeroName] += winRate
				} else {
					choiceHeroRateMap[choiceHeroName] += 0.4
				}
			}
			choiceHeroRateMap[choiceHeroName] /= float32(len(heroList))
		}
		show += ("<b>用户ID:" + overview.AccountId + "</b><br>")
		if nickName, ok := ConfigData.nickNames[overview.AccountId];ok {
			show += ("<b>昵称:" + nickName + "</b><br>")
        }
		show += ("<b>全部比赛:" + strconv.Itoa(overview.Players.MatchCount) + "场</b><br>")
		choiceHeroList := MapSort(choiceHeroRateMap)
		var count int
		for _, choiceHero := range choiceHeroList {
			show = fmt.Sprintf("%s%s:%.1f%%&nbsp;&nbsp;&nbsp;&nbsp;",show,choiceHero.Key,choiceHero.Value*100)
			count++
			if (count % 5) == 0 {
				show += "<br>"
            }
        }
		show += "<br><br><br>"
	}
	show += "</html>"

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
		for name, counts := range overview.Players.HeroCounts {
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
					name, overview.Players.HeroBeatWins[name], counts, winRate*100, (winRate-heroWinRate[originHeroName])*100)
				heroBeatWinRate[originHeroName][data] = winRate - heroWinRate[originHeroName]
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


