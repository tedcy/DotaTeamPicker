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
)

var getMatchDetails = "https://api.steampowered.com/IDOTA2Match_570/GetMatchDetails/v001/?"
var getMatchHistory = "https://api.steampowered.com/IDOTA2Match_570/GetMatchHistory/V001/?matches_requested=500"

var HeroNameMap = map[string]int {
		"Abaddon": 102,
		"Alchemist": 73,
		"Ancient Apparition": 68,
		"Anti-Mage": 1,
		"Arc Warden": 113,
		"Axe": 2,
		"Bane": 3,
		"Batrider": 65,
		"Beastmaster": 38,
		"Bloodseeker": 4,
		"Bounty Hunter": 62,
		"Brewmaster": 78,
		"Bristleback": 99,
		"Broodmother": 61,
		"Centaur Warrunner": 96,
		"Chaos Knight": 81,
		"Chen": 66,
		"Clinkz": 56,
		"Clockwerk": 51,
		"Crystal Maiden": 5,
		"Dark Seer": 55,
		"Dazzle": 50,
		"Death Prophet": 43,
		"Disruptor": 87,
		"Doom": 69,
		"Dragon Knight": 49,
		"Drow Ranger": 6,
		"Earth Spirit": 107,
		"Earthshaker": 7,
		"Elder Titan": 103,
		"Ember Spirit": 106,
		"Enchantress": 58,
		"Enigma": 33,
		"Faceless Void": 41,
		"Gyrocopter": 72,
		"Huskar": 59,
		"Invoker": 74,
		"Io": 91,
		"Jakiro": 64,
		"Juggernaut": 8,
		"Keeper of the Light": 90,
		"Kunkka": 23,
		"Legion Commander": 104,
		"Leshrac": 52,
		"Lich": 31,
		"Lifestealer": 54,
		"Lina": 25,
		"Lion": 26,
		"Lone Druid": 80,
		"Luna": 48,
		"Lycan": 77,
		"Magnus": 97,
		"Medusa": 94,
		"Meepo": 82,
		"Mirana": 9,
		"Morphling": 10,
		"Naga Siren": 89,
		"Natures Prophet": 53,
		"Necrophos": 36,
		"Night Stalker": 60,
		"Nyx Assassin": 88,
		"Ogre Magi": 84,
		"Omniknight": 57,
		"Oracle": 111,
		"Outworld Devourer": 76,
		"Phantom Assassin": 44,
		"Phantom Lancer": 12,
		"Phoenix": 110,
		"Puck": 13,
		"Pudge": 14,
		"Pugna": 45,
		"Queen of Pain": 39,
		"Razor": 15,
		"Riki": 32,
		"Rubick": 86,
		"Sand King": 16,
		"Shadow Demon": 79,
		"Shadow Fiend": 11,
		"Shadow Shaman": 27,
		"Silencer": 75,
		"Skywrath Mage": 101,
		"Slardar": 28,
		"Slark": 93,
		"Sniper": 35,
		"Spectre": 67,
		"Spirit Breaker": 71,
		"Storm Spirit": 17,
		"Sven": 18,
		"Techies": 105,
		"Templar Assassin": 46,
		"Terrorblade": 109,
		"Tidehunter": 29,
		"Timbersaw": 98,
		"Tinker": 34,
		"Tiny": 19,
		"Treant Protector": 83,
		"Troll Warlord": 95,
		"Tusk": 100,
		"Undying": 85,
		"Ursa": 70,
		"Vengeful Spirit": 20,
		"Venomancer": 40,
		"Viper": 47,
		"Visage": 92,
		"Warlock": 37,
		"Weaver": 63,
		"Windranger": 21,
		"Winter Wyvern": 112,
		"Witch Doctor": 30,
		"Wraith King": 42,
		"Zeus": 22,
}

var HeroIdMap map[int]string

