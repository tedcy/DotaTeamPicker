package picker

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type apiServer struct {
	Version      string
	Compile      string
	Players      map[string]*PlayerInfo
	overviewLock sync.Mutex
	history      *AllHistory
	fetchChan    chan string
	db            *sql.DB
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
	api.InitDb()
	api.Load()
	api.StartDaemonRoutines()
	m.Use(gzip.All())
	m.Use(func(req *http.Request, c martini.Context, w http.ResponseWriter) {
		if req.Method == "GET" && strings.HasPrefix(req.URL.Path, "/teampickwrwithoutjson") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
	})
	r := martini.NewRouter()
	r.Get("/", func(r render.Render) {
		r.Redirect("/overview")
	})
	r.Get("/overview", api.showOverview)
	r.Get("/herooverviewtest", api.HeroOverviewTest)
	r.Get("/herooverview", api.HeroOverview)
	r.Get("/fetch/:account_id", api.fetchId)
	r.Get("/teampickwr/:herolist", api.teamPickAdvantage)
	r.Get("/teampick/:herolist", api.teamPickWinRate)
	r.Get("/teampickd/:herolist", api.teamPickerWinRateDefault)
	r.Get("/teampickdtest/:herolist", api.teamPickerWinRateDefaultTest)
	r.Get("/teampickwrwithoutjson/:herolist", api.teamPickWinRateWithoutJSON)
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return m
}

func (s *apiServer) InitDb() {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/DOTAMATCH")

	if err != nil {
		log.Printf("Open database error: %s\n", err)
	}
	//defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	s.db = db
}

func (s *apiServer) StartDaemonRoutines() {
	s.fetchChan = make(chan string, 4096)
	go func() {
		for {
			accountId := <-s.fetchChan
			if len(accountId) == 0 {
				fmt.Println("fetchChan broken")
				return
			}
			fmt.Println("fetchId", accountId)
			s.fetchOneId(accountId)
		}
	}()
	s.history = &AllHistory{}
	s.history.InitDb(s.db,s.Players)
	s.history.LoadDb()
	if ConfigData.testFetchMatches {
		go func() {
			s.history.FetchProcess()
		}()
	}
}

func (s *apiServer) fetchId(params martini.Params) (int, string) {
	accountId := params["account_id"]

	s.fetchChan <- accountId
	return 200, "send fetch request success"
}

func (s *apiServer) fetchOneId(accountId string) {
	if s.Players[accountId] == nil {
		//loadFromMysql
		s.Players[accountId] = new(PlayerInfo)
		s.Players[accountId].Init()
	}
	//fetchOne
	//for hero 0 - 109
	//for fetch
	//从steam抓
	key := ConfigData.key
	reqUrl := getMatchHistory + "account_id=" + accountId + "&key=" + key
	data := httpGet(reqUrl + "&matches_requested=1")
	//fmt.Printf("%s\n",data);
	var matchHistory MatchHistory
	json.Unmarshal(data, &matchHistory)
	//fmt.Println(matchHistory);

	if matchHistory.Result.NumResults == 0 {
		log.Println("no match accountId: ",accountId)
		return
	}
	defer s.Save(accountId)
	s.Players[accountId].MaxMatchSeq = matchHistory.Result.Matches[0].MatchSeqNum
	for heroId, _ := range HeroIdStrMap {
		data = httpGet(reqUrl + "&matches_requested=100" + "&hero_id=" + heroId)
		json.Unmarshal(data, &matchHistory)
		if matchHistory.Result.NumResults == 0 {
			fmt.Println(heroId,"OK")
			continue
		}
		for {
			fmt.Printf("matchHistory.Result.NumResults %d\n", matchHistory.Result.NumResults)
			//一轮解析开始
			var curMatchId int
			for _, match := range matchHistory.Result.Matches {
				//获取数据
				curMatchId = match.MatchId
				data = httpGet(getMatchDetails + "match_id=" + strconv.Itoa(curMatchId) + "&key=" + key)
				var matchDetails MatchDetails
				json.Unmarshal(data, &matchDetails)

				if s.Players[accountId].updatePlayerInfo(accountId, &matchDetails) == 0 {
					log.Println("MATCHID", curMatchId)
				} else {
					log.Println("ERRID", curMatchId)
				}
			}
			//s.Players[accountId].MatchCount += matchHistory.Result.NumResults
			fmt.Printf("MatchCount %d\n", s.Players[accountId].MatchCount)
			//一轮解析结束

			data = httpGet(reqUrl + "&matches_requested=100" + "&start_at_match_id=" + strconv.Itoa(curMatchId-1) + "&hero_id=" + heroId)
			json.Unmarshal(data, &matchHistory)
			if matchHistory.Result.NumResults == 0 {
				fmt.Println(heroId,"OK")
				break
			}
			//重复解析，说明解析完毕
			if curMatchId == matchHistory.Result.Matches[0].MatchId {
				fmt.Println("OK")
				break;
			}
		}
	}
	fmt.Println("OK")
	return
}

