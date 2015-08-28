package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/JeremyOT/harborisland/harborisland/monitor"
	"github.com/JeremyOT/harborisland/harborisland/task"
	"github.com/JeremyOT/structflag"
)

type TaskDef struct {
	Tasks    []*task.Task `json:"tasks"`
	Registry string       `json:"registry"`
}

func LoadTasks(taskDefPath string) (tasks *TaskDef, err error) {
	f, err := os.Open(taskDefPath)
	if err != nil {
		return
	}
	defer f.Close()
	var taskDef TaskDef
	if err = json.NewDecoder(f).Decode(&taskDef); err != nil {
		return
	}
	tasks = &taskDef
	return
}

func main() {
	var config struct {
		pollInterval time.Duration
		taskDef      string
		workDir      string
	}
	argTask := &task.Task{}
	flag.DurationVar(&config.pollInterval, "poll-interval", time.Minute, "The frequency with which to poll registered tasks for updates.")
	flag.StringVar(&config.taskDef, "task-def", "", "The path to a JSON file containing task definitions.")
	flag.StringVar(&config.workDir, "work-dir", "", "The path to a directory to use as the build directory.")
	structflag.StructToFlags("task", argTask)
	flag.Parse()
	taskMon, err := monitor.New(config.pollInterval, config.workDir)
	if err != nil {
		log.Panicln(err)
	}
	if argTask.ImageName != "" {
		taskMon.AddTask(argTask)
	}
	if config.taskDef != "" {
		tasks, err := LoadTasks(config.taskDef)
		if err != nil {
			log.Panicln(err)
		}
		for _, task := range tasks.Tasks {
			if task.Registry == "" {
				task.Registry = tasks.Registry
			}
			taskMon.AddTask(task)
		}
	}
	taskMon.Start()
	taskMon.Wait()
}
