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

func FetchMatchsBySeq(seq int) {
	
}

func (h *allHistory) fetchProcess(db *DB) {
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
			h.fetchSeqEnd = nowMatchIds[len(MatchIds) - 1].Seq
        }
	}
}
