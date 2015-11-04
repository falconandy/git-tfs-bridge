package git_tfs_bridge_test

import (
	"testing"
	bridge "github.com/falconandy/git-tfs-bridge"
	"log"
)

func TestHistory(t *testing.T) {
	tfs := bridge.OpenTfsRepository(`D:\Projects\Sungero\Main\Common`)
	log.Println(tfs.GetHistory(20))
}