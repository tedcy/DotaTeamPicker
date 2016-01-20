package picker

import (
	"encoding/json"
	"strconv"
)

type PlayerInfo struct {
	MatchCount     int
	MaxMatchId     int
	HeroWins       map[string]int
	HeroCounts     map[string]int
	HeroBeatWins   map[string]int
	HeroBeatCounts map[string]int
}

func (p *PlayerInfo) Init() {
	p.HeroWins = make(map[string]int)
	p.HeroCounts = make(map[string]int)
	p.HeroBeatWins = make(map[string]int)
	p.HeroBeatCounts = make(map[string]int)
}

func (p *PlayerInfo) updatePlayerInfo(accountId string, data []byte) {
	if p.HeroWins == nil {
		p.Init()
	}

	var matchDetails MatchDetails
	json.Unmarshal(data, &matchDetails)
	//判断是否全部玩家选择英雄
	for _, player := range matchDetails.Result.Players {
		if player.HeroId == 0 {
			return
        }
    }
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
					p.HeroBeatCounts[MergeHeroName(HeroName, enemyHeroName)]++
					if matchDetails.Result.RadiantWin == true {
						//对抗获胜次数+1
						p.HeroBeatWins[MergeHeroName(HeroName, enemyHeroName)]++
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
					p.HeroBeatCounts[MergeHeroName(HeroName, enemyHeroName)]++
					if matchDetails.Result.RadiantWin == false {
						//对抗获胜次数+1
						p.HeroBeatWins[MergeHeroName(HeroName, enemyHeroName)]++
					}
				}
				if matchDetails.Result.RadiantWin == false {
					//获胜次数+1
					p.HeroWins[HeroName]++
				}
			}
		}
	}
}
