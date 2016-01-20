package picker

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"sort"
	"strings"
)

func httpGet(url string) []byte {
	fmt.Println("DEBUG-URL" + url)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body
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
