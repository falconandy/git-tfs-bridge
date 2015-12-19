package git_tfs_bridge_test

import (
	"testing"
	bridge "github.com/falconandy/git-tfs-bridge"
	"log"
	"github.com/sabhiram/go-git-ignore"
)

func TestHistory(t *testing.T) {
	tfs, _ := bridge.OpenTfsRepository(`D:\Projects\Sungero\Icons\Kernel`)
	log.Println(tfs)
	//log.Println(tfs.GetHistoryAfter(100914, false))
	log.Println(tfs.GetHistoryFrom(100914, true))
	//log.Println(tfs.GetHistory(50000, 100))
}

func TestUpdate(t *testing.T) {
	tfs, _ := bridge.OpenTfsRepository(`D:\Projects\Sungero\Main\Common`)
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

func TestGetTfsRepositories(t *testing.T) {
	tfsRepos, _ := bridge.GetTfsRepositories(`D:\Projects\Sungero\Main`)
	for _, repo := range tfsRepos {
		log.Println(repo.GetPath())
	}
}

func TestToGit(t *testing.T) {
	bridge.ImportFromTfs(`D:\Projects\_sun\Icons`, `D:\Projects\Sungero\Icons\`, 100914)
}