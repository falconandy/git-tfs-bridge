package git_tfs_bridge

import (
	"os/exec"
	"os"
	"fmt"
	"io/ioutil"
	"bytes"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/charmap"
	"log"
	"bufio"
	"strings"
	"github.com/sabhiram/go-git-ignore"
)

type TfsRepository struct {
	path string
	workfold string
}

func (repo *TfsRepository) GetPath() string {
	return repo.path
}

func OpenTfsRepository(path string) (*TfsRepository, error) {
	workfold, err := getWorkfold(path)
	if err != nil {
		return nil, err
	}
	return &TfsRepository{path:path, workfold:workfold}, nil
}

func getWorkfold(path string) (string, error) {
	output, err := execTfCommand("workfold", path)
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	workfold := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, " $") {
			sepIndex := strings.Index(line, ": ")
			workfold = strings.TrimSpace(line[:sepIndex])
			break
		}
	}
	return workfold, nil
}

func (repo *TfsRepository) Update(changeset int) {
	commandArgs := []string { "get", repo.path, "/recursive", "/noprompt", "/overwrite" }
	if changeset > 0 {
		commandArgs = append(commandArgs, fmt.Sprintf("/version:C%d", changeset))
	}
	_, err := execTfCommand(commandArgs...)
	if err != nil {
		log.Println(err)
	}
}

func (repo *TfsRepository) IsClean() bool {
	commandArgs := []string { "status", repo.path, "/recursive", "/format:brief" }
	output, err := execTfCommand(commandArgs...)
	if err != nil {
		log.Println(err)
		return false
	}
	return strings.TrimSpace(ansi2utf8(output)) == "There are no pending changes."
}

func (repo *TfsRepository) GetHistory(gitIgnore *ignore.GitIgnore, fromChangeset int, count int) []*TfsHistoryItem {
	var history []*TfsHistoryItem
	commandArgs := []string { "history", repo.path, "/recursive", "/noprompt", "/format:Detailed", fmt.Sprintf("/stopafter:%d", count) }
	if fromChangeset > 0 {
		commandArgs = append(commandArgs, fmt.Sprintf("/version:C%d", fromChangeset))
	}
	output, err := execTfCommand(commandArgs...)
	if err != nil {
		log.Println(err)
	} else {
		history = parseHistory(repo, gitIgnore, ansi2utf8(output), count)
	}
	return history
}

func (repo *TfsRepository) GetHistoryFrom(gitIgnore *ignore.GitIgnore, startChangeset int, includeStartChangeset bool) []*TfsHistoryItem {
	var result []*TfsHistoryItem
	fromChangeset := 0
	for {
		history := repo.GetHistory(gitIgnore, fromChangeset, 100)
		if len(history) == 0 {
			break
		} else if history[0].changeset < startChangeset || (!includeStartChangeset && history[0].changeset == startChangeset) {
			break
		} else if history[len(history)-1].changeset > startChangeset || (includeStartChangeset && history[len(history)-1].changeset == startChangeset) {
			result = append(result, history...)
			fromChangeset = history[len(history)-1].changeset - 1
			continue
		} else {
			for _, histItem := range history {
				if histItem.changeset > startChangeset || (includeStartChangeset && histItem.changeset == startChangeset) {
					result = append(result, histItem)
				} else {
					break
				}
			}
			break
		}
	}
	return result
}

func execTfCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("tf", args...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func ansi2utf8(input []byte) string {
	sr := bytes.NewReader(input)
	tr := transform.NewReader(sr, charmap.Windows1251.NewDecoder())
	buf, err := ioutil.ReadAll(tr)
	if err != err {
		return ""
	}
	return string(buf)
}
