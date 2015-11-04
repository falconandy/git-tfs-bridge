package git_tfs_bridge

import (
	"time"
	"strings"
	"bytes"
	"strconv"
	"bufio"
	"fmt"
)

type TfsHistoryItem struct {
	changeset int
	author string
	comment string
	date time.Time
}

func (item TfsHistoryItem) String() string {
	return fmt.Sprintf("CS%d %v %s\n%s", item.changeset, item.date, item.author, item.comment)
}

func parseHistory(history string, count int) []*TfsHistoryItem {
	const historyDelimiter string = "------------------------------"
	var changeset int
	var author string
	var comment string
	var date time.Time
	result := make([]*TfsHistoryItem, 0, count)
	scanner := bufio.NewScanner(strings.NewReader(history))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, historyDelimiter) && strings.HasSuffix(line, historyDelimiter) {
			if changeset != 0 {
				result = append(result, &TfsHistoryItem{changeset, author, comment, date})
			}
			changeset, author, comment, date = 0, "", "", time.Time{}
		}
		if strings.HasPrefix(line, "Changeset:") && changeset == 0 {
			changeset, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Changeset:")))
		}
		if strings.HasPrefix(line, "User:") && author == "" {
			author = strings.TrimSpace(strings.TrimPrefix(line, "User:"))
		}
		if strings.HasPrefix(line, "Date:") && date.IsZero() {
			dateStr := strings.TrimSpace(strings.TrimPrefix(line, "Date:"))
			date, _ = parseMaybeRussianDate("2 January 2006 15:04:05", dateStr)
		}
		if line == "Comment:" && comment == "" {
			var buffer bytes.Buffer
			for scanner.Scan() {
				line := scanner.Text()
				if line == "Items:" {
					break
				}
				buffer.WriteString(line)
				buffer.WriteString("\r\n")
			}
			comment = strings.TrimSpace(buffer.String())
		}
	}
	return result
}