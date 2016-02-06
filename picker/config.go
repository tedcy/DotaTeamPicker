package picker

import (
	"fmt"
	"strings"
	"strconv"
	"io/ioutil"
)

var ConfigData = &struct {
	key string
	Addr string
	testFetchMatches bool
	mysqlAddr string
	nickNames map[string]string
	limit6_86 int64
}{}

func LoadConfig() {
	data, _ := ioutil.ReadFile("config")
	ConfigData.nickNames = make(map[string]string)
	params := strings.Split(string(data),"\n")
	for _, param := range params {
		if len(param) == 0 {
			break
        }
		value := strings.Split(param," ")
		if value[0] == "key" {
			ConfigData.key = value[1]
        }
		if value[0] == "addr" {
			ConfigData.Addr = value[1]
        }
		if value[0] == "testFetchMatches" {
			if value[1] == "open" {
				ConfigData.testFetchMatches = true
            }
        }
		if value[0] == "mysqlAddr" {
			ConfigData.mysqlAddr = value[1]
        }
		if value[0] == "limit6.86" {
			ConfigData.limit6_86, _ = strconv.ParseInt(value[1],10,64)
        }
		if value[0] == "nickNames" {
			nameStrs := strings.Split(value[1],";")
			for _, nameStr := range nameStrs {
				names := strings.Split(nameStr,",")
				ConfigData.nickNames[names[0]] = names[1]
            }
		}
    }
	fmt.Println(ConfigData)
}