func (s *apiServer) Save(accountId string) {
	data, _ := json.Marshal(s.Players[accountId])
		
	stmtIns, err := s.db.Prepare("INSERT INTO PlayerInfo VALUES ( ?, ? )")

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	defer stmtIns.Close()

	id,_ := strconv.Atoi(accountId)
	_, err = stmtIns.Exec(id, string(data))

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
}


func (s *apiServer) Load() {
	s.Players = make(map[string]*PlayerInfo)
	stmtOut, err := s.db.Prepare("SELECT * FROM PlayerInfo")

	if err != nil {
		log.Printf("%s\n", err)
		return
    }
	defer stmtOut.Close()

	rows,err := stmtOut.Query()

	if err != nil {
		log.Printf("%s\n", err)
		return
    }

	defer rows.Close()

	var accountId int
	var data string
	for rows.Next() {
		err := rows.Scan(&accountId, &data)

		log.Println(data)
		if err != nil {
			fmt.Printf("%s\n", err)
        }
		var p PlayerInfo	
		json.Unmarshal([]byte(data), &p)
		s.Players[strconv.Itoa(accountId)] = &p
    }

	err = rows.Err()
	if err != nil {
		log.Printf("%s\n", err)
		return
    }
}

func (s *apiServer) teamPickAdvantage(params martini.Params) (int, string) {
	heroListStr := params["herolist"]
	heroList := strings.Split(heroListStr, "-")
	if len(heroList) == 0 {
		return 200, "NoHero"
	}
	var choiceHeroRateMapsForShow []choiceForShow

	//加锁获取一个overview镜像
	overviewTemp := []PlayerOverview{}
	s.overviewLock.Lock()
	for accountId, p := range s.Players {
		var overview PlayerOverview
		overview.AccountId = accountId
		overview.Players = *p
		overviewTemp = append(overviewTemp, overview)
	}
	s.overviewLock.Unlock()

	allHeroWinRate := make(map[string]float32)
	for Id,_ := range HeroIdStrMap {
		allHeroWinRate[Id] = float32(s.history.Hero.HeroWins[Id]) / float32(s.history.Hero.HeroCounts[Id])	
    }
	allHeroRateMap := s.history.Hero.showHeroInfoOverview()
	for _, overview := range overviewTemp {
		heroBeatWinRate := make(map[string]map[string]float32)
		heroWinRate := make(map[string]float32)
		for name, counts := range overview.Players.HeroCounts {
			winRate := float32(overview.Players.HeroWins[name]) / float32(counts)
			heroWinRate[name] = winRate
		}
		for name, counts := range overview.Players.HeroBeatCounts {
			if counts >= 5 {
				originHeroName, enemyHeroName := SplitHeroName(name)
				if heroBeatWinRate[enemyHeroName] == nil {
					heroBeatWinRate[enemyHeroName] = make(map[string]float32)
				}
				winRate := float32(overview.Players.HeroBeatWins[name]) / float32(counts)
				//data := fmt.Sprintf("%s%3d -%3d，胜率%4.4g%%，克制指数%4.4g%%\n",
				//name, overview.Players.HeroBeatWins[name],counts, winRate*100,  (winRate - heroWinRate[originHeroName]) * 100)
				heroBeatWinRate[enemyHeroName][originHeroName] = winRate + allHeroWinRate[enemyHeroName] - heroWinRate[originHeroName] - 0.5
				//log.Println(originHeroName,enemyHeroName,winRate,allHeroWinRate[enemyHeroName],heroWinRate[originHeroName])
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
				}else {
					choiceHeroRateMap[choiceHeroName] += allHeroRateMap[choiceHeroName][targetHeroName]
                }
				//log.Println(choiceHeroName,targetHeroName,heroBeatWinRate[targetHeroName][choiceHeroName],allHeroRateMap[choiceHeroName][targetHeroName])
			}
			choiceHeroRateMap[choiceHeroName] /= float32(len(heroList))
		}
		//输出格式准备
		choiceHeroRateMapForShow := &choiceForShow{AccountId: overview.AccountId, ChoiceHeroRateMap: choiceHeroRateMap}
		if nickName, ok := ConfigData.nickNames[overview.AccountId]; ok {
			choiceHeroRateMapForShow.NickName = nickName
		} else {
			choiceHeroRateMapForShow.NickName = "未定义的昵称"
		}
		choiceHeroRateMapsForShow = append(choiceHeroRateMapsForShow, *choiceHeroRateMapForShow)
	}
	data, err := json.Marshal(choiceHeroRateMapsForShow)
	if err != nil {
		return 200, err.Error()
    }

	return 200, string(data)
}

type choiceForShow struct {
	AccountId         string
	NickName          string
	ChoiceHeroRateMap map[string]float32
}

