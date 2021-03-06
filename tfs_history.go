package git_tfs_bridge

import (
	"time"
	"strings"
	"bytes"
	"strconv"
	"bufio"
	"fmt"
	"path/filepath"
	"github.com/sabhiram/go-git-ignore"
)

type TfsHistoryItem struct {
	changeset int
	author string
	comment string
	date time.Time
	affectedPaths []string
	repo *TfsRepository
}

func (item *TfsHistoryItem) String() string {
	return fmt.Sprintf("CS%d %v %s\n%s\n%s\n", item.changeset, item.date, item.author, item.comment, strings.Join(item.affectedPaths, "\n"))
}

func (item *TfsHistoryItem) IsAffected(path string) bool {
	for _, affectedPath := range item.affectedPaths {
		if strings.Contains(affectedPath, path) {
			return true
		}
	}
	return false
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

func (item *TfsHistoryItem) GetRepo() *TfsRepository {
	return item.repo
}

func parseHistory(repo *TfsRepository, gitIgnore *ignore.GitIgnore, history string, count int) []*TfsHistoryItem {
	workfold := repo.workfold + `/`
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
				result = append(result, &TfsHistoryItem{changeset, author, comment, date, affectedPaths, repo})
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
					affectedPath := line[sepIndex+1:len(line)]
					if strings.HasPrefix(affectedPath, workfold) {
						affectedPath = strings.TrimPrefix(affectedPath, workfold)
						affectedPath = strings.Replace(affectedPath, "/", string(filepath.Separator), -1)
						sepIndex = strings.Index(affectedPath, ";")
						if sepIndex >= 0 {
							affectedPath = affectedPath[:sepIndex]
						}
						if !gitIgnore.MatchesPath(affectedPath) {
							affectedPaths = append(affectedPaths, affectedPath)
						}
					}
				}
			}
		}
	}
	if changeset != 0 {
		result = append(result, &TfsHistoryItem{changeset, author, comment, date, affectedPaths, repo})
	}
	return result
}