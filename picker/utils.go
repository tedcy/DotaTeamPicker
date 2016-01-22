package picker

import (
	"net/http"
	"io/ioutil"
	"sort"
	"strings"
)

func httpGet(url string) []byte {
	for i := 0;i != 20; i++ {
		//log.Println("DEBUG-URL:" + url)
		resp, err := http.Get(url)
		if err != nil {
			if i != 0 {
				//log.Println("retry get ",url," ",i,"error")
            }
			continue
        }
		if resp.StatusCode != 200{
			if i != 0 {
				//log.Println("retry get ",url," ",i,"error")
            }
			resp.Body.Close()
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil{
			return body
		}
		return nil
    }
	return nil
}

func MapSort(p map[string]float32) PairList {
	pl := make(PairList, len(p))
	i := 0
	for k, v := range p {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value float32
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }


func MergeHeroName(heroName string, enemyHeroName string) string {
	return heroName + "-" + enemyHeroName
}

func SplitHeroName(name string) (string, string) {
	s := strings.Split(name, "-")
	return s[0], s[1]
}
