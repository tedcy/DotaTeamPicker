package picker

import (
	"encoding/json"
	"log"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"time"
)

/*CREATE TABLE `MatchHistory` (
	`fetchSeqStart` int(11) NOT NULL,
	`fetchSeqEnd` int(11) NOT NULL,
	`zeroSeq` int(11) NOT NULL,
	`nouse` int(11) NOT NULL,
	PRIMARY KEY (`nouse`)
);*/
/*
INSERT INTO MatchHistory VALUES( 0, 0, 0, 1);
*/

type AllHistory struct {
	fetchSeqStart int64
	fetchSeqEnd   int64
	zeroSeq       int64
	Players      map[string]*PlayerInfo
	Hero		  HeroInfo
	db            *sql.DB
}

/*CREATE TABLE `MatchWithPlayers` (
	`matchId` int(11) NOT NULL,
	`playerIds` char(109) NOT NULL,
	PRIMARY KEY (`matchId`)
);*/

type MatchInfoMatch struct {
	Players [10] Player `json:"players"`
	MatchId     int64 `json:"match_id"`
	MatchSeqNum int64 `json:"match_seq_num"`
	StartTime   int64 `json:"start_time"`
	RadiantWin bool `json:"radiant_win"`
	HumanPlayers int `json:"human_players"`
}

type MatchInfo struct {
	Result struct {
		Matches []MatchInfoMatch `json:"matches"`
	} `json:"result"`
}

var getMatchHistoryWithSeq = "https://api.steampowered.com/IDOTA2Match_570/GetMatchHistoryBySequenceNum/V001/?"
var getOneMatch = "https://api.steampowered.com/IDOTA2Match_570/GetMatchHistory/V001/?"

func FetchMatchsBySeq(seq int64) MatchInfo {
	var reqUrl string
	if seq == 0 {
		reqUrl = getOneMatch + "key=" + ConfigData.key
    }else {
		reqUrl = getMatchHistoryWithSeq + "key=" + ConfigData.key + "&start_at_match_seq_num=" + strconv.FormatInt(seq,10)
    }
	log.Println(reqUrl)
	data := httpGet(reqUrl)
	var matchInfo MatchInfo
	json.Unmarshal(data, &matchInfo)
	return matchInfo
} 

/*
**************************************** fetchSeqEnd

recorded

**************************************** fetchSeqStart

no record

**************************************** zeroSeq
*/

func (h *AllHistory) FetchProcess() {
	//历史记录已经从sql里加载完毕
	for {
		//第一次运行
		if h.fetchSeqEnd == 0 {
			//h.fetchSeqStart = 1700000000
			//h.fetchSeqEnd = 1700000000
			data := httpGet("https://api.steampowered.com/IDOTA2Match_570/GetMatchHistory/V001/?matches_requested=1&key=09D4D194967FAA9D99D9884BA0BEF3F7")
			var matchHistory MatchHistory
			json.Unmarshal(data, &matchHistory)
			h.fetchSeqStart = int64(matchHistory.Result.Matches[0].MatchSeqNum)
			h.fetchSeqEnd = int64(matchHistory.Result.Matches[0].MatchSeqNum)
			log.Println("start",h.fetchSeqStart,"end",h.fetchSeqEnd)
			time.Sleep(time.Second)
			continue
		}
		//还没达到686的时间戳
		if h.zeroSeq == 0 {
			matchInfo := FetchMatchsBySeq(h.fetchSeqEnd)
			if matchInfo.Result.Matches == nil {
				time.Sleep(time.Second * 20)
				continue
            }
			//如果达到时间戳，那么记录下该matchId
			if h.checkMatchReachZeroLine(matchInfo) {
				h.SaveHistory()
				continue
            }
			h.fetchSeqEnd = matchInfo.Result.Matches[len(matchInfo.Result.Matches) - 1].MatchSeqNum + 1
			log.Println("start",h.fetchSeqStart,"end",h.fetchSeqEnd)
			h.SaveHistory()
			time.Sleep(time.Second * 2)
			continue
		}
		matchInfo := FetchMatchsBySeq(h.fetchSeqEnd)
		if matchInfo.Result.Matches == nil {
			time.Sleep(time.Second * 20)
			continue
        }
		//需要+1，否则会重复拉取一次
		h.fetchSeqEnd = matchInfo.Result.Matches[len(matchInfo.Result.Matches) - 1].MatchSeqNum + 1
		h.SaveMatches(matchInfo.Result.Matches)
		log.Println("start",h.fetchSeqStart,"end",h.fetchSeqEnd)
		h.SaveHistory()
	}
}

