package models

type CrystalLeagueMatches []CrystalLeagueMatch

type CrystalLeagueMatch struct {
	CrystalLeagueLevel int       `json:"CrystalLeagueLevel"`
	MatchID            string    `json:"MatchId"`
	MatchTime          string    `json:"MatchTime"`
	Team1              TeamStats `json:"Team1"`
	Team2              TeamStats `json:"Team2"`
	Winner             int       `json:"Winner"`
}

type TeamStats struct {
	Players []Player `json:"Players"`
	Tickets int      `json:"Tickets"`
	// Timeline []TimelineEvent `json:"Timeline"`
}

type Player struct {
	Deaths  int    `json:"Deaths"`
	Fame    int64  `json:"Fame"`
	Healing int    `json:"Healing"`
	Kills   int    `json:"Kills"`
	Name    string `json:"Name"`
}

// type TimelineEvent struct {
// 	EventType interface{} `json:"EventType"`
// 	Tickets   int         `json:"Tickets"`
// 	TimeStamp time.Time   `json:"TimeStamp"`
// }
