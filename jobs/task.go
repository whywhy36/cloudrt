package jobs

import (
	"encoding/json"
	"time"
)

// TaskState is state of task
type TaskState int

// Task states
const (
	TaskCreated   TaskState = iota // task is created, not ready for exec
	TaskPending                    // task is ready for execution
	TaskRunning                    // task is running
	TaskWaiting                    // task is waiting for sub-tasks
	TaskReverting                  // canceling/rolling back
	TaskStucked                    // error state, unable to retry or rollback
	TaskSuccess                    // task completed successfully
	TaskFailure                    // task completed but failed
	TaskAborted                    // task is canceled/rolled back
)

// TaskError is the type for error when task failed
type TaskError struct {
	Type       string    `json:"type"`        // error type
	Message    string    `json:"message"`     // error Message
	Output     []byte    `json:"output"`      // arbitrary output
	HappenedAt time.Time `json:"happened-at"` // time when task failed
}

// Task defines the details of a task`
type Task struct {
	ID        string        `json:"id"`         // globally unique task id
	ParentID  string        `json:"parent-id"`  // parent task id
	JobID     string        `json:"job-id"`     // job id
	Name      string        `json:"name"`       // task name
	Params    []byte        `json:"params"`     // encoded parameters
	Timeout   time.Duration `json:"timeout"`    // heartbeat Timeout
	State     TaskState     `json:"state"`      // current state
	Retries   uint          `json:"retries"`    // current retry number
	Stage     string        `json:"stage"`      // stage resume to
	Data      []byte        `json:"data"`       // task specific data
	Output    []byte        `json:"output"`     // output when completed
	Errors    []TaskError   `json:"errors"`     // errors happened
	CreatedAt time.Time     `json:"created-at"` // task creation time
	UpdatedAt time.Time     `json:"updated-at"` // last modification time
}

// GetParams extracts the parameters
func (t *Task) GetParams(p interface{}) error {
	params := t.Params
	if params == nil {
		return nil
	}
	return json.Unmarshal(params, p)
}

// GetData retieves and decodes the data
func (t *Task) GetData(d interface{}) error {
	data := t.Data
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, d)
}

// SetData encodes and saves the data
func (t *Task) SetData(d interface{}) *Task {
	encoded, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	t.Data = encoded
	return t
}

// GetOutput decodes the output
func (t *Task) GetOutput(p interface{}) error {
	output := t.Output
	if output == nil {
		return nil
	}
	return json.Unmarshal(output, p)
}

// SetOutput encodes and saves the output
func (t *Task) SetOutput(p interface{}) *Task {
	encoded, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	t.Output = encoded
	return t
}

// TaskSubmitter defines the contract which submits a task
type TaskSubmitter interface {
	SubmitTask(*Task) error
}

// TaskBuilder is a helper to build a task
type TaskBuilder struct {
	Submitter TaskSubmitter
	ID        string
	Name      string
	Params    interface{}
}

// NewTask starts defining a task
func NewTask(name string) *TaskBuilder {
	return &TaskBuilder{Name: name}
}

// SetID specifies the globally unqiue ID of task
func (b *TaskBuilder) SetID(id string) *TaskBuilder {
	b.ID = id
	return b
}

// With specifies the parameters which will be encoded later
func (b *TaskBuilder) With(params interface{}) *TaskBuilder {
	b.Params = params
	return b
}

// Build builds the task
func (b *TaskBuilder) Build() *Task {
	task := &Task{ID: b.ID}
	if task.ID == "" {
		// TODO generate a unique ID
	}
	if b.Params != nil {
		encoded, err := json.Marshal(b.Params)
		if err != nil {
			panic(err)
		}
		task.Params = encoded
	}
	return task
}

// Submit submits the task for execution
func (b *TaskBuilder) Submit() (*Task, error) {
	task := b.Build()
	return task, b.Submitter.SubmitTask(task)
}

// TaskFn is the function to execute the task
type TaskFn func(Context) error

// Stage defines a named stage with specified task function
type Stage struct {
	Name string // name of the stage
	Fn   TaskFn // task function
}

// TaskExec is the implemetation of the task
type TaskExec struct {
	Name   string  // name of the task
	Stages []Stage // stages in the task
}
