package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

type TaskStatus int

type TaskType int

const (
	StatusUnknown TaskStatus = iota
	StatusIdle
	StatusInProcess
	StatusCompleted
)

const (
	TaskTypeUnknown TaskType = iota
	TaskTypeMap
	TaskTypeReduce
	TaskTypeExit
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type Task struct {
	MapTask    *MapTask
	ReduceTask *ReduceTask
	TaskID     int
	TaskType   TaskType
}

type MapTask struct {
	FileName         string
	PartitionsAmount int
}

type ReduceTask struct {
	MapsAmount int
}

type GetTaskArg struct{}

type GetTaskReply struct {
	Task
}

type DoneTaskArg struct {
	TaskID   int
	TaskType TaskType
}
type DoneTaskReply struct{}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
