package mr

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
	"sort"
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
		arg := new(GetTaskArg)
		reply := new(GetTaskReply)
		if ok := call("Coordinator.GetTask", arg, reply); !ok {
			fmt.Println("Worker shutting down...")
			return
		}

		switch reply.TaskType {
		case TaskTypeMap:
			contents, err := os.ReadFile(reply.MapTask.FileName)
			if err != nil {
				log.Fatalf("Opening file: %v", err)
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
					log.Fatalf("Marshaling: %v", err)
				}
			}

			doneArg := DoneTaskArg{TaskID: reply.TaskID, TaskType: TaskTypeMap}
			doneReply := new(DoneTaskReply)
			call("Coordinator.TaskDone", &doneArg, doneReply)

		case TaskTypeReduce:
			partitions, err := UnmarshalKeyValues(reply.TaskID, reply.ReduceTask.MapsAmount)
			if err != nil {
				log.Fatal(err)
			}

			sort.Slice(partitions, func(i, j int) bool {
				return partitions[i].Key < partitions[j].Key
			})

			outFileName := fmt.Sprintf("mr-out-%d", reply.TaskID)
			ofile, err := os.Create(outFileName)
			if err != nil {
				log.Fatal(err)
			}

			defer func() {
				log.Fatal(ofile.Close())
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

			doneArgs := DoneTaskArg{
				TaskType: TaskTypeReduce,
				TaskID:   reply.TaskID,
			}
			doneReply := new(DoneTaskReply)
			call("Coordinator.TaskDone", &doneArgs, doneReply)

		case TaskTypeExit:
			fmt.Println("Worker shutting down...")
			return
		}
	}
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
