package main

import "football_bot/internal/footballdata"

func MatchStatus(
	match *footballdata.Match,
) string {

	switch {

	case match.Status == "FINISHED":
		return "Full Time"

	case match.Score.Duration == "PENALTY_SHOOTOUT":
		return "Penalties"

	case match.Score.Duration == "EXTRA_TIME":
		return "Extra Time"

	case match.Status == "PAUSED":
		return "Half Time"

	case match.Status == "IN_PLAY":
		return "Live"

	case match.Status == "TIMED":
		return "Not Started"

	default:
		return match.Status
	}
}
