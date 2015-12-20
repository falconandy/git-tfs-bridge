package git_tfs_bridge

import (
	"path/filepath"
	"io/ioutil"
	"github.com/sabhiram/go-git-ignore"
	"log"
	"sort"
	"fmt"
	"strings"
	"os"
)

func ImportFromTfs(gitRoot string, tfsRoot string, fromChangeset int) {
	importFromTfs(gitRoot, tfsRoot, fromChangeset, false)
}

func InitFromTfs(gitRoot string, tfsRoot string, initChangeset int) {
	importFromTfs(gitRoot, tfsRoot, initChangeset, true)
}

func importFromTfs(gitRoot string, tfsRoot string, startChangeset int, includeStartChangeset bool) {
	tfsRepos, _ := GetTfsRepositories(tfsRoot)
	for _, tfsRepo := range tfsRepos {
		if !tfsRepo.IsClean() {
			log.Printf("Tfs repo \"%s\" is dirty. Clean it first.", tfsRepo.GetPath())
			return
		}
	}
	gitRepo, _ := OpenGitRepository(gitRoot)
	if !gitRepo.IsClean() {
		log.Printf("Git repo \"%s\" is dirty. Clean it first.", gitRepo.path)
		return
	}
	gitIgnore, _ := ignore.CompileIgnoreFile(filepath.Join(gitRoot, ".gitignore"))
	joinedHistory, changesets := getJoinedHistory(tfsRepos, startChangeset, includeStartChangeset)
	for i, changeset := range changesets {
		if !importTfsChangeset(tfsRoot, gitRepo, changeset, joinedHistory, gitIgnore, includeStartChangeset && i == 0) {
			log.Println("Something is wrong!!! Git repo \"%s\" shold be clean after commit.")
			return
		}
	}
}

func importTfsChangeset(tfsRoot string, gitRepo *GitRepository, changeset int, joinedHistory map[int][]*TfsHistoryItem, gitIgnore *ignore.GitIgnore, init bool) bool {
	log.Println(changeset)
	for _, historyItem := range joinedHistory[changeset] {
		repo := historyItem.GetRepo()
		repo.Update(changeset)
		tfsSubfolder := repo.GetPath()
		gitSubfolder := filepath.Join(gitRepo.path, filepath.Base(repo.GetPath()))
		var checkAffected func(string) bool
		if !init {
			checkAffected = historyItem.IsAffected
		}
		leftOnly, rightOnly, diffs := CompareDirectories(gitSubfolder, tfsSubfolder, gitIgnore, checkAffected)
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
			path := filepath.Join(gitSubfolder, path)
			os.MkdirAll(filepath.Dir(path), 0777)
			err = ioutil.WriteFile(path, content, 0666)
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
	gitRepo.StageAll()
	historyItem := joinedHistory[changeset][0]
	comment, author, date := historyItem.GetComment(), historyItem.GetAuthor(), historyItem.GetDate()
	author = strings.TrimPrefix(author, `NT_WORK\`)
	if comment != "" {
		comment += "\r\n\r\n"
	}
	comment += fmt.Sprintf("git-tfs-bridge: imported from TFS %d", historyItem.GetChangeset())
	gitRepo.Commit(comment, fmt.Sprintf("%s <%s@directum.ru>", author, author), date)

	return gitRepo.IsClean()
}

func GetTfsRepositories(tfsRoot string) ([]*TfsRepository, error) {
	items, err := ioutil.ReadDir(tfsRoot)
	if err != nil {
		return nil, err
	}
	tfsRepos := make([]*TfsRepository, 0, len(items))
	for _, info := range items {
		if !info.IsDir() {
			continue
		}

		tfsRepo, err := OpenTfsRepository(filepath.Join(tfsRoot, info.Name()))
		if err == nil {
			tfsRepos = append(tfsRepos, tfsRepo)
		}
	}
	return tfsRepos, nil
}

func getJoinedHistory(tfsRepos []*TfsRepository, startChangeset int, includeStartChangeset bool) (joinedHistory map[int][]*TfsHistoryItem, orderedChangesets []int) {
	joinedHistory = make(map[int][]*TfsHistoryItem)
	for _, tfs := range tfsRepos {
		history := tfs.GetHistoryFrom(startChangeset, includeStartChangeset)
		for _, item := range history {
			changeset := item.GetChangeset()
			if _, ok := joinedHistory[changeset]; !ok {
				joinedHistory[changeset] = make([]*TfsHistoryItem, 0, len(tfsRepos))
			}
			joinedHistory[changeset] = append(joinedHistory[changeset], item)
		}
	}
	orderedChangesets = make([]int, 0, len(joinedHistory))
	for changeset := range joinedHistory {
		orderedChangesets = append(orderedChangesets, changeset)
	}
	sort.Ints(orderedChangesets)
	return joinedHistory, orderedChangesets
}