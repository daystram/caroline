package util

import (
	"fmt"

	"github.com/daystram/carol/internal/domain"
)

func FormatQueue(q *domain.Queue) string {
	if q.CurrentTrack == -1 || len(q.Tracks) == 0 {
		return "```\nNothing is playing!\n```"
	}

	out := "```\n"
	for i, t := range q.Tracks {
		if i == q.CurrentTrack {
			out += fmt.Sprintf("[%d] ", i+1)
		} else {
			out += fmt.Sprintf(" %d  ", i+1)
		}
		out += fmt.Sprintf("%16.16s    [@%s]\n", t.Query, t.QueuedByUsername)
	}
	out += "```"
	return out
}
