package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const cInterpreter = "C:\\Users\\willi\\Code\\clox\\cmake-build-debug\\clox.exe"

var suites = map[string]Suite{}

func main() {
	initSuites()

	var path = flag.String("interpreter", "", "Path to lox interpreter")
	flag.Parse()
	if _, err := os.Stat(*path); errors.Is(err, os.ErrNotExist) {
		log.Panicln(err)
	}

	keys := make([]string, len(suites))

	i := 0
	for k := range suites {
		keys[i] = k
		i++
	}
	runSuites(keys)

	//log.Printf("running: %s", *path)
	//lox := exec.Command(*path)
	//bytes, err := lox.Output()
	//if err != nil {
	//	log.Panicln(err)
	//}
	//result := string(bytes)
	//log.Println(result)
}

func initSuites() {
	var tests []string
	err := filepath.Walk("tests/bool", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			tests = append(tests, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	var cSuite = Suite{
		executable: cInterpreter,
		tests:      tests,
	}
	suites["c"] = cSuite
}

func runSuites(names []string) {
	for _, name := range names {
		log.Printf("======= Suite: %s ======", name)
		runSuite(suites[name])
	}
}

func runSuite(suite Suite) {
	for _, test := range suite.tests {
		runTest(test)
	}
}

func runTest(path string) {
	if strings.Contains(path, "benchmark") {
		return
	}

	log.Printf("Running test: %s", path)
}

type Suite struct {
	executable string
	args       []string
	tests      []string
}

type Test struct {
	path           string
	expectedOutput []ExpectedOutput
	expectedErrors []error
	failures       []string
}

type ExpectedOutput struct {
}
