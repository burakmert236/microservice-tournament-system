package events

const (
	// Streams
	UserEventsStream       = "USER_EVENTS"
	TournamentEventsStream = "TOURNAMENT_EVENTS"

	// Events
	UserCreated = "events.user.created"
	UserLevelUp = "events.user.levelUp"

	TournamentParticipationScoreUpdated = "events.tournament.participationScoreUpdated"
	TournamentEntered                   = "events.tournament.entered"

	// Event Wildcards
	UserEventsWildcard       = "events.user.*"
	TournamentEventsWildcard = "events.tournament.*"
)
