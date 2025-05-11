package mr

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
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

	for {
		getTaskArg := GetTaskArg{}
		getTaskReply := new(GetTaskReply)
		ok := call("Coordinator.GetTask", getTaskArg, &getTaskReply)
		if !ok {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		switch getTaskReply.TaskType {
		case "map":
			contents, err := os.ReadFile(getTaskReply.FileName)
			if err != nil {
				log.Fatalf("Opening file: %v", err)
			}

			kva := mapf(getTaskReply.FileName, string(contents))

			buckets := make([][]KeyValue, getTaskReply.BucketsAmount)
			for _, kv := range kva {
				i := ihash(kv.Key) % getTaskReply.BucketsAmount
				buckets[i] = append(buckets[i], kv)
			}

			for i, bucket := range buckets {
				intmdFileName := fmt.Sprintf("mr-%d-%d", getTaskReply.BucketsAmount, i)
				_, err := MarshalKeyValues(intmdFileName, bucket)
				if err != nil {
					log.Fatalf("Marshaling: %v", err)
				}
			}

			doneArg := DoneTaskArg{TaskNum: getTaskReply.TaskNum, TaskType: "map"}
			doneResp := new(DoneTaskReply)
			ok := call("Coordnator.TaskDone", doneArg, doneResp)
			if !ok {
				time.Sleep(500 * time.Millisecond)
				continue
			}

		case "reduce":
			// TODO: reduce
		}
	}

	// CallExample()
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
