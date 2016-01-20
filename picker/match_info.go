package picker

var getMatchDetails = "https://api.steampowered.com/IDOTA2Match_570/GetMatchDetails/v001/?"
var getMatchHistory = "https://api.steampowered.com/IDOTA2Match_570/GetMatchHistory/V001/?"
var getMatchHistoryFromDotamax = "http://dotamax.com/player/match/"

type MatchDetails struct {
	Result struct {
		Players [10]struct {
			AccountId  int `json:"account_id"`
			PlayerSlot int `json:"player_slot"`
			HeroId     int `json:"hero_id"`
		} `json:"players"`
		RadiantWin bool `json:"radiant_win"`
	} `json:"result"`
}

type MatchHistory struct {
	Result struct {
		Status     int
		NumResults int `json:"num_results"`
		Matches    []struct {
			MatchId int `json:"match_id"`
		} `json:"matches"`
	} `json:"result"`
}
