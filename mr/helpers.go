package mr

import (
	"encoding/json"
	"fmt"
	"os"
)

func MarshalKeyValues(fname string, kva []KeyValue) (*os.File, error) {
	file, err := os.CreateTemp("./", fname+"-")
	if err != nil {
		return nil, fmt.Errorf("Creating file: %w", err)
	}
	defer func() {
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			os.Remove(file.Name())
		}
	}()

	enc := json.NewEncoder(file)
	if err = enc.Encode(&kva); err != nil {
		return nil, fmt.Errorf("Encoding: %w", err)
	}

	if err = os.Rename(file.Name(), fname); err != nil {
		return nil, fmt.Errorf("Renaming: %w", err)
	}

	if err = file.Sync(); err != nil {
		return nil, fmt.Errorf("Syncing: %w", err)
	}

	return file, nil
}

func UnmarshalKeyValues(taskNum int, bucketsAmount int) ([][]KeyValue, error) {
	buckets := make([][]KeyValue, bucketsAmount)

	for i := range buckets {
		fileName := fmt.Sprintf("mr-%d-%d", taskNum, i)
		f, err := os.Open(fileName)
		if err != nil {
			return nil, fmt.Errorf("Opening file: %w", err)
		}

		var kvl []KeyValue
		dec := json.NewDecoder(f)
		if err = dec.Decode(&kvl); err != nil {
			return nil, fmt.Errorf("Decoding: %w", err)
		}

		buckets[i] = kvl
	}

	return buckets, nil
}
