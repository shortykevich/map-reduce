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
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			os.Remove(file.Name())
		}
	}()

	enc := json.NewEncoder(file)
	for _, kv := range kva {
		err := enc.Encode(&kv)
		if err != nil {
			return fmt.Errorf("Encoding: %w", err)
		}
	}

	if err = os.Rename(file.Name(), fname); err != nil {
		return fmt.Errorf("Renaming: %w", err)
	}

	if err = file.Sync(); err != nil {
		return fmt.Errorf("Syncing: %w", err)
	}

	return nil
}
