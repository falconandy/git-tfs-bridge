package git_tfs_bridge

import (
	"os"
	"errors"
	"os/exec"
	"time"
)

type GitRepository struct {
	path string
}

func OpenGitRepository(path string) (*GitRepository, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		err = errors.New("Can't open git repository at `%s`. Not a directory.")
		return nil, err
	}
	cmd := exec.Command("git", "-C", path, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New(err.Error() + ". " + string(output))
	}
	return &GitRepository{path:path}, nil
}

func (repo *GitRepository) IsClean() bool {
	output, err := repo.execCommand("status", "--short")
	return len(output) == 0 && err == nil
}

func (repo *GitRepository) StageAll() {
	repo.execCommand("add", "-A")
}

func (repo *GitRepository) Commit(message string, author string, date time.Time) {
	repo.execCommand("commit", "-m", message, "--author="+author, "--date="+date.String())
}

func (repo *GitRepository) execCommand(args ...string) ([]byte, error) {
	args = append([]string{"-C", repo.path }, args...)
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}