package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type tempManager struct {
	openFiles []*os.File
}

func newTempManager() (*tempManager, error) {
	tm := new(tempManager)

	// test if we can create temp file and write to it
	testFile, err := ioutil.TempFile("", "testnet-api-test-")
	if err != nil {
		return nil, fmt.Errorf("fail to create temp file: %w", err)
	}
	defer func() {
		testFile.Close()
		os.Remove(testFile.Name())
	}()
	if _, err := testFile.WriteString("test"); err != nil {
		return nil, fmt.Errorf("fail to write to temp file: %w", err)
	}

	return tm, nil
}

func (tm *tempManager) NewFile(suffix string) *os.File {
	file, err := ioutil.TempFile("", fmt.Sprintf("testnet-api-%s-", suffix))
	if err != nil {
		// We know it will likely succeed because we test ioutil.TempFile when initialize TempManager, but we still log for auditing exceptions.
		log.Printf("tempmanager: fail to create temp file: %v\n", err)
		return nil
	}
	tm.openFiles = append(tm.openFiles, file)
	return file
}

func (tm *tempManager) Clean() {
	for _, file := range tm.openFiles {
		file.Close()
		os.Remove(file.Name())
	}
}
