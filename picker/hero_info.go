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
	HeroId := [10]string{}
	for i,p := range m.Players {
		HeroId[i] = strconv.Itoa(p.HeroId)
	}
	for _,id := range HeroId {
		h.HeroCounts[id]++
    }
	for _,id1 := range HeroId[0:4] {
		for _,id2 := range HeroId[5:9] {
			h.HeroBeatCounts[MergeHeroName(id1,id2)]++
			if m.RadiantWin {
				h.HeroBeatWins[MergeHeroName(id1,id2)]++
			}
		}
    }
	if m.RadiantWin {
		for _,id := range HeroId[0:4] {
			h.HeroCounts[id]++
		}
    }else {
		for _,id := range HeroId[5:9] {
			h.HeroCounts[id]++
		}
    }
}
