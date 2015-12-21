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

func importFromTfs(gitRoot string, tfsRoot string, startChangeset int, needInit bool) {
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
	joinedHistory, changesets := getJoinedHistory(tfsRepos, gitIgnore, startChangeset, needInit)
	if needInit {
		if !initTfsChangeset(tfsRepos, gitRepo, changesets[0], joinedHistory, gitIgnore) {
			log.Println("Something is wrong!!! Git repo \"%s\" shold be clean after commit.")
			return
		}
		changesets = changesets[1:]
	}
	for _, changeset := range changesets {
		if !importTfsChangeset(gitRepo, changeset, joinedHistory, gitIgnore) {
			log.Println("Something is wrong!!! Git repo \"%s\" shold be clean after commit.")
			return
		}
	}
}

func importTfsChangeset(gitRepo *GitRepository, changeset int, joinedHistory map[int][]*TfsHistoryItem, gitIgnore *ignore.GitIgnore) bool {
	repos := make([]*TfsRepository, 0, 10)
	checkAffected := make(map[*TfsRepository]func(string) bool)
	for _, historyItem := range joinedHistory[changeset] {
		repo := historyItem.GetRepo()
		repos = append(repos, repo)
		checkAffected[repo] = historyItem.IsAffected
	}
	return importTfsChangesetFiles(repos, gitRepo, changeset, gitIgnore, checkAffected, joinedHistory[changeset][0])
}

func initTfsChangeset(tfsRepos []*TfsRepository, gitRepo *GitRepository, changeset int, joinedHistory map[int][]*TfsHistoryItem, gitIgnore *ignore.GitIgnore) bool {
	checkAffected := make(map[*TfsRepository]func(string) bool)
	return importTfsChangesetFiles(tfsRepos, gitRepo, changeset, gitIgnore, checkAffected, joinedHistory[changeset][0])
}

func importTfsChangesetFiles(tfsRepos []*TfsRepository, gitRepo *GitRepository, changeset int, gitIgnore *ignore.GitIgnore, checkAffected map[*TfsRepository]func(string) bool, sampleHistoryItem *TfsHistoryItem) bool {
	log.Println(changeset)
	for _, repo := range tfsRepos {
		repo.Update(changeset)
		tfsSubfolder := repo.GetPath()
		gitSubfolder := filepath.Join(gitRepo.path, filepath.Base(repo.GetPath()))
		leftOnly, rightOnly, diffs := CompareDirectories(gitSubfolder, tfsSubfolder, gitIgnore, checkAffected[repo])
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
	comment, author, date := sampleHistoryItem.GetComment(), sampleHistoryItem.GetAuthor(), sampleHistoryItem.GetDate()
	author = strings.TrimPrefix(author, `NT_WORK\`)
	if comment != "" {
		comment += "\r\n\r\n"
	}
	comment += fmt.Sprintf("git-tfs-bridge: imported from TFS %d", changeset)
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
		if !info.IsDir() || info.Name() == "AutoScripts" {
			continue
		}

		tfsRepo, err := OpenTfsRepository(filepath.Join(tfsRoot, info.Name()))
		if err == nil {
			tfsRepos = append(tfsRepos, tfsRepo)
		}
	}
	return tfsRepos, nil
}

func getJoinedHistory(tfsRepos []*TfsRepository, gitIgnore *ignore.GitIgnore, startChangeset int, includeStartChangeset bool) (joinedHistory map[int][]*TfsHistoryItem, orderedChangesets []int) {
	joinedHistory = make(map[int][]*TfsHistoryItem)
	for _, tfs := range tfsRepos {
		history := tfs.GetHistoryFrom(gitIgnore, startChangeset, includeStartChangeset)
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