func (h *AllHistory) InitDb(db *sql.DB, Players map[string]*PlayerInfo) {
	h.db = db
	h.Players = Players
}

func (h *AllHistory) LoadDb() {
	stmtOut, err := h.db.Prepare("SELECT * FROM MatchHistory WHERE nouse = ?")

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(1)

	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	defer rows.Close()

	var nouse int
	for rows.Next() {
		err := rows.Scan(&h.fetchSeqStart, &h.fetchSeqEnd, &h.zeroSeq, &nouse)
		var _ = nouse

		if err != nil {
			log.Printf("%s\n", err)
		}
		log.Println(h)
	}

	err = rows.Err()
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	h.Hero.LoadDb(h.db)
}

func (h *AllHistory) SaveHistory() {
	h.Hero.SaveDb(h.db)
	stmtSaveHistory, err := h.db.Prepare("UPDATE MatchHistory SET fetchSeqStart = ?,fetchSeqEnd = ?,zeroSeq = ? WHERE nouse = 1")

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	defer stmtSaveHistory.Close()

	_, err = stmtSaveHistory.Exec(h.fetchSeqStart, h.fetchSeqEnd, h.zeroSeq)
}

func (h *AllHistory) SaveMatches(matches []MatchInfoMatch) {

	/*stmtIns, err := h.db.Prepare("INSERT INTO MatchWithPlayers VALUES( ?, ? )")

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	defer stmtIns.Close()*/

	var i int
	for _, m := range matches {
		var accountIds string
		valid := true
		//检查比赛内容是否有10人参与，是否全部选择了英雄
		if m.HumanPlayers < 10 {
			valid = false
        } else {
			for _, p := range m.Players {
				if p.HeroId == 0 {
					valid = false
					break
				}
				//if accountIds == "" {
				//	accountIds = strconv.FormatInt(p.AccountId,10)
				//}
				//accountIds += (";" + strconv.FormatInt(p.AccountId,10))
			}
        }
		if !valid {
			continue
        }
		//更新已经注册的玩家数据
		for _,catchP := range m.Players {
			p, ok := h.Players[strconv.FormatInt(catchP.AccountId,10)] 
			//存在并且大于已经记录的最大值
			if ok && m.MatchSeqNum > int64(p.MaxMatchSeq) {
				log.Println("fetchAllHistory match accountId",catchP.AccountId,m.MatchSeqNum)
				var matchDetails MatchDetails
				matchDetails.Result.Players = m.Players
				matchDetails.Result.RadiantWin = m.RadiantWin
				p.updatePlayerInfo(strconv.FormatInt(catchP.AccountId,10),&matchDetails)
            }
        }
		//更新所有英雄对战胜率
		h.Hero.updateHeroInfo(&m)
		i++
		log.Println("ID: ",m.MatchId,"Players: ",accountIds)
		/*_, err = stmtIns.Exec(m.MatchId, accountIds)

		if err != nil {
			log.Printf("%s\n", err)
			return
		}*/
	}
	log.Println("parse: ",len(matches),"save: ",i)
}

func (h *AllHistory) checkMatchReachZeroLine(matchInfo MatchInfo) bool{
	if matchInfo.Result.Matches[len(matchInfo.Result.Matches) - 1].StartTime < ConfigData.limit6_86 {
		return false
    }
	for _, m := range matchInfo.Result.Matches {
		//log.Println(m)
		if m.StartTime >= ConfigData.limit6_86 {
			h.zeroSeq = m.MatchSeqNum
			h.fetchSeqEnd = m.MatchSeqNum
			return true
		}
	}
	return false
}
