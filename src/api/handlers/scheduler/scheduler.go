package schedulerHandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"orla-alert/solte.lab/src/api"
	"orla-alert/solte.lab/src/scheduler"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

type SchedulerHandler struct {
	Scheduler *scheduler.Scheduler
}

type RequestParams struct {
	TaskName      string `json:"task_name"`
	TaskStartTime string `json:"task_start_time"`
	TaskInterval  string `json:"task_interval"`
}

func (h *SchedulerHandler) Register(s *api.Server) (string, chi.Router) {
	routes := chi.NewRouter()
	routes.Use(s.SetRequestID)
	routes.Use(s.LogRequest)

	routes.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		tasksInfo := h.Scheduler.TasksListInfo()
		if tasksInfo == nil {
			s.Error(w, r, http.StatusBadRequest, fmt.Errorf("can't find any tasks"))
			return
		}

		s.Respond(w, r, http.StatusOK, tasksInfo)
		return
	})

	routes.Get("/tasks/{task}/status", func(w http.ResponseWriter, r *http.Request) {
		task := chi.URLParam(r, "task")
		taskInfo, err := h.Scheduler.TaskInfo(task)
		if err != nil {
			s.Error(w, r, http.StatusNotFound, fmt.Errorf("can't retrive task. error: %v", err))
			return
		}

		_ = taskInfo.Status.String()
		w.Header().Add("Content-Type", "application/json")
		s.Respond(w, r, http.StatusOK, taskInfo)
		return
	})

	routes.Post("/tasks/{task}/run-once", func(w http.ResponseWriter, r *http.Request) {
		task := chi.URLParam(r, "task")
		err := h.Scheduler.RunOnNextTick(task)
		if err != nil {
			s.Error(w, r, http.StatusBadRequest, nil)
		}
		s.Respond(w, r, http.StatusOK, fmt.Sprintf("The task will be attempted to run on the next scheduled cycle"))
	})

	routes.Post("/tasks/{task}/schedule-task", func(w http.ResponseWriter, r *http.Request) {
		var task RequestParams

		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			s.Error(w, r, http.StatusBadRequest, errors.New(fmt.Sprintf("incorrect or unsupported request parameters err: %v", err)))
			return
		}

		interval, err := time.ParseDuration(task.TaskInterval)
		if err != nil {
			s.Error(w, r, http.StatusBadRequest, errors.New(fmt.Sprintf("incorrect or unsupported request parameters err: %v", err)))
			return
		}

		startTime, err := h.Scheduler.ParseStartTime(task.TaskStartTime)
		if err != nil {
			s.Error(w, r, http.StatusBadRequest, errors.New(fmt.Sprintf("incorrect or unsupported request parameters err: %v", err)))
			return
		}

		taskName := chi.URLParam(r, "task")

		err = h.Scheduler.ScheduleTask(taskName, startTime, interval)
		if err != nil {
			s.Error(w, r, http.StatusBadRequest, errors.New(fmt.Sprintf("%v", err)))
			return
		}

		s.Respond(w, r, http.StatusOK, fmt.Sprintf("task %s scheduled to run at %s, with interval %s", taskName, interval, startTime))
	})

	routes.Post("/tasks/{task}/cancel", func(w http.ResponseWriter, r *http.Request) {
		task := chi.URLParam(r, "task")
		err := h.Scheduler.Cancel(task)
		if err != nil {
			s.Error(w, r, http.StatusBadRequest, nil)
		}
		s.Respond(w, r, http.StatusOK, fmt.Sprintf("Task execution canceled"))
	})

	return "/scheduler", routes
}
