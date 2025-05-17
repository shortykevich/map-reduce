package mr_test

import (
	"os"
	"slices"
	"testing"

	"github.com/shrtyk/map-reduce/mr"
)

func TestHelpers(t *testing.T) {
	testFileName := "mr-0-0"
	testData := []mr.KeyValue{
		{Key: "string", Value: "1"},
		{Key: "string1", Value: "1"},
		{Key: "string2", Value: "1"},
		{Key: "string3", Value: "1"},
		{Key: "string4", Value: "1"},
		{Key: "string5", Value: "1"},
		{Key: "string6", Value: "1"},
		{Key: "string7", Value: "1"},
	}

	t.Cleanup(func() {
		if err := os.Remove(testFileName); err != nil {
			t.Fatalf("Deleteing: %v", err)
		}
	})

	if err := mr.MarshalKeyValues(testFileName, testData); err != nil {
		t.Fatalf("Encoding: %v", err)
	}

	decodedData, err := mr.UnmarshalKeyValues(0, 1)
	if err != nil {
		t.Fatalf("Decoding: %v", err)
	}

	if !slices.Equal(testData, decodedData) {
		t.Errorf("got: %v, wanted: %v", decodedData, testData)
	}
}
