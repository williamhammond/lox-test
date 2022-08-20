package main

import (
	"errors"
	"flag"
	"log"
	"os"
)

const cInterpreter = "C:\\Users\\willi\\Code\\clox\\cmake-build-debug\\clox.exe"
const javaInterpreter = "C:\\Program Files\\Common Files\\Oracle\\Java\\javapath\\java.exe"

const compileErrorCode = 65
const runtimeErrorCode = 70

var suites = map[string]Suite{}

var passed = 0
var failed = 0

func main() {
	initSuites()

	var path = flag.String("interpreter", "", "Path to lox interpreter")
	flag.Parse()
	if _, err := os.Stat(*path); *path != "" && errors.Is(err, os.ErrNotExist) {
		log.Panicln(err)
	}

	suiteNames := make([]string, len(suites))
	i := 0
	for k := range suites {
		suiteNames[i] = k
		i++
	}
	result := runSuites(suiteNames)
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
			if directory.Name() != "limit" {
				var javaSuite = Suite{
					name:       "java - " + directory.Name(),
					executable: javaInterpreter,
					tests:      tests,
					language:   "java",
				}
				suites[javaSuite.name] = javaSuite
			}
		}
	}
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
