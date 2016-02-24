package picker

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

func httpGet(url string) []byte {
	defer log.Println("httpGet Finished")
	for i := 0; i != 20; i++ {
		log.Println("DEBUG-URL: " + url)
		c := http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					deadline := time.Now().Add(25 * time.Second)
					c, err := net.DialTimeout(netw, addr, time.Second*20)
					if err != nil {
						return nil, err
					}
					c.SetDeadline(deadline)
					return c, nil
				},
			},
		}
		resp, err := c.Get(url)
		if err != nil {
			if i != 0 {
				log.Println("retry get ", url, " ", i)
			}
			continue
		}
		if resp.StatusCode != 200 {
			if i != 0 {
				log.Println("retry get ", url, " ", i)
			}
			resp.Body.Close()
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil {
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
