package picker

import (
	"fmt"
	"strings"
	"io/ioutil"
)

var ConfigData = &struct {
	key string
	Addr string
	nickNames map[string]string
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
