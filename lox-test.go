package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const cInterpreter = "C:\\Users\\willi\\Code\\clox\\cmake-build-debug\\clox.exe"

// const javaInterpreter = "java -jar C:\\Users\\willi\\Code\\lox\\out\\artifacts\\lox_jar\\lox.jar"
const javaInterpreter = "C:\\Program Files\\Common Files\\Oracle\\Java\\javapath\\java.exe"

//expectedOutputPattern, err := regexp.Compile("// expect: ?(.*)")
//expectedErrorPattern, err  := regexp.Compile("// (Error.*)")
//errorLinePattern, err := regexp.Compile("// \[((java|c) )?line (\d+)\] (Error.*)")
//expectedRuntimeErrorPattern, err := regexp.Compile("// expect runtime error: (.+)")
//syntaxErrorPattern, err := regexp.Compile("\[.*line (\d+)\] (Error.+)")
//stackTracePattern, err := regexp.Compile("\[line (\d+)\]")
//nonTestPattern, err := regexp.Compile()r"// nontest")

const Red = "\033[31m"
const Reset = "\033[0m"
const Green = "\033[32m"

var suites = map[string]Suite{}

var suite Suite

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
	//var cSuite = Suite{
	//	executable: cInterpreter,
	//	tests:      tests,
	//}
	//suites["c"] = cSuite

	var javaSuite = Suite{
		executable: javaInterpreter,
		tests:      tests,
	}
	suites["java"] = javaSuite
}

func runSuites(names []string) bool {
	successful := true
	for _, name := range names {
		if !runSuite(name) {
			successful = false
		}
	}
	return successful
}

func runSuite(name string) bool {
	log.Printf("====== Suite: %s ======", name)

	suite = suites[name]
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

	err := test.parse()
	if err != nil {
		log.Panicln(err)
	}

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
	line   int
	output string
}

type TestParseError struct {
	path string
}

func (e *TestParseError) Error() string {
	return fmt.Sprintf("%s failed to parse", e.path)
}

type Test struct {
	path           string
	expectedOutput []ExpectedOutput
	expectedErrors []error
	failures       []string
}

func (t *Test) run() []string {
	lox := exec.Command(suite.executable, "-jar", "C:\\Users\\willi\\Code\\lox\\out\\artifacts\\lox_jar\\lox.jar", t.path)
	bytes, err := lox.Output()
	if err != nil {
		log.Panicln(err)
	}
	result := string(bytes)
	outputLines := strings.Split(result, "\n")
	t.validateOutput(outputLines)
	return t.failures
}

func (t *Test) validateOutput(lines []string) {
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	index := 0
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if index >= len(t.expectedOutput) {
			t.fail(fmt.Sprintf("Got output %s when none was expected", line), []string{})
		}

		expected := t.expectedOutput[index]
		if line != expected.output {
			t.fail(fmt.Sprintf("Expected output %s on line %d but got line %s", expected.output, expected.line, line), []string{})
		}
		index++
	}
}

func (t *Test) fail(message string, lines []string) {
	t.failures = append(t.failures, message)
	for _, line := range lines {
		t.failures = append(t.failures, line)
	}
}

func (t *Test) parse() error {
	file, err := os.Open(t.path)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	lineNum := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		nonTestPattern, err := regexp.Compile("// nontest")
		if err != nil {
			return &TestParseError{t.path}
		}
		if nonTestPattern.MatchString(line) {
			return nil
		}

		expectedOutputPattern, err := regexp.Compile("// expect: ?(.*)")
		if err != nil {
			return &TestParseError{t.path}
		}
		match := expectedOutputPattern.FindStringSubmatch(line)
		if len(match) > 0 {
			expectedOutput := ExpectedOutput{
				line:   lineNum,
				output: match[1],
			}
			t.expectedOutput = append(t.expectedOutput, expectedOutput)
		}
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return nil
}
