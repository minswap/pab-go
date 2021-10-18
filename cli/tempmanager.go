package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type TempManager struct {
	openFiles []*os.File
}

func NewTempManager() (*TempManager, error) {
	tm := new(TempManager)

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

func (tm *TempManager) NewFile(suffix string) *os.File {
	file, err := ioutil.TempFile("", fmt.Sprintf("testnet-api-%s-", suffix))
	if err != nil {
		// We know it will likely succeed because we test ioutil.TempFile when initialize TempManager, but we still log for auditing exceptions.
		log.Printf("tempmanager: fail to create temp file: %v\n", err)
		return nil
	}
	tm.openFiles = append(tm.openFiles, file)
	return file
}

func (tm *TempManager) Clean() {
	for _, file := range tm.openFiles {
		file.Close()
		os.Remove(file.Name())
	}
}
