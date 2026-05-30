package pages

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

type VoteBox struct {
	Value   string
	Voters  []string
	Outlier bool
}

func isRevealed(until time.Time) bool {
	if until.IsZero() {
		return false
	}
	return time.Since(until) > 1500*time.Millisecond
}

func voteBoxes(cards []string, allVotes map[string]string) []VoteBox {
	byCard := map[string][]string{}
	for name, val := range allVotes {
		byCard[val] = append(byCard[val], name)
	}
	minIdx, maxIdx := -1, -1
	for i, c := range cards {
		if len(byCard[c]) > 0 {
			if minIdx == -1 {
				minIdx = i
			}
			maxIdx = i
		}
	}
	if minIdx == -1 {
		return nil
	}
	totalVoters := len(allVotes)
	hasMajority := false
	for _, voters := range byCard {
		if len(voters) >= 2 {
			hasMajority = true
			break
		}
	}
	boxes := make([]VoteBox, 0, maxIdx-minIdx+1)
	for i := minIdx; i <= maxIdx; i++ {
		voters := byCard[cards[i]]
		outlier := totalVoters > 2 && hasMajority && len(voters) == 1 && (i == minIdx || i == maxIdx)
		boxes = append(boxes, VoteBox{Value: cards[i], Voters: voters, Outlier: outlier})
	}
	return boxes
}

func outlierBoxClass(outlier bool) string {
	if outlier {
		return "bg-error/10 border-error/60"
	}
	return "bg-base-200 border-base-300"
}

func outlierValueClass(outlier bool) string {
	if outlier {
		return "text-error"
	}
	return ""
}

func cardSelectClass(card, myVote string) string {
	if card == myVote {
		return "btn-primary"
	}
	return "btn-outline"
}

func cardVoteVals(card string) string {
	b, _ := json.Marshal(map[string]string{"value": card})
	return string(b)
}

func computeCountdown(until time.Time) (label string, active bool, untilMs string) {
	if until.IsZero() {
		return "Countdown", false, "0"
	}
	remaining := time.Until(until).Seconds()
	if remaining > 0 {
		secs := int(math.Ceil(remaining))
		if secs < 1 {
			secs = 1
		}
		if secs > 3 {
			secs = 3
		}
		return strconv.Itoa(secs), true, strconv.FormatInt(until.UnixMilli(), 10)
	}
	if remaining > -1.5 {
		return "Pointed!", true, strconv.FormatInt(until.UnixMilli(), 10)
	}
	return "Countdown", false, "0"
}

func isOkNoHelpCards(cards []string) bool {
	if len(cards) != 3 {
		return false
	}
	return cards[0] == "No." && cards[1] == "Help" && cards[2] == "OK"
}

func seatStyle(i, n int) string {
	if n == 0 {
		return ""
	}
	angle := 360.0 / float64(n) * float64(i)
	return fmt.Sprintf(
		"transform: translate(-50%%, -50%%) rotate(%.4fdeg) translate(0, -180px) rotate(-%.4fdeg);",
		angle, angle,
	)
}
