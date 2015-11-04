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
)

type TfsRepository struct {
	path string
}

func OpenTfsRepository(path string) *TfsRepository {
	return &TfsRepository{path:path}
}

func (repo *TfsRepository) GetHistory(count int) []*TfsHistoryItem {
	var history []*TfsHistoryItem
	output, err := repo.execCommand("history", repo.path, "/recursive", "/noprompt", "/format:Detailed", fmt.Sprintf("/stopafter:%d", count))
	if err != nil {
		log.Println(err)
	} else {
		history = parseHistory(ansi2utf8(output), count)
	}
	return history
}

func (repo *TfsRepository) execCommand(args ...string) ([]byte, error) {
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