func (s *apiServer) teamPickWinRate(params martini.Params) (int, string) {
	heroListStr := params["herolist"]
	//var show string
	heroList := strings.Split(heroListStr, "-")
	if len(heroList) == 0 {
		return 200, "NoHero"
	}
	var choiceHeroRateMapsForShow []choiceForShow

	//加锁获取一个overview镜像
	overviewTemp := []PlayerOverview{}
	s.overviewLock.Lock()
	for accountId, p := range s.Players {
		var overview PlayerOverview
		overview.AccountId = accountId
		overview.Players = *p
		overviewTemp = append(overviewTemp, overview)
	}
	s.overviewLock.Unlock()

	for _, overview := range overviewTemp {
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

		//输出格式准备
		choiceHeroRateMapForShow := &choiceForShow{AccountId: overview.AccountId, ChoiceHeroRateMap: choiceHeroRateMap}
		if nickName, ok := ConfigData.nickNames[overview.AccountId]; ok {
			choiceHeroRateMapForShow.NickName = nickName
		} else {
			choiceHeroRateMapForShow.NickName = "未定义的昵称"
		}
		choiceHeroRateMapsForShow = append(choiceHeroRateMapsForShow, *choiceHeroRateMapForShow)
		//show += ("用户ID" + overview.AccountId + string(data) + "\n")
	}
	data, _ := json.Marshal(choiceHeroRateMapsForShow)

	return 200, string(data)
}

func (s *apiServer) teamPickWinRateWithoutJSON(params martini.Params) (int, string) {
	heroListStr := params["herolist"]
	show := "<html>"
	heroList := strings.Split(heroListStr, "-")
	if len(heroList) == 0 {
		return 200, "NoHero"
	}

	//加锁获取一个overview镜像
	overviewTemp := []PlayerOverview{}
	s.overviewLock.Lock()
	for accountId, p := range s.Players {
		var overview PlayerOverview
		overview.AccountId = accountId
		overview.Players = *p
		overviewTemp = append(overviewTemp, overview)
	}
	s.overviewLock.Unlock()

	for _, overview := range overviewTemp {
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
		if nickName, ok := ConfigData.nickNames[overview.AccountId]; ok {
			show += ("<b>昵称:" + nickName + "</b><br>")
		}
		show += ("<b>全部比赛:" + strconv.Itoa(overview.Players.MatchCount) + "场</b><br>")
		choiceHeroList := MapSort(choiceHeroRateMap)
		var count int
		for _, choiceHero := range choiceHeroList {
			show = fmt.Sprintf("%s%s:%.1f%%&nbsp;&nbsp;&nbsp;&nbsp;", show, choiceHero.Key, choiceHero.Value*100)
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

func (s *apiServer) teamPickerWinRateDefault(params martini.Params) (int, string) {
	heroListStr := params["herolist"]
	//var show string
	heroList := strings.Split(heroListStr, "-")
	if len(heroList) == 0 {
		return 200, "NoHero"
	}
	choiceHeroRateMap := s.history.Hero.showHeroInfo(heroList)
	data,err := json.Marshal(choiceHeroRateMap)
	if err != nil {
		return 200, err.Error()
    }
	show := string(data)

	return 200, show
}

func (s *apiServer) teamPickerWinRateDefaultTest(params martini.Params) (int, string) {
	heroListStr := params["herolist"]
	var show string
	heroList := strings.Split(heroListStr, "-")
	if len(heroList) == 0 {
		return 200, "NoHero"
	}
	show += "敌方英雄列表"
	for _,id := range heroList {
		show += "-"
		show += HeroIdStrMap[id]
    }
	show += "\n"
	choiceHeroRateMap := s.history.Hero.showHeroInfo(heroList)
	pList := MapSort(choiceHeroRateMap)
	show += "可选英雄克制指数\n"
	for _,p := range pList {
		show = fmt.Sprintf("%s%s:%.1f%%\n", show, HeroIdStrMap[p.Key], p.Value*100)
    }

	return 200, show
}

func (s *apiServer) HeroOverview(params martini.Params) (int, string) {
	choiceHeroRateMap := s.history.Hero.showHeroInfoOverview()
	data,err := json.Marshal(choiceHeroRateMap)
	if err != nil {
		return 200, err.Error()
    }
	show := string(data)

	return 200, show
}

func (s *apiServer) HeroOverviewTest(params martini.Params) (int, string) {
	var show string
	choiceHeroRateMap := s.history.Hero.showHeroInfoOverview()
	for Id1,Id1Map := range choiceHeroRateMap {
		pList := MapSort(Id1Map)
		for _,p := range pList {
			show = fmt.Sprintf("%s%svs%s:%.1f%%\n", show, HeroIdStrMap[Id1], HeroIdStrMap[p.Key], p.Value*100)
        }
    }

	return 200, show
}

func (s *apiServer) showOverview() (int, string) {
	var show string

	//加锁获取一个overview镜像
	overviewTemp := []PlayerOverview{}
	s.overviewLock.Lock()
	for accountId, p := range s.Players {
		var overview PlayerOverview
		overview.AccountId = accountId
		overview.Players = *p
		overviewTemp = append(overviewTemp, overview)
	}
	s.overviewLock.Unlock()

	for _, overview := range overviewTemp {
		show += ("玩家ID:" + overview.AccountId + "\n")
		if nickName, ok := ConfigData.nickNames[overview.AccountId]; ok {
			show += ("昵称:" + nickName + "\n")
		}
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
		show += "\n\n\n\n\n"
	}
	return 200, string(show)
}
