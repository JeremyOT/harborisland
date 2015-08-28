package monitor

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/JeremyOT/harborisland/harborisland/task"
	"github.com/JeremyOT/harborisland/harborisland/vcs"
)

type Monitor struct {
	pollInterval time.Duration
	workDir      string
	tasks        []*task.Task
	wait         chan struct{}
	quit         chan struct{}
}

func New(pollInterval time.Duration, workDir string) (*Monitor, error) {
	return &Monitor{
		pollInterval: pollInterval,
		workDir:      workDir,
	}, nil
}

func (m *Monitor) AddTask(t *task.Task) {
	m.tasks = append(m.tasks, t)
}

func (m *Monitor) PerformTask(t *task.Task) {
	if err := os.MkdirAll(m.workDir, 0777); err != nil {
		log.Println("Failed to create working directory:", err)
		return
	}
	if err := os.Chdir(m.workDir); err != nil {
		log.Println("Failed to update working directory:", err)
		return
	}
	taskPath := path.Join(m.workDir, t.ImageName, t.GetTag())
	if taskVCS, err := vcs.GetVCS(t.GetVCSType(), taskPath, t.RepositoryURL, t.GetBranch()); err != nil {
		log.Printf("Failed to update task %#v due to error: %s\n", t, err)
		return
	} else {
		if changed, err := taskVCS.Update(); err != nil {
			log.Printf("Failed to update task %#v due to error: %s\n", t, err)
			return
		} else if !changed {
			log.Printf("No change for task %v\n", t)
			return
		}
	}
	if err := os.Chdir(taskPath); err != nil {
		log.Println("Failed to enter task directory:", err)
		return
	}
	log.Printf("Change detected for task %v\n", t)
	if err := t.Build(); err != nil {
		log.Printf("Build failed for task %v: %s\n", t, err)
		return
	}
	if err := t.PostBuild(); err != nil {
		log.Printf("Post build failed for task %v: %s\n", t, err)
	}
	return
}

func (m *Monitor) PerformTasks() {
	for _, task := range m.tasks {
		m.PerformTask(task)
	}
}

func (m *Monitor) run() {
	m.PerformTasks()
	ticker := time.Tick(m.pollInterval)
	for {
		select {
		case <-ticker:
			m.PerformTasks()
		case <-m.quit:
			close(m.wait)
			return
		}
	}
}

func (m *Monitor) Start() {
	m.quit = make(chan struct{})
	m.wait = make(chan struct{})
	go m.run()
}

func (m *Monitor) Stop() {
	close(m.quit)
	<-m.wait
}

func (m *Monitor) Wait() {
	<-m.wait
}
