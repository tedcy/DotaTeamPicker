package DotaTeamPicker

//https://api.steampowered.com/IDOTA2Match_570/GetMatchDetails/v001/?match_id=2069245018
//https://api.steampowered.com/IDOTA2Match_570/GetMatchHistory/V001/?account_id=144725945&matches_requested=1

type Player struct {                                                               
    Account_id string                                                              
    Player_slot int                                                                
    Hero_id int                                                                    
}                                                                                  
                                                                                   
type Result struct {                                                               
    Players [10]Player                                                             
    Radiant_win string                                                             
}                                                                                  
