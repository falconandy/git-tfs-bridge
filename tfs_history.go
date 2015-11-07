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
	affectedPaths []string
}

type TfsHistory []*TfsHistoryItem

func (history TfsHistory) Len() int {
	return len(history)
}

func (history TfsHistory) Less(i, j int) bool {
	return history[i].changeset < history[j].changeset
}

func (history TfsHistory) Swap(i, j int) {
	history[i], history[j] = history[j], history[i]
}

func (item *TfsHistoryItem) String() string {
	return fmt.Sprintf("CS%d %v %s\n%s\n%s\n", item.changeset, item.date, item.author, item.comment, strings.Join(item.affectedPaths, "\n"))
}

func (item *TfsHistoryItem) GetChangeset() int {
	return item.changeset
}

func (item *TfsHistoryItem) GetComment() string {
	return item.comment
}

func (item *TfsHistoryItem) GetAuthor() string {
	return item.author
}

func (item *TfsHistoryItem) GetDate() time.Time {
	return item.date
}

func parseHistory(history string, count int) []*TfsHistoryItem {
	const historyDelimiter string = "------------------------------"
	var changeset int
	var author string
	var comment string
	var date time.Time
	affectedPaths := make([]string, 0, 20)
	location, _ := time.LoadLocation("Local")
	result := make([]*TfsHistoryItem, 0, count)
	scanner := bufio.NewScanner(strings.NewReader(history))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, historyDelimiter) && strings.HasSuffix(line, historyDelimiter) {
			if changeset != 0 {
				result = append(result, &TfsHistoryItem{changeset, author, comment, date, affectedPaths})
			}
			changeset, author, comment, date, affectedPaths = 0, "", "", time.Time{}, make([]string, 0, 20)
		}
		if strings.HasPrefix(line, "Changeset:") && changeset == 0 {
			changeset, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Changeset:")))
		}
		if strings.HasPrefix(line, "User:") && author == "" {
			author = strings.TrimSpace(strings.TrimPrefix(line, "User:"))
		}
		if strings.HasPrefix(line, "Date:") && date.IsZero() {
			dateStr := strings.TrimSpace(strings.TrimPrefix(line, "Date:"))
			date, _ = parseMaybeRussianDate("2 January 2006 15:04:05", dateStr, location)
		}
		if line == "Comment:" && comment == "" {
			var buffer bytes.Buffer
			for scanner.Scan() {
				line := scanner.Text()
				if line == "Items:" {
					break
				}
				line = strings.TrimPrefix(line, "  ")
				buffer.WriteString(line)
				buffer.WriteString("\r\n")
			}
			comment = strings.TrimSpace(buffer.String())
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					break
				}
				line = strings.TrimPrefix(line, "  ")
				sepIndex := strings.Index(line, " $")
				if sepIndex >= 0 {
					affectedPaths = append(affectedPaths, line[sepIndex+1:len(line)])
				}
			}
		}
	}
	return result
}