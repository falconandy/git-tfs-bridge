package git_tfs_bridge_test

import (
	"testing"
	bridge "github.com/falconandy/git-tfs-bridge"
	"log"
	"github.com/sabhiram/go-git-ignore"
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
	leftOnly, rightOnly, diffs = bridge.CompareDirectories(`D:\Projects\_sun\Main\Kernel`, `D:\Projects\Sungero\Main\Kernel`, gitIgnore)
	log.Println(leftOnly)
	log.Println(rightOnly)
	log.Println(diffs)
	leftOnly, rightOnly, diffs = bridge.CompareDirectories(`D:\Projects\_sun\Main\Content`, `D:\Projects\Sungero\Main\Content`, gitIgnore)
	log.Println(leftOnly)
	log.Println(rightOnly)
	log.Println(diffs)
	leftOnly, rightOnly, diffs = bridge.CompareDirectories(`D:\Projects\_sun\Main\Report`, `D:\Projects\Sungero\Main\Report`, gitIgnore)
	log.Println(leftOnly)
	log.Println(rightOnly)
	log.Println(diffs)
	leftOnly, rightOnly, diffs = bridge.CompareDirectories(`D:\Projects\_sun\Main\Workflow`, `D:\Projects\Sungero\Main\Workflow`, gitIgnore)
	log.Println(leftOnly)
	log.Println(rightOnly)
	log.Println(diffs)
	leftOnly, rightOnly, diffs = bridge.CompareDirectories(`D:\Projects\_sun\Main\SDS`, `D:\Projects\Sungero\Main\SDS`, gitIgnore)
	log.Println(leftOnly)
	log.Println(rightOnly)
	log.Println(diffs)
}