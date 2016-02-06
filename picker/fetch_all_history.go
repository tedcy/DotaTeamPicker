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

type AllHistory struct {
	fetchSeqStart int64
	fetchSeqEnd   int64
	zeroSeq       int64
	db            *sql.DB
}

/*CREATE TABLE `MatchWithPlayers` (
	`matchId` int(11) NOT NULL,
	`playerIds` char(109) NOT NULL,
	PRIMARY KEY (`matchId`)
);*/

type MatchInfoMatch struct {
	Players [10]struct {
		AccountId int64 `json:"account_id"`
		HeroId    int`json:"hero_id"`
	} `json:"players"`
	MatchId     int64 `json:"match_id"`
	MatchSeqNum int64 `json:"match_seq_num"`
	StartTime   int64 `json:"start_time"`
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
		reqUrl = getOneMatch + "&key=" + ConfigData.key
    }else {
		reqUrl = getMatchHistoryWithSeq + "&key=" + ConfigData.key + "&start_at_match_seq_num=" + strconv.FormatInt(seq,10)
    }
	log.Println(reqUrl)
	data := httpGet(reqUrl)
	var matchInfo MatchInfo
	json.Unmarshal(data, &matchInfo)
	return matchInfo
} 

/*
**************************************** fetchSeqStart

recorded

**************************************** fetchSeqEnd

no record

**************************************** zeroSeq
*/

func (h *AllHistory) FetchProcess() {
	//历史记录已经从sql里加载完毕
	for {
		if h.fetchSeqEnd == 0 {
			matchInfo := FetchMatchsBySeq(0)
			h.fetchSeqStart = matchInfo.Result.Matches[0].MatchSeqNum
			h.fetchSeqEnd = matchInfo.Result.Matches[0].MatchSeqNum
			log.Println("start",h.fetchSeqStart,"end",h.fetchSeqEnd)
			time.Sleep(time.Second)
			continue
		}
		//历史记录未抓取完整
		if h.zeroSeq == 0 {
			matchInfo := FetchMatchsBySeq(h.fetchSeqEnd)
			if matchInfo.Result.Matches = nil {
				time.Sleep(time.Second)
				continue
            }
			//检查到686的时间戳
			//如果达到时间戳，那么记录下该matchId
			matches := h.checkMatchReachZeroLine(matchInfo)
			h.fetchSeqEnd = matchInfo.Result.Matches[len(matchInfo.Result.Matches) - 1].MatchSeqNum
			if matches == nil {
				continue
            }
			h.SaveMatches(matches)
			log.Println("start",h.fetchSeqStart,"end",h.fetchSeqEnd)
			h.SaveHistory()
			continue
		}
		//抓取最新记录
		matchInfo := FetchMatchsBySeq(h.fetchSeqEnd)
		if matchInfo.Result.Matches = nil {
			time.Sleep(time.Second)
			continue
        }
		//检查是否到zeroSeq
		//如果达到，那么zeroSeq = fetchSeqEnd = fetchSeqStart
		matches := h.checkMatchReachZeroLine(matchInfo)
		h.fetchSeqEnd = matchInfo.Result.Matches[len(matchInfo.Result.Matches) - 1].MatchSeqNum
		if matches == nil {
			continue
        }
		h.SaveMatches(matches)
		log.Println("start",h.fetchSeqStart,"end",h.fetchSeqEnd)
		h.SaveHistory()
		continue
	}
}

func (h *AllHistory) InitDb() {
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
	h.db = db
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
}

func (h *AllHistory) SaveHistory() {
	stmtSaveHistory, err := h.db.Prepare("INSERT INTO MatchHistory VALUES( ?, ?, ?, ?)")

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	defer stmtSaveHistory.Close()

	_, err = stmtSaveHistory.Exec(h.fetchSeqStart, h.fetchSeqEnd, h.zeroSeq, 1)
}

func (h *AllHistory) SaveMatches(matches []MatchInfoMatch) {

	stmtIns, err := h.db.Prepare("INSERT INTO MatchWithPlayers VALUES( ?, ? )")

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	defer stmtIns.Close()

	var i int
	for _, m := range matches {
		var accountIds string
		valid := true
		for _, p := range m.Players {
			if p.AccountId == 4294967295 || p.HeroId == 0 {
				valid = false
				break
			}
			if accountIds == "" {
				accountIds = strconv.FormatInt(p.AccountId,10)
            }
			accountIds += (";" + strconv.FormatInt(p.AccountId,10))
		}
		if !valid {
			continue
        }
		i++
		log.Println("ID: ",m.MatchId,"Players: ",accountIds)
		_, err = stmtIns.Exec(m.MatchId, accountIds)

		if err != nil {
			log.Printf("%s\n", err)
			return
		}
	}
	log.Println("parse: ",len(matches),"save: ",i)
}

func (h *AllHistory) checkMatchReachZeroLine(matchInfo MatchInfo) []MatchInfoMatch {
	matches := make([]MatchInfoMatch,len(matchInfo.Result.Matches))
	if h.zeroSeq == 0 {
		for _, m := range matchInfo.Result.Matches {
			//log.Println(m)
			if m.StartTime <= ConfigData.limit6_86 {
				h.zeroSeq = m.MatchSeqNum
				break
			}
			matches = append(matches, m)
		}
		return matches
	}
	for _, m := range matchInfo.Result.Matches {
		//log.Println(m)
		if m.MatchSeqNum <= h.zeroSeq {
			h.fetchSeqEnd = h.fetchSeqStart
			h.zeroSeq = h.fetchSeqStart
			break
		}
		matches = append(matches, m)
	}
	return matches
}
