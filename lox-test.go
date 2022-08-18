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
const Red = "\033[31m"
const Reset = "\033[0m"
const Green = "\033[32m"

var suites = map[string]Suite{}

var passed = 0
var failed = 0

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
	result := runSuites(keys)
	if !result {
		os.Exit(1)
	}
	os.Exit(0)

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

func runSuites(names []string) bool {
	successful := true
	for _, name := range names {
		log.Printf("====== Suite: %s ======", name)
		if !runSuite(suites[name]) {
			successful = false
		}
	}
	return successful
}

func runSuite(suite Suite) bool {
	passed = 0
	failed = 0
	for _, test := range suite.tests {
		runTest(test)
	}
	isSuccessful := failed == 0
	if isSuccessful {
		log.Printf("All "+Green+"%d"+Reset+" tests passed!", passed)
	} else {

		log.Printf(Green+"%d"+Reset+" tests passed. "+Red+"%d"+Reset+" tests failed", passed, failed)
	}
	return isSuccessful
}

func runTest(path string) {
	if strings.Contains(path, "benchmark") {
		return
	}

	test := Test{
		path: path,
	}

	log.Printf("Running test: %s", path)
	failures := test.run()
	if len(failures) == 0 {
		passed++
	} else {
		failed++
		log.Printf(Red+"FAIL: %s"+Reset, path)
		for _, failure := range failures {
			log.Println(Red + failure + Reset)
		}
		log.Println()
	}
}

type Suite struct {
	executable string
	args       []string
	tests      []string
}

type ExpectedOutput struct {
}

type Test struct {
	path           string
	expectedOutput []ExpectedOutput
	expectedErrors []error
	failures       []string
}

func (t Test) run() []string {
	return []string{"test failed!"}
}
