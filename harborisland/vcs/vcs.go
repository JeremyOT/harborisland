package vcs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type MakeVCSFunction func(path, url, branch string) (VCS, error)

var supportedVCS = map[string]MakeVCSFunction{
	"git": NewGitVCS,
}

func runCommand(command string, args ...string) (err error) {
	output, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s %s failed %s:\n%s", command, args, err, string(output))
	}
	return
}

type VCS interface {
	Update() (bool, error)
	Type() string
}

func GetVCS(name, path, url, branch string) (vcs VCS, err error) {
	vcsFunc := supportedVCS[name]
	if vcsFunc == nil {
		return nil, fmt.Errorf("Unsupported VCS type: %s", name)
	}
	return vcsFunc(path, url, branch)
}

type gitVCS struct {
	path   string
	url    string
	branch string
}

func NewGitVCS(path, url, branch string) (VCS, error) {
	return &gitVCS{
		path:   path,
		url:    url,
		branch: branch,
	}, nil
}

func (v *gitVCS) clone() (err error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return
	}
	if err = os.MkdirAll(v.path, 0777); err != nil {
		return
	}
	if err = os.Chdir(v.path); err != nil {
		return
	}
	if err = runCommand(gitPath, "init"); err != nil {
		return
	}
	if err = runCommand(gitPath, "remote", "add", "-t", v.branch, "-f", "origin", v.url); err != nil {
		return
	}
	if err = runCommand(gitPath, "fetch", "--depth", "1"); err != nil {
		return
	}
	if err = runCommand(gitPath, "checkout", v.branch); err != nil {
		return
	}
	return
}

func (v *gitVCS) getCommit() (hash string, err error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return
	}
	output, err := exec.Command(gitPath, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s rev-parse HEAD failed %s:\n%s", gitPath, err, string(output))
	}
	hash = string(output)
	return
}

func (v *gitVCS) update() (changed bool, err error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return
	}
	if err = os.Chdir(v.path); err != nil {
		return
	}
	initialHash, err := v.getCommit()
	if err != nil {
		return
	}
	if err = runCommand(gitPath, "reset", "--hard", "HEAD"); err != nil {
		return
	}
	if err = runCommand(gitPath, "checkout", v.branch); err != nil {
		return
	}
	if err = runCommand(gitPath, "clean", "-f", "-d"); err != nil {
		return
	}
	if err = runCommand(gitPath, "pull", "origin", v.branch); err != nil {
		return
	}
	newHash, err := v.getCommit()
	if err != nil {
		return
	}
	changed = newHash != initialHash
	return
}

func (v *gitVCS) Type() string {
	return "git"
}

func (v *gitVCS) Update() (changed bool, err error) {
	dir, err := os.Stat(v.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
		err = v.clone()
		if err == nil {
			changed = true
		}
		return
	}
	if !dir.IsDir() {
		err = errors.New("Path is not a directory")
		return
	}
	changed, err = v.update()
	return
}
