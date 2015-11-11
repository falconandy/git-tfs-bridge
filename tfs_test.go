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
	log.Println(tfs)
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
		bridge.TraverseDirectory(`D:\Projects\_sun\Main\`, gitIgnore, nil)
	}
}

func TestCompareDirectory(t *testing.T) {
	gitIgnore, _ := ignore.CompileIgnoreFile(`D:\Projects\_sun\Main\.gitignore`)
	leftOnly, rightOnly, diffs := bridge.CompareDirectories(`D:\Projects\_sun\Main\Common`, `D:\Projects\Sungero\Main\Common`, gitIgnore, nil)
	log.Println(leftOnly)
	log.Println(rightOnly)
	log.Println(diffs)
}

func TestToGit(t *testing.T) {
	start_changeset := 98866
	tfs_root_folder := `D:\Projects\Sungero\Main\`
	git_root_folder := `D:\Projects\_sun\Main`
	tfs_subfolders := []string { "Common", "Kernel", "Content", "Report", "Workflow", "SDS" }
	tfs_reps := make([]*bridge.TfsRepository, len(tfs_subfolders))
	for i, tfs_subfolder := range tfs_subfolders {
		tfs_reps[i] = bridge.OpenTfsRepository(filepath.Join(tfs_root_folder, tfs_subfolder))
	}
	git, _ := bridge.OpenGitRepository(git_root_folder)
	gitIgnore, _ := ignore.CompileIgnoreFile(filepath.Join(git_root_folder, ".gitignore"))
	all_history := make(map[int][]*bridge.TfsHistoryItem)
	for _, tfs := range tfs_reps {
		history := tfs.GetHistoryAfter(start_changeset)
		for _, item := range history {
			changeset := item.GetChangeset()
			if _, ok := all_history[changeset]; !ok {
				all_history[changeset] = make([]*bridge.TfsHistoryItem, 0, len(tfs_subfolders))
			}
			all_history[changeset] = append(all_history[changeset], item)
		}
	}
	all_changesets := make([]int, 0, len(all_history))
	for changeset := range all_history {
		all_changesets = append(all_changesets, changeset)
	}
	sort.Ints(all_changesets)
	for _, changeset := range all_changesets {
		log.Println(changeset)
		for _, historyItem := range all_history[changeset] {
			repo := historyItem.GetRepo()
			repo.Update(changeset)
			tfsSubfolder := filepath.Join(tfs_root_folder, filepath.Base(repo.GetPath()))
			gitSubfolder := filepath.Join(git_root_folder, filepath.Base(repo.GetPath()))
			leftOnly, rightOnly, diffs := bridge.CompareDirectories(gitSubfolder, tfsSubfolder, gitIgnore, historyItem.IsAffected)
			if len(leftOnly) + len(rightOnly) + len(diffs) == 0 {
				continue
			}
			for path := range leftOnly {
				os.Remove(filepath.Join(gitSubfolder, path))
			}
			for path := range rightOnly {
				content, err := ioutil.ReadFile(filepath.Join(tfsSubfolder, path))
				if err != nil {
					log.Println(err)
					continue
				}
				err = ioutil.WriteFile(filepath.Join(gitSubfolder, path), content, 0666)
				if err != nil {
					log.Println(err)
					continue
				}
			}
			for path := range diffs {
				content, err := ioutil.ReadFile(filepath.Join(tfsSubfolder, path))
				if err != nil {
					log.Println(err)
					continue
				}
				err = ioutil.WriteFile(filepath.Join(gitSubfolder, path), content, 0666)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
		git.StageAll()
		historyItem := all_history[changeset][0]
		comment, author, date := historyItem.GetComment(), historyItem.GetAuthor(), historyItem.GetDate()
		author = strings.TrimPrefix(author, `NT_WORK\`)
		if comment != "" {
			comment += "\r\n\r\n"
		}
		comment += fmt.Sprintf("git-tfs-bridge: imported from TFS %d", historyItem.GetChangeset())
		git.Commit(comment, fmt.Sprintf("%s <%s@directum.ru>", author, author), date)
	}
}