package picker

import (
	"strconv"
	"encoding/json"
	"log"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type HeroInfo struct {
	HeroWins       map[string]int
	HeroCounts     map[string]int
	HeroBeatWins   map[string]int
	HeroBeatCounts map[string]int
}

/*CREATE TABLE `HeroInfo` (
	`Info` MEDIUMTEXT NOT NULL,
	`nouse` int(11) NOT NULL,
	PRIMARY KEY (`nouse`)
);*/
/*
INSERT INTO HeroInfo VALUES( "", 1)
*/

func (h *HeroInfo) SaveDb(db *sql.DB) {
	data, _ := json.Marshal(h)
		
	stmtIns, err := db.Prepare("UPDATE HeroInfo SET Info = ? WHERE nouse = 1")

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	defer stmtIns.Close()

	_, err = stmtIns.Exec(string(data))

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
}

func (h *HeroInfo) LoadDb(db *sql.DB) {
	h.HeroWins = make(map[string]int)
	h.HeroCounts = make(map[string]int)
	h.HeroBeatWins = make(map[string]int)
	h.HeroBeatCounts = make(map[string]int)
	stmtOut, err := db.Prepare("SELECT * FROM HeroInfo")

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

	var data string
	var nouse int
	for rows.Next() {
		err := rows.Scan(&data, &nouse)

		log.Println(data)
		if err != nil {
			log.Printf("%s\n", err)
        }	
		json.Unmarshal([]byte(data), h)
    }

	err = rows.Err()
	if err != nil {
		log.Printf("%s\n", err)
		return
    }
}

func (h *HeroInfo) updateHeroInfo(m *MatchInfoMatch){
	log.Println(m.MatchSeqNum)
	HeroId := [10]string{}
	for i,p := range m.Players {
		HeroId[i] = strconv.Itoa(p.HeroId)
	}
	for _,id := range HeroId {
		log.Println("old ",id, h.HeroCounts[id])
		h.HeroCounts[id]++
    }
	for _,id1 := range HeroId[0:5] {
		for _,id2 := range HeroId[5:10] {
			log.Println(MergeHeroName(id1,id2))
			h.HeroBeatCounts[MergeHeroName(id1,id2)]++
			if m.RadiantWin {
				h.HeroBeatWins[MergeHeroName(id1,id2)]++
			}
		}
    }
	for _,id2 := range HeroId[5:10] {
		for _,id1 := range HeroId[0:5] {
			log.Println(MergeHeroName(id2,id1))
			h.HeroBeatCounts[MergeHeroName(id2,id1)]++
			if !m.RadiantWin {
				h.HeroBeatWins[MergeHeroName(id2,id1)]++
			}
		}
    }
	if m.RadiantWin {
		for _,id := range HeroId[0:5] {
			log.Println("Radiant old ",id, h.HeroWins[id])
			h.HeroWins[id]++
		}
    }else {
		for _,id := range HeroId[5:10] {
			log.Println("Dire old ",id, h.HeroWins[id])
			h.HeroWins[id]++
		}
    }
}

func (h *HeroInfo) showHeroInfo(heroList []string) map[string]float32{
	heroBeatWinRate := make(map[string]map[string]float32)
	heroWinRate := make(map[string]float32)
	
	for Id,_ := range HeroIdStrMap {
		heroWinRate[Id] = float32(h.HeroWins[Id]) / float32(h.HeroCounts[Id])	
    }
	for Id1,_ := range HeroIdStrMap {
		for Id2,_ := range HeroIdStrMap {
			if Id1 == Id2 {
				continue
            }
			name := MergeHeroName(Id1,Id2)
			if heroBeatWinRate[Id2] == nil {
				heroBeatWinRate[Id2] = make(map[string]float32)
			}
			if h.HeroBeatCounts[name] != 0{
				winRate := float32(h.HeroBeatWins[name]) / float32(h.HeroBeatCounts[name])
				heroBeatWinRate[Id2][Id1] = winRate - heroWinRate[Id1]
            }else {
				heroBeatWinRate[Id2][Id1] = 0
            }
		}
	}
	choiceHeroRateMap := make(map[string]float32)
	for Id,_ := range HeroIdStrMap {
		for _, targetHeroName := range heroList {
			choiceHeroRateMap[Id] += heroBeatWinRate[targetHeroName][Id]
		}
		choiceHeroRateMap[Id] /= float32(len(heroList))
	}
	return choiceHeroRateMap
}
