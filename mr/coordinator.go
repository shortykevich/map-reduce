package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

type TaskRecord struct {
	Task   *Task
	Status TaskStatus
}

type Coordinator struct {
	// Your definitions here.
	mu    sync.RWMutex
	done  bool
	tasks map[int]TaskRecord
}

func (c *Coordinator) GetTask(args *GetTaskArg, reply *GetTaskReply) error {
	return nil
}

func (c *Coordinator) TaskDone(args *GetTaskArg, reply *GetTaskReply) error {
	return nil
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// mrcoordinator/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false

	// TODO

	return ret
}

// create a Coordinator.
// mrcoordinator/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := new(Coordinator)

	// TODO
	tasks := make(map[int]TaskRecord, len(files))
	for i, filename := range files {
		tasks[i] = TaskRecord{
			Task: &Task{
				MapTask: &MapTask{
					FileName:         filename,
					PartitionsAmount: nReduce,
				},
				ReduceTask: nil,
				TaskID:     i,
				TaskType:   TaskTypeMap,
			},
			Status: StatusIdle,
		}
	}
	c.tasks = tasks

	c.server()
	return c
}

// func initMapTasks() map[int]GetTaskReply {

// }
