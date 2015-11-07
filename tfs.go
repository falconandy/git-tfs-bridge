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
	"sort"
	"bufio"
	"strings"
)

type TfsRepository struct {
	path string
	workfold string
}

func OpenTfsRepository(path string) *TfsRepository {
	workfold := getWorkfold(path)
	return &TfsRepository{path:path, workfold:workfold}
}

func getWorkfold(path string) string {
	output, err := execTfCommand("workfold", path)
	if err != nil {
		log.Println(err)
		return ""
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
	return workfold
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

func (repo *TfsRepository) GetHistory(fromChangeset int, count int) []*TfsHistoryItem {
	var history []*TfsHistoryItem
	commandArgs := []string { "history", repo.path, "/recursive", "/noprompt", "/format:Detailed", fmt.Sprintf("/stopafter:%d", count) }
	if fromChangeset > 0 {
		commandArgs = append(commandArgs, fmt.Sprintf("/version:C%d", fromChangeset))
	}
	output, err := execTfCommand(commandArgs...)
	if err != nil {
		log.Println(err)
	} else {
		history = parseHistory(repo.workfold, ansi2utf8(output), count)
	}
	return history
}

func (repo *TfsRepository) GetHistoryAfter(changeset int) TfsHistory {
	var result TfsHistory
	fromChangeset := 0
	for {
		history := repo.GetHistory(fromChangeset, 100)
		if len(history) == 0 {
			break
		} else if history[0].changeset <= changeset {
			break
		} else if history[len(history)-1].changeset > changeset {
			result = append(result, history...)
			fromChangeset = history[len(history)-1].changeset - 1
			continue
		} else {
			for _, histItem := range history {
				if histItem.changeset > changeset {
					result = append(result, histItem)
				} else {
					break
				}
			}
			break
		}
	}
	sort.Sort(result)
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
