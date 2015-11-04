package git_tfs_bridge_test

import (
	"testing"
	bridge "github.com/falconandy/git-tfs-bridge"
	"log"
)

func TestHistory(t *testing.T) {
	tfs := bridge.OpenTfsRepository(`D:\Projects\Sungero\Main\Common`)
	log.Println(tfs.GetHistoryAfter(95000))
	log.Println(tfs.GetHistory(50000, 100))
}