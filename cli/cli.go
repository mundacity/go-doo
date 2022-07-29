package cli

import (
	"strings"
	"time"

	godoo "github.com/mundacity/go-doo"
)

var CliContext godoo.ICliContext

// Feature in progress... TODO
// relates to priority queue and how remote storage returns the 'next' item
type priorityMode string

const (
	deadline priorityMode = "d"
	none     priorityMode = "n"
	low      priorityMode = "l"
	medium   priorityMode = "m"
	high     priorityMode = "h"
)

// if user is using a date range, get the upper bound of that range
func getUpperDateBound(dateText string, dateLayout string) time.Time {
	splt := splitDates(dateText)
	var d time.Time

	if len(splt) > 1 {
		d, _ = time.Parse(dateLayout, splt[1])
	}

	return d
}

func splitDates(s string) []string {
	return strings.Split(s, ":")
}
