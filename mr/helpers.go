package mr

import (
	"encoding/json"
	"fmt"
	"os"
)

func MarshalKeyValues(fname string, kva []KeyValue) error {
	file, err := os.CreateTemp("./", fname+"-")
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

	return nil
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
