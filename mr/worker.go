package mr

import (
	"fmt"
	"hash/fnv"
	"log"
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

	for {
		arg := GetTaskArg{}
		reply := new(GetTaskReply)
		ok := call("Coordinator.GetTask", arg, &reply)
		if !ok {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		switch reply.TaskType {
		case "map":
			contents, err := os.ReadFile(reply.FileName)
			if err != nil {
				log.Fatalf("Opening file: %v", err)
			}

			kva := mapf(reply.FileName, string(contents))

			buckets := make([][]KeyValue, reply.PartitionsAmount)
			for _, kv := range kva {
				i := ihash(kv.Key) % reply.PartitionsAmount
				buckets[i] = append(buckets[i], kv)
			}

			for i, bucket := range buckets {
				intmdFileName := fmt.Sprintf("mr-%d-%d", reply.PartitionsAmount, i)
				if err := MarshalKeyValues(intmdFileName, bucket); err != nil {
					log.Fatalf("Marshaling: %v", err)
				}
			}

			doneArg := DoneTaskArg{TaskNum: reply.TaskNum, TaskType: "map"}
			doneReply := new(DoneTaskReply)
			ok := call("Coordnator.TaskDone", doneArg, doneReply)
			if !ok {
				time.Sleep(500 * time.Millisecond)
				continue
			}

		case "reduce":
			buckets, err := UnmarshalKeyValues(reply.MapTaskNum, reply.Partition)
			if err != nil {
				log.Fatal(err)
			}

			sort.Slice(buckets, func(i, j int) bool {
				return buckets[i].Key < buckets[j].Key
			})

			outFileName := fmt.Sprintf("mr-out-%d", reply.Partition)
			ofile, err := os.Create(outFileName)
			if err != nil {
				log.Fatal(err)
			}

			defer func() {
				log.Fatal(ofile.Close())
			}()

			i := 0
			for i < len(buckets) {
				j := i + 1
				for j < len(buckets) && buckets[j].Key == buckets[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, buckets[k].Value)
				}
				output := reducef(buckets[i].Key, values)

				fmt.Fprintf(ofile, "%v %v\n", buckets[i].Key, output)

				i = j
			}

			doneArgs := DoneTaskArg{TaskType: "reduce", TaskNum: reply.Partition}
			doneReply := new(DoneTaskReply)
			ok := call("Coordinator.TaskDone", doneArgs, doneReply)
			if !ok {
				time.Sleep(500 * time.Millisecond)
				continue
			}
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
