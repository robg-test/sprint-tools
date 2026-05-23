package pages

import "strconv"

func participantCountStr(participants []string) string {
	return strconv.Itoa(len(participants))
}
