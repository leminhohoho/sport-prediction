package models

import "time"

type Matches struct {
	Id           string
	Url          string
	Date         time.Time
	TournamentId string
	Stage        string
	Team1Id      string
	Team2Id      string
	TeamWon      string
}

type MatchesMaps struct {
	MatchId      string
	MapName      string
	Team1Score   string
	Team2Score   string
	TeamDefFirst string
	TeamAtkFirst string
}

type MatchesMapsTeamsPlayersPerformance struct {
	MatchId      string
	MapName      string
	TeamId       string
	PlayerName   string
	PlayerId     string
	AgentName    string
	Rating       string
	Acs          string
	Kills        string
	Deaths       string
	Assists      string
	Kast         string
	Adr          string
	HSPercentage string
	FirstKills   string
	FirstDeaths  string
	N2Kills      string
	N3Kills      string
	N4Kills      string
	N5Kills      string
	N1v1         string
	N1v2         string
	N1v3         string
	N1v4         string
	N1v5         string
}

type Teams struct {
	Id   string
	url  string
	name string
}
