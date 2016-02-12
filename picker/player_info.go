package picker

import (
	"encoding/json"
	"strconv"
	"sync"
)

/*CREATE TABLE `PlayerInfo` (
	`AccountId` int(11) NOT NULL,
	`Info` MEDIUMTEXT NOT NULL,
	PRIMARY KEY (`AccountId`)
);*/
/*
INSERT INTO PlayerInfo VALUES (0,"")
*/

type PlayerInfo struct {
	MatchCount     int
	MaxMatchSeq    int
	HeroWins       map[string]int
	HeroCounts     map[string]int
	HeroBeatWins   map[string]int
	HeroBeatCounts map[string]int
	lock		   sync.Mutex
}

func (p *PlayerInfo) Init() {
	p.HeroWins = make(map[string]int)
	p.HeroCounts = make(map[string]int)
	p.HeroBeatWins = make(map[string]int)
	p.HeroBeatCounts = make(map[string]int)
}

func (p *PlayerInfo) updatePlayerInfo(accountId string, data []byte) int{
	var matchDetails MatchDetails
	json.Unmarshal(data, &matchDetails)
	//判断是否全部玩家选择英雄
	for _, player := range matchDetails.Result.Players {
		if player.HeroId == 0 {
			return -1
        }
    }
	var useName string
	var useWin bool
	var enemyBeat []string
	var hasUse bool
	for _, player := range matchDetails.Result.Players {
		//数据分析的玩家
		if strconv.Itoa(player.AccountId) == accountId {
			hasUse = true
			useName = strconv.Itoa(player.HeroId)
			//fmt.Println(strconv.Itoa(player.HeroId) + " " + HeroName)
			//天辉
			if player.PlayerSlot < 5 {
				var enemyHeroName string
				for _, enemyPlayer := range matchDetails.Result.Players[5:] {
					//夜宴英雄名字
					enemyHeroName = strconv.Itoa(enemyPlayer.HeroId)
					enemyBeat = append(enemyBeat,enemyHeroName)
				}
				if matchDetails.Result.RadiantWin == true {
					useWin = true
				}
			}
			//夜宴
			if player.PlayerSlot > 5 {
				var enemyHeroName string
				for _, enemyPlayer := range matchDetails.Result.Players[:5] {
					//天辉英雄名字
					enemyHeroName = strconv.Itoa(enemyPlayer.HeroId)
					enemyBeat = append(enemyBeat,enemyHeroName)
				}
				if matchDetails.Result.RadiantWin == false {
					useWin = true
				}
			}
			break;
		}
	}
	if !hasUse {
		return -1
    }
	p.lock.Lock()
	p.MatchCount++
	//英雄出场数+1
	p.HeroCounts[useName]++
	//对抗次数+1
	for _,enemyHeroName := range enemyBeat {
		p.HeroBeatCounts[MergeHeroName(useName, enemyHeroName)]++
    }
	if useWin {
		//获胜次数+1
		p.HeroWins[useName]++
		for _,enemyHeroName := range enemyBeat {
			//对抗获胜次数+1
			p.HeroBeatWins[MergeHeroName(useName, enemyHeroName)]++
		}
	}
	//fmt.Printf("MatchCount %d\n", p.MatchCount)
	p.lock.Unlock()
	return 0
}
