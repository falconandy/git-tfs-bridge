package git_tfs_bridge

import (
	"time"
	"strings"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"os"
	"github.com/sabhiram/go-git-ignore"
)

var russian2englishMonths = map[string]string {
	"января": "January",
	"февраля": "February",
	"марта": "March",
	"апреля": "April",
	"мая": "May",
	"июня": "June",
	"июля": "July",
	"августа": "August",
	"сентября": "September",
	"октября": "October",
	"ноября": "November",
	"декабря": "December",
}

func parseMaybeRussianDate(layout string, value string, location *time.Location) (time.Time, error) {
	value = strings.Replace(value, " г.", "", 1)
	for rus, eng := range russian2englishMonths {
		if strings.Contains(value, rus) {
			value = strings.Replace(value, rus, eng, 1)
			break
		}
	}
	return time.ParseInLocation(layout, value, location)
}

func TraverseDirectory(root string, gitIgnore *ignore.GitIgnore) map[string]string {
	traverseInfo := make(map[string]string)
	filepath.Walk(root, func (path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == ".git" || info.Name() == ".gitignore" || info.Name() == ".hg" || info.Name() == ".hgignore" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		relativePath := strings.TrimLeft(strings.TrimPrefix(path, root), string(filepath.Separator))
		if info.IsDir() {
			relativePath += string(filepath.Separator)
		}
		if gitIgnore.MatchesPath(relativePath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			content, err := ioutil.ReadFile(path)
			if err == nil {
				traverseInfo[relativePath] = fmt.Sprintf("%x", md5.Sum(content))
			} else {
				traverseInfo[relativePath] = "ERROR"
			}
		}
		return nil
	})
	return traverseInfo
}

func CompareDirectories(leftPath string, rightPath string, gitIgnore *ignore.GitIgnore) (leftOnly map[string]struct{}, rightOnly map[string]struct{}, diffs map[string]struct{}) {
	left := TraverseDirectory(leftPath, gitIgnore)
	right := TraverseDirectory(rightPath, gitIgnore)
	leftOnly = make(map[string]struct{})
	rightOnly = make(map[string]struct{})
	diffs = make(map[string]struct{})
	for leftPath, leftHash := range left {
		if rightHash, ok := right[leftPath]; ok {
			if rightHash != leftHash {
				diffs[leftPath] = struct{}{}
			}
			delete(left, leftPath)
			delete(right, leftPath)
		} else {
			leftOnly[leftPath] = struct{}{}
			delete(left, leftPath)
		}
	}
	for rightPath, _ := range right {
		rightOnly[rightPath] = struct{}{}
	}
	return leftOnly, rightOnly, diffs
}