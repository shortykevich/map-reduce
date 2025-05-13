package mr

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

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

	doneArg := DoneTaskArg{TaskID: reply.TaskID, TaskType: TaskTypeMap}
	doneReply := new(DoneTaskReply)
	call("Coordinator.TaskDone", &doneArg, doneReply)
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
		log.Print(ofile.Close())
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
}

func MarshalKeyValues(fname string, kva []KeyValue) error {
	file, err := os.CreateTemp(".", fname+"-")
	if err != nil {
		return fmt.Errorf("Creating file: %w", err)
	}

	defer func() {
		if err != nil {
			if deleteErr := os.Remove(file.Name()); deleteErr != nil {
				err = fmt.Errorf("Deleting file: %w", deleteErr)
			}
		}
	}()

	enc := json.NewEncoder(file)
	if err = enc.Encode(&kva); err != nil {
		return fmt.Errorf("Encoding: %w", err)
	}

	if err = file.Sync(); err != nil {
		return fmt.Errorf("Syncing: %w", err)
	}

	if err = file.Close(); err != nil {
		return fmt.Errorf("Closing: %w", err)
	}

	if err = os.Rename(file.Name(), fname); err != nil {
		return fmt.Errorf("Renaming: %w", err)
	}

	return err
}

func UnmarshalKeyValues(partition int, mapTasksAmount int) ([]KeyValue, error) {
	buckets := []KeyValue{}

	for i := range mapTasksAmount {
		fileName := fmt.Sprintf("mr-%d-%d", i, partition)
		f, err := os.Open(fileName)
		if err != nil {
			return nil, fmt.Errorf("Opening file: %w", err)
		}

		var kvl []KeyValue
		dec := json.NewDecoder(f)
		if err = dec.Decode(&kvl); err != nil {
			return nil, fmt.Errorf("Decoding: %w", err)
		}

		buckets = append(buckets, kvl...)
	}

	return buckets, nil
}
