package picker

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type allHistory struct {
	fetchSeqStart int
	fetchSeqEnd int
	zeroSeq int
}

type matchInfo struct {
	Result struct {
		Matches    []struct {
			Players [10]struct {
				AccountId int `json:"account_id"`
				HeroId int `json:"hero_id"`
			}`json:"players"`
			MatchId int `json:"match_id"`
			MatchSeqNum int `json:"match_seq_num"`
		} `json:"matches"`
	} `json:"result"`
}

var getMatchHistoryWithSeq = "https://api.steampowered.com/IDOTA2Match_570/GetMatchHistoryBySequenceNum/V001/?start_at_match_seq_num=1866720879"

func FetchMatchsBySeq(seq int) {
	
}

func (h *allHistory) fetchProcess(db *DB) {
	//历史记录已经从sql里加载完毕
	for {
		if h.fetchSeqEnd == 0 {
			nowMatchIds := FetchMatchsBySeq(0)
			h.fetchSeqEnd = h.fetchSeqStart = nowMatchIds[0]
			time.Sleep(time.Second)
			continue
        }
		//历史记录未抓取完整
		if zeroSeq == 0 {
			MatchIds := FetchMatchsBySeq(h.fetchSeqEnd)
			//检查到686的时间戳
			//如果达到时间戳，那么记录下该matchId
			checkMatchReachZeroLine(MatchIds,zeroSeq)
			SaveDb(DB,MatchIds)
			h.fetchSeqEnd = nowMatchIds[len(MatchIds) - 1].Seq
			continue
        }
		//抓取最新记录
		MatchIds := FetchMatchsBySeq(h.fetchSeqEnd)
		//检查是否到zeroSeq
		//如果达到，那么fetchSeqEnd = fetchSeqStart
		checkMatchReachZeroLine(MatchIds,zeroSeq)
		SaveDb(DB,MatchIds)
		h.fetchSeqEnd = nowMatchIds[len(MatchIds) - 1].Seq
		continue
	}
}
