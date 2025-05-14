package mr

import (
	"fmt"
	"hash/fnv"
	"log"
	"math/rand/v2"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// mrworker/mrworker.go calls this function.
func Worker(
	mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	workerID := generateWorkerID()
	log.Printf("Worker %v starting...\n", workerID)
	for {
		arg, reply := &GetTaskArg{WorkerID: workerID}, new(GetTaskReply)
		call("Coordinator.GetTask", arg, reply)

		switch reply.TaskType {
		case TaskTypeMap:
			processMapTask(mapf, reply)
			responseTaskDone(workerID, TaskTypeMap, reply)
		case TaskTypeReduce:
			processReduceTask(reducef, reply)
			responseTaskDone(workerID, TaskTypeReduce, reply)
		case TaskTypeWait:
			time.Sleep(500 * time.Millisecond)
		case TaskTypeExit:
			responseTaskDone(workerID, TaskTypeExit, reply)
			log.Printf("Worker %v shutting down...\n", workerID)
			return
		}
	}
}

func generateWorkerID() uint64 {
	return (uint64(time.Now().UnixNano()) << 12) | uint64(rand.IntN(4096))
}

func processMapTask(mapf func(string, string) []KeyValue, reply *GetTaskReply) {
	contents, err := os.ReadFile(reply.MapTask.FileName)
	if err != nil {
		log.Printf("Opening file: %v", err)
		return
	}

	kva := mapf(reply.MapTask.FileName, string(contents))

	partitions := make([][]KeyValue, reply.MapTask.PartitionsAmount)
	for _, kv := range kva {
		i := ihash(kv.Key) % reply.MapTask.PartitionsAmount
		partitions[i] = append(partitions[i], kv)
	}

	for i, partition := range partitions {
		intmdFileName := fmt.Sprintf("mr-%d-%d", reply.TaskID, i)
		if err := MarshalKeyValues(intmdFileName, partition); err != nil {
			log.Printf("Marshaling: %v", err)
			return
		}
	}
}

func processReduceTask(reducef func(string, []string) string, reply *GetTaskReply) {
	partitions, err := UnmarshalKeyValues(reply.TaskID, reply.ReduceTask.MapsAmount)
	if err != nil {
		log.Print(err)
		return
	}

	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Key < partitions[j].Key
	})

	outFileName := fmt.Sprintf("mr-out-%d", reply.TaskID)
	ofile, err := os.Create(outFileName)
	if err != nil {
		log.Print(err)
		return
	}

	defer func() {
		if closeErr := ofile.Close(); closeErr != nil {
			log.Print(closeErr)
		}
	}()

	i := 0
	for i < len(partitions) {
		j := i + 1
		for j < len(partitions) && partitions[j].Key == partitions[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, partitions[k].Value)
		}
		output := reducef(partitions[i].Key, values)

		fmt.Fprintf(ofile, "%v %v\n", partitions[i].Key, output)

		i = j
	}
}

func responseTaskDone(workerID uint64, tasktype TaskType, reply *GetTaskReply) {
	doneReply, doneArg := new(DoneTaskReply), DoneTaskArg{
		TaskID:   reply.TaskID,
		WorkerID: workerID,
		TaskType: tasktype,
	}
	call("Coordinator.TaskDone", &doneArg, doneReply)
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args any, reply any) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
