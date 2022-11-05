package lichess

import "fmt"

func (lc *LichessConnector) ShouldAccept(ch Challenge) (bool, string) {
	if !lc.acceptingChallenges() {
		return false, "generic"
	}
	if ch.Challenger.Title == "BOT" && !lc.acceptingBotChallenges() {
		return false, "noBot"
	}
	if !lc.acceptableVariant(ch.Variant.Key) {
		return false, "variant"
	}
	if !lc.acceptableTimeControl(ch.Speed) {
		return false, "timeControl"
	}
	if !lc.acceptRated(ch.Rated) {
		return false, "casual"
	}

	return true, ""
}

func (lc *LichessConnector) acceptingChallenges() bool {
	return lc.Config.Lichess.ChallengePolicy.Accept
}

func (lc *LichessConnector) acceptingBotChallenges() bool {
	return lc.Config.Lichess.ChallengePolicy.AcceptBot
}

func (lc *LichessConnector) acceptableVariant(variant string) bool {
	for _, v := range lc.Config.Lichess.ChallengePolicy.Variant {
		if v == variant {
			return true
		}
	}
	return false
}

func (lc *LichessConnector) acceptableTimeControl(timeControl string) bool {
	for _, tc := range lc.Config.Lichess.ChallengePolicy.TimeControl {
		if tc == timeControl {
			return true
		}
	}
	fmt.Println(timeControl)
	return false
}

func (lc *LichessConnector) acceptRated(rated bool) bool {
	return !rated || lc.Config.Lichess.ChallengePolicy.Rated && rated
}
