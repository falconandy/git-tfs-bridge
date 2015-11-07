package git_tfs_bridge_test

import (
	"testing"
	bridge "github.com/falconandy/git-tfs-bridge"
	"log"
	"github.com/sabhiram/go-git-ignore"
	"sort"
	"io/ioutil"
	"path/filepath"
	"strings"
	"fmt"
	"os"
)

func TestHistory(t *testing.T) {
	tfs := bridge.OpenTfsRepository(`D:\Projects\Sungero\Main\Common`)
	log.Println(tfs.GetHistoryAfter(98000))
	log.Println(tfs.GetHistory(50000, 100))
}

func TestUpdate(t *testing.T) {
	tfs := bridge.OpenTfsRepository(`D:\Projects\Sungero\Main\Common`)
	tfs.Update(95000)
	tfs.Update(0)
}

func TestTraverseDirectory(t *testing.T) {
	gitIgnore, err := ignore.CompileIgnoreFile(`D:\Projects\_sun\Main\.gitignore`)
	if err != nil {
		log.Println(err)
	} else {
		bridge.TraverseDirectory(`D:\Projects\_sun\Main\`, gitIgnore)
	}
}

func TestCompareDirectory(t *testing.T) {
	gitIgnore, _ := ignore.CompileIgnoreFile(`D:\Projects\_sun\Main\.gitignore`)
	leftOnly, rightOnly, diffs := bridge.CompareDirectories(`D:\Projects\_sun\Main\Common`, `D:\Projects\Sungero\Main\Common`, gitIgnore)
	log.Println(leftOnly)
	log.Println(rightOnly)
	log.Println(diffs)
}

func TestToGit(t *testing.T) {
	tfs := bridge.OpenTfsRepository(`D:\Projects\Sungero\Main\Common`)
	git, _ := bridge.OpenGitRepository(`D:\Projects\_sun\Main`)
	gitIgnore, _ := ignore.CompileIgnoreFile(`D:\Projects\_sun\Main\.gitignore`)
	history := bridge.TfsHistory(tfs.GetHistoryAfter(98866))
	sort.Sort(history)
	for _, historyItem := range history {
		log.Println(historyItem.GetChangeset())
		tfs.Update(historyItem.GetChangeset())
		leftOnly, rightOnly, diffs := bridge.CompareDirectories(`D:\Projects\_sun\Main\Common`, `D:\Projects\Sungero\Main\Common`, gitIgnore)
		if len(leftOnly) + len(rightOnly) + len(diffs) == 0 {
			continue
		}
		for path := range leftOnly {
			os.Remove(filepath.Join(`D:\Projects\_sun\Main\Common`, path))
		}
		for path := range rightOnly {
			content, err := ioutil.ReadFile(filepath.Join(`D:\Projects\Sungero\Main\Common`, path))
			if err != nil {
				log.Println(err)
				continue
			}
			err = ioutil.WriteFile(filepath.Join(`D:\Projects\_sun\Main\Common`, path), content, 0666)
			if err != nil {
				log.Println(err)
				continue
			}
		}
		for path := range diffs {
			content, err := ioutil.ReadFile(filepath.Join(`D:\Projects\Sungero\Main\Common`, path))
			if err != nil {
				log.Println(err)
				continue
			}
			err = ioutil.WriteFile(filepath.Join(`D:\Projects\_sun\Main\Common`, path), content, 0666)
			if err != nil {
				log.Println(err)
				continue
			}
		}
		git.StageAll()
		comment, author, date := historyItem.GetComment(), historyItem.GetAuthor(), historyItem.GetDate()
		author = strings.TrimPrefix(author, `NT_WORK\`)
		git.Commit(comment, fmt.Sprintf("%s <%s@directum.ru>", author, author), date)
	}
}