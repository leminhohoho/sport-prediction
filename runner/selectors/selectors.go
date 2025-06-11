package selectors

const (
	VLRggIndicator          = `body > header > nav > a.header-logo > img`
	MatchDate               = "#wrapper > div.col-container > div > div.wf-label.mod-large"
	MatchTournamentAnchor   = `#wrapper > div.col-container > div.col.mod-3 > div.wf-card.match-header > div.match-header-super > div:nth-child(1) > a`
	MatchHeader             = `#wrapper > div.col-container > div.col.mod-3 > div.wf-card.match-header > div.match-header-super > div:nth-child(1) > a > div > div.match-header-event-series`
	MatchTeam1Anchor        = `#wrapper > div.col-container > div.col.mod-3 > div.wf-card.match-header > div.match-header-vs > a.match-header-link.wf-link-hover.mod-1`
	MatchTeam2Anchor        = `#wrapper > div.col-container > div.col.mod-3 > div.wf-card.match-header > div.match-header-vs > a.match-header-link.wf-link-hover.mod-2`
	MatchMapGenericSelector = `#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div[data-game-id!="all"].vm-stats-game`
	Team1OverallResult      = `#wrapper > div.col-container > div.col.mod-3 > div.wf-card.match-header > div.match-header-vs > div > div.match-header-vs-score > div:nth-child(1) > span:nth-child(1)`
	Team2OverallResult      = `#wrapper > div.col-container > div.col.mod-3 > div.wf-card.match-header > div.match-header-vs > div > div.match-header-vs-score > div:nth-child(1) > span:nth-child(3)`

	TournamentName           = `#wrapper > div.col-container > div > div.wf-card.mod-event.mod-header.mod-full > div.event-header > div.event-desc > div > h1`
	TournamentPrize          = `#wrapper > div.col-container > div > div.wf-card.mod-event.mod-header.mod-full > div.event-header > div.event-desc > div > div > div:nth-child(2) > div.event-desc-item-value`
	TournamentTeamsContainer = `#wrapper > div.col-container > div > div.event-container > div.event-content > div.event-teams-container`
)
