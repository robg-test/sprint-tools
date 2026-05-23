package pages

import "strconv"

func participantCountStr(participants []string) string {
	return strconv.Itoa(len(participants))
}

func totalCountStr(players, watchers []string) string {
	return strconv.Itoa(len(players) + len(watchers))
}
