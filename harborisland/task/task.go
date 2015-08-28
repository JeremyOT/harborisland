package task

import (
	"fmt"
	"log"
	"os/exec"
	"path"
)

const (
	DefaultVCS        = "git"
	DefaultBranch     = "master"
	DefaultTag        = "latest"
	DefaultDockerfile = "Dockerfile"
	DefaultDirectory  = "."
)

type Task struct {
	ImageName        string   `json:"image_name"`
	RepositoryURL    string   `json:"repository"`
	VCSType          string   `json:"vcs"`
	Tag              string   `json:"tag"`
	Branch           string   `json:"branch"`
	Registry         string   `json:"registry"`
	BuildCommand     []string `json:"build" flag:"-"`
	PostBuildCommand []string `json:"postbuild" flag:"-"`
	Dockerfile       string   `json:"dockerfile"`
	Directory        string   `json:"directory"`
}

func (t *Task) GetVCSType() string {
	if t.VCSType == "" {
		return DefaultVCS
	}
	return t.VCSType
}

func (t *Task) GetBranch() string {
	if t.Branch == "" {
		return DefaultBranch
	}
	return t.Branch
}

func (t *Task) GetTag() string {
	if t.Tag == "" {
		return DefaultTag
	}
	return t.Tag
}

func (t *Task) GetName() string {
	name := fmt.Sprintf("%s:%s", t.ImageName, t.GetTag())
	if t.Registry != "" {
		name = path.Join(t.Registry, name)
	}
	return name
}

func (t *Task) GetBuildCommand() []string {
	if t.BuildCommand != nil {
		return t.BuildCommand
	}
	return []string{"docker", "build", "-t", t.GetName(), "-f", t.GetDockerfile(), t.GetDirectory()}
}

func (t *Task) runCommand(command []string) (output []byte, err error) {
	executable, err := exec.LookPath(command[0])
	if err != nil {
		return
	}
	output, err = exec.Command(executable, command[1:len(command)]...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s:\n%s", err, string(output))
	}
	return
}

func (t *Task) Build() (err error) {
	buildCommand := t.GetBuildCommand()
	log.Printf("Building %v with %s\n", t, buildCommand)
	if len(buildCommand) < 1 {
		err = fmt.Errorf("Invalid build command: %s", buildCommand)
		return
	}
	_, err = t.runCommand(buildCommand)
	if err != nil {
		log.Println("Built", t)
	}
	return
}

func (t *Task) PostBuild() (err error) {
	if t.PostBuildCommand != nil && len(t.PostBuildCommand) > 0 {
		_, err := t.runCommand(t.PostBuildCommand)
		if err != nil {
			log.Println("Failed to run command:", t.PostBuildCommand)
			return err
		}
	}
	if t.Registry == "" {
		return
	}
	pushCommand := []string{"docker", "push", t.GetName()}
	_, err = t.runCommand(pushCommand)
	if err != nil {
		log.Printf("Pushed container for task %v with %s\n", t, pushCommand)
	}
	return
}

func (t *Task) GetDockerfile() string {
	if t.Dockerfile == "" {
		return path.Join(t.GetDirectory(), DefaultDockerfile)
	}
	return path.Join(t.GetDirectory(), t.Dockerfile)
}

func (t *Task) GetDirectory() string {
	if t.Directory == "" {
		return DefaultDirectory
	}
	return t.Directory
}

func (t *Task) String() string {
	return t.GetName()
}
