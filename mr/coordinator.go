package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type TaskRecord struct {
	Task   *Task
	Status TaskStatus
}

type Coordinator struct {
	// Your definitions here.
	mu          sync.Mutex
	JobDone     bool
	MapTasks    []*TaskRecord
	ReduceTasks []*TaskRecord
	mapDone     bool
	reduceDone  bool
}

func (c *Coordinator) CleanUp() {
	maps := len(c.MapTasks)
	reduces := len(c.ReduceTasks)

	for i := range maps {
		for j := range reduces {
			fname := fmt.Sprintf("mr-%d-%d", i, j)
			if err := os.Remove(fname); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (c *Coordinator) setTaskTimeout(task *TaskRecord, d time.Duration) {
	select {
	case <-time.After(d):
		c.mu.Lock()
		defer c.mu.Unlock()
		if task.Status != StatusCompleted {
			task.Status = StatusIdle
			return
		}
	}
}

func (c *Coordinator) GetTask(args *GetTaskArg, reply *GetTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.mapDone {
		for id, mt := range c.MapTasks {
			if mt.Status == StatusIdle {
				reply.Task = Task{
					TaskID:   id,
					TaskType: TaskTypeMap,
					MapTask: &MapTask{
						FileName:         mt.Task.MapTask.FileName,
						PartitionsAmount: mt.Task.MapTask.PartitionsAmount,
					},
				}
				mt.Status = StatusInProcess

				go c.setTaskTimeout(mt, 10*time.Second)
				return nil
			}
		}
		// All map tasks been asigned but still
		return nil
	}

	if !c.reduceDone {
		for id, rt := range c.ReduceTasks {
			if rt.Status == StatusIdle {
				reply.Task = Task{
					TaskID:   id,
					TaskType: TaskTypeReduce,
					ReduceTask: &ReduceTask{
						MapsAmount: len(c.MapTasks),
					},
				}
				rt.Status = StatusInProcess

				go c.setTaskTimeout(rt, 10*time.Second)
				return nil
			}
		}
		return nil
	}

	// reply.TaskType = TaskTypeExit
	return nil
}

func (c *Coordinator) TaskDone(args *DoneTaskArg, reply *DoneTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch args.TaskType {
	case TaskTypeMap:
		c.MapTasks[args.TaskID].Status = StatusCompleted

		c.mapDone = true
		for _, mt := range c.MapTasks {
			if mt.Status != StatusCompleted {
				c.mapDone = false
				break
			}
		}

	case TaskTypeReduce:
		c.ReduceTasks[args.TaskID].Status = StatusCompleted

		c.reduceDone = true
		for _, rt := range c.ReduceTasks {
			if rt.Status != StatusCompleted {
				c.reduceDone = false
				break
			}
		}
	}
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
	c.mu.Lock()
	defer c.mu.Unlock()

	resp := c.reduceDone
	return resp
}

// create a Coordinator.
// mrcoordinator/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := new(Coordinator)

	mapTasks := make([]*TaskRecord, len(files))
	for i, filename := range files {
		mapTasks[i] = &TaskRecord{
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
	c.MapTasks = mapTasks

	reduceTasks := make([]*TaskRecord, nReduce)
	for i := range nReduce {
		reduceTasks[i] = &TaskRecord{
			Task: &Task{
				ReduceTask: &ReduceTask{
					MapsAmount: len(files),
				},
				MapTask:  nil,
				TaskID:   i,
				TaskType: TaskTypeReduce,
			},
			Status: StatusIdle,
		}
	}
	c.ReduceTasks = reduceTasks

	c.server()
	fmt.Println("Coordinator started...")
	return c
}

// func initMapTasks() map[int]GetTaskReply {

// }
