package main

import (
	"encoding/json"
	"os"
	"sync"
)

// -------- Data Models --------

type Function struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Language string `json:"language"`
	Code     []byte `json:"code"`
}

type Execution struct {
	ID          string `json:"id"`
	FunctionID  string `json:"function_id"`
	Status      string `json:"status"`
	Output      []byte `json:"output"`
	Error       string `json:"error"`
}

// -------- Store --------

type Store struct {
	mu             sync.Mutex
	functionsFile  string
	executionsFile string
}

func NewStore() *Store {
	return &Store{
		functionsFile:  "data/functions.json",
		executionsFile: "data/executions.json",
	}
}

// -------- Helpers --------

func readJSON(path string, target interface{}) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}

func writeJSON(path string, data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// -------- Store APIs --------

func (s *Store) SaveFunction(fn Function) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	funcs := map[string]Function{}
	_ = readJSON(s.functionsFile, &funcs)

	funcs[fn.ID] = fn
	return writeJSON(s.functionsFile, funcs)
}

func (s *Store) SaveExecution(exec Execution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	execs := map[string]Execution{}
	_ = readJSON(s.executionsFile, &execs)

	execs[exec.ID] = exec
	return writeJSON(s.executionsFile, execs)
}

func (s *Store) GetExecution(id string) (Execution, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	execs := map[string]Execution{}
	_ = readJSON(s.executionsFile, &execs)

	exec, ok := execs[id]
	return exec, ok
}
