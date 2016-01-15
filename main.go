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
var getMatchHistory = "https://api.steampowered.com/IDOTA2Match_570/GetMatchHistory/V001/?matches_requested=1"

type MatchDetails struct {
	Result struct {
		Players [10] struct {
			Account_id int	`json:"account_id"`                                                              
			Player_slot int `json:"player_slot"`                                        
			Hero_id int `json:"hero_id"`                                                                    
		}`json:"players"`                                                            
		Radiant_win bool `json:"radiant_win"`                                                             
	} `json:"result"`
}

type MatchHistory struct {
	Result struct {
		Status int
		Matches []struct {
			Match_id int
		} `json:"matches"`
	} `json:"result"`
}

type apiServer struct {
	Version string
	Compile string
}

func httpGet(url string) ([]byte){
	fmt.Println(url)
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

func (s *apiServer) fetchId(params martini.Params) (int, string) {
	account_id := params["account_id"]
	key := params["key"]
	data := httpGet(getMatchHistory + "&account_id=" + account_id + "&key=" + key)
	//fmt.Printf("%s\n",data);
	var matchHistory MatchHistory
	json.Unmarshal(data,&matchHistory)
	//fmt.Println(matchHistory);
	match_id := strconv.Itoa(matchHistory.Result.Matches[0].Match_id)
	data = httpGet(getMatchDetails + "&match_id=" + match_id + "&key=" + key)
	var matchDetails MatchDetails
	json.Unmarshal(data,&matchDetails)
	for _, player := range matchDetails.Result.Players {
		fmt.Println(strconv.Itoa(player.Account_id))
    }
	return 200, match_id
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
	l, err := net.Listen("tcp", "172.17.140.76:8081")
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
	err := serve()
	fmt.Printf("%v\n",err)
	for {
		time.Sleep(time.Second)
    }
}