func initHeroIdMap() {
	for _, heroName := range HeroIdMap {
		HeroIdMap[HeroNameMap[heroName]] = heroName
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
		Matches []struct {
			MatchId int `json:"match_id"`
		} `json:"matches"`
	} `json:"result"`
}

type PlayerInfo struct {
	heroWins map[string]int
	heroCounts map[string]int
	heroBeatWins map[string]int
	heroBeatCounts map[string]int
}

func (p *PlayerInfo) Init() {
	p.heroWins = make(map[string]int)
	p.heroCounts = make(map[string]int)
	p.heroBeatWins = make(map[string]int)
	p.heroBeatCounts = make(map[string]int)
}

type apiServer struct {
	Version string
	Compile string
	Players map[string]*PlayerInfo
}

func httpGet(url string) ([]byte){
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return body
}

func (s *apiServer) overview() (int, string) {
	var o struct {Version string;Compile string}
	o.Version = s.Version
	o.Compile = s.Compile
	b, _ := json.Marshal(o)
	return 200, string(b)
}

func (s *apiServer) updatePlayerInfo(accountId string, data []byte) {
	var matchDetails MatchDetails
	json.Unmarshal(data,&matchDetails)
	for _, player := range matchDetails.Result.Players {
		fmt.Println(player.AccountId)
		fmt.Println(accountId)
		if strconv.Itoa(player.AccountId) == accountId {
			HeroName := HeroIdMap[player.HeroId]
			if(s.Players[accountId].heroWins == nil) {
				s.Players[accountId].Init()
            }
			s.Players[accountId].heroBeatCounts[HeroName]++
			if player.PlayerSlot < 5 && matchDetails.Result.RadiantWin == true {
				s.Players[accountId].heroBeatWins[HeroName]++
            }
			fmt.Println(HeroName)
        }
	}
	fmt.Println(s.Players[accountId].heroBeatWins["Medusa"])
	fmt.Println(s.Players[accountId].heroBeatCounts["Medusa"])
	fmt.Println("")
}

func (s *apiServer) fetchId(params martini.Params) (int, string) {
	accountId := params["account_id"]
	key := params["key"]
	data := httpGet(getMatchHistory + "&account_id=" + accountId + "&key=" + key)
	//fmt.Printf("%s\n",data);
	var matchHistory MatchHistory
	json.Unmarshal(data,&matchHistory)
	fmt.Println(matchHistory);
	var player PlayerInfo
	s.Players[accountId] = &player
	var matchId string
	for _, match := range matchHistory.Result.Matches {
		if match.MatchId == 0 {
			return 502, "ERROR"
		}
		
		matchId = strconv.Itoa(match.MatchId)
		data = httpGet(getMatchDetails + "&match_id=" + matchId + "&key=" + key)
		s.updatePlayerInfo(accountId, data)
    }
	return 200, "OK"
}

func newApiServer() http.Handler{
	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(render.Renderer())
	m.Use(func(w http.ResponseWriter, req *http.Request, c martini.Context) {
		path := req.URL.Path
		if strings.HasPrefix(path, "/overview") {
			var remoteAddr = req.RemoteAddr
			var headerAddr string
			for _, key := range []string{"X-Real-IP", "X-Forwarded-For"} {
				if val := req.Header.Get(key); val != "" {
					headerAddr = val
					break
				}
            }
			fmt.Printf("API call %s from %s [%s]\n", path, remoteAddr, headerAddr)
        }
		c.Next()
    })
	api := &apiServer{Version: "1.00", Compile: "go"}
	api.Players = make(map[string]*PlayerInfo)
	m.Use(func(c martini.Context, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
    })
	r := martini.NewRouter()
	r.Get("/", func(r render.Render) {
		r.Redirect("/overview")
    })
	r.Get("/overview", api.overview)
	r.Get("/fetch/:account_id/:key", api.fetchId)
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return m
}

func serve() error{
	l, err := net.Listen("tcp", "192.168.52.128:8081")
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
