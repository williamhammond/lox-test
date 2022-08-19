package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
)

const cInterpreter = "C:\\Users\\willi\\Code\\clox\\cmake-build-debug\\clox.exe"
const javaInterpreter = "C:\\Program Files\\Common Files\\Oracle\\Java\\javapath\\java.exe"

const compileErrorCode = 65
const runtimeErrorCode = 70

var suites = map[string]Suite{}

// var suite Suite
var passed = 0
var failed = 0

func main() {
	initSuites()

	var path = flag.String("interpreter", "", "Path to lox interpreter")
	flag.Parse()
	if _, err := os.Stat(*path); *path != "" && errors.Is(err, os.ErrNotExist) {
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
	directories, err := os.ReadDir("tests/")
	if err != nil {
		panic(err)
	}
	for _, directory := range directories {
		var tests []string
		if directory.IsDir() {
			testFiles, err := os.ReadDir("tests/" + directory.Name())
			if err != nil {
				panic(err)
			}
			for _, testFile := range testFiles {
				tests = append(tests, "tests/"+directory.Name()+"/"+testFile.Name())
			}
			var javaSuite = Suite{
				name:       "java - " + directory.Name(),
				executable: javaInterpreter,
				tests:      tests,
				language:   "java",
			}
			suites[javaSuite.name] = javaSuite
		}
	}

	//var cSuite = Suite{
	//	executable: cInterpreter,
	//	tests:      tests,
	//}
	//suites["c"] = cSuite
}

func runSuites(names []string) bool {
	successful := true
	for _, name := range names {
		suite := suites[name]
		if !suite.run() {
			successful = false
		}
	}
	return successful
}

type Suite struct {
	name       string
	executable string
	language   string
	args       []string
	tests      []string
}

func (s *Suite) run() bool {
	log.Printf("====== Suite: %s ======", s.name)

	passed = 0
	failed = 0

	for _, test := range s.tests {
		s.runTest(test)
	}
	isSuccessful := failed == 0
	if isSuccessful {
		log.Printf("All "+Green+"%d"+Reset+" tests passed!", passed)
	} else {

		log.Printf(Green+"%d"+Reset+" tests passed. "+Red+"%d"+Reset+" tests failed", passed, failed)
	}
	return isSuccessful
}

func (s *Suite) runTest(path string) {
	if strings.Contains(path, "benchmark") {
		return
	}

	test := Test{
		path: path,
	}

	log.Printf("Running test: %s", path)

	err := test.parse(s.language)
	if err != nil {
		log.Panicln(err)
	}

	lox := exec.Command(s.executable, "-jar", "C:\\Users\\willi\\Code\\lox\\out\\artifacts\\lox_jar\\lox.jar", test.path)
	failures := test.run(lox)
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
