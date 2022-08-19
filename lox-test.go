package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const cInterpreter = "C:\\Users\\willi\\Code\\clox\\cmake-build-debug\\clox.exe"
const javaInterpreter = "C:\\Program Files\\Common Files\\Oracle\\Java\\javapath\\java.exe"

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
	name       string
	executable string
	language   string
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
	path                 string
	expectedRuntimeError string
	expectedExitCode     int
	runtimeErrorLine     int
	expectations         int
	expectedOutput       []ExpectedOutput
	expectedErrors       []string
	failures             []string
}

func (t *Test) run() []string {
	lox := exec.Command(suite.executable, "-jar", "C:\\Users\\willi\\Code\\lox\\out\\artifacts\\lox_jar\\lox.jar", t.path)

	var stdout, stderr bytes.Buffer
	lox.Stdout = &stdout
	lox.Stderr = &stderr

	err := lox.Run()
	var exitCode = 0
	if exitError, ok := err.(*exec.ExitError); ok {
		// very fragile, error message is "Status Code X"
		// It seems like go doesn't support returning exit codes?
		s := strings.Split(exitError.Error(), " ")[2]
		if s != "0" {
			exitCode, err = strconv.Atoi(s)
			if err != nil {
				panic(err)
			}
		}
	}

	outputLines := strings.Split(string(stdout.Bytes()), "\n")
	errorLines := strings.Split(string(stderr.Bytes()), "\n")

	if t.expectedRuntimeError != "" {
		t.validateRuntimeError(errorLines)
	} else {
		t.validateCompileErrors(errorLines)
	}
	t.validateExitCode(exitCode, errorLines)
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

// expectedOutputPattern, err := regexp.Compile("// expect: ?(.*)")
// expectedErrorPattern, err  := regexp.Compile("// (Error.*)")
// errorLinePattern, err := regexp.Compile("// \[((java|c) )?line (\d+)\] (Error.*)")
// expectedRuntimeErrorPattern, err := regexp.Compile("// expect runtime error: (.+)")
// syntaxErrorPattern, err := regexp.Compile("\[.*line (\d+)\] (Error.+)")
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

		expectedOutputPattern := regexp.MustCompile("// expect: ?(.*)")
		match := expectedOutputPattern.FindStringSubmatch(line)
		if len(match) > 0 {
			expectedOutput := ExpectedOutput{
				line:   lineNum,
				output: match[1],
			}
			t.expectedOutput = append(t.expectedOutput, expectedOutput)
			t.expectations++
			lineNum++
			continue
		}

		expectedErrorPattern := regexp.MustCompile("// (Error.*)")
		match = expectedErrorPattern.FindStringSubmatch(line)
		if len(match) > 0 {
			expectedError := fmt.Sprintf("[%d] %s", lineNum, match[1])
			t.expectedErrors = append(t.expectedErrors, expectedError)

			t.expectedExitCode = 65
			t.expectations++
			lineNum++
			continue
		}

		errorLinePattern := regexp.MustCompile("// \\[((java|c) )?line (\\d+)\\] (Error.*)")
		match = errorLinePattern.FindStringSubmatch(line)
		if len(match) > 0 {
			language := match[2]
			if language == "" || language == suite.language {
				t.expectedErrors = append(t.expectedErrors, fmt.Sprintf("[%s] %s", match[3], match[4]))
				t.expectedExitCode = 65
				t.expectations++
			}
		}

		expectedRuntimeErrorPattern := regexp.MustCompile("// expect runtime error: (.+)")
		match = expectedRuntimeErrorPattern.FindStringSubmatch(line)
		if len(match) > 0 {
			t.runtimeErrorLine = lineNum
			t.expectedRuntimeError = match[1]
			t.expectedExitCode = 70
			t.expectations++
		}

		if len(t.expectedErrors) > 0 && t.expectedRuntimeError != "" {
			return &TestParseError{t.path}
		}

		lineNum++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return nil
}

func (t *Test) validateExitCode(exitCode int, lines []string) {
	if t.expectedExitCode == exitCode {
		return
	}

	if len(lines) > 10 {
		lines = lines[:10]
		lines = append(lines, "(truncated...)")
	}

	t.fail("Expected return code $_expectedExitCode and got $exitCode. Stderr:", lines)

}

func (t *Test) validateRuntimeError(lines []string) {
	if len(lines) < 2 {
		t.fail(fmt.Sprintf("Expected runtime error %s but got none", t.expectedRuntimeError), []string{})
		return
	}

	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if lines[0] != t.expectedRuntimeError {
		t.fail(fmt.Sprintf("Expected runtime error %s and got:", t.expectedRuntimeError), []string{})
		t.fail(lines[0], []string{})
	}

	var stackLines = lines[1:]
	stackTracePattern := regexp.MustCompile("[line (d+)]")

	var foundStackTrace = false
	for _, line := range stackLines {
		line = strings.TrimSuffix(line, "\r")
		match := stackTracePattern.MatchString(line)
		if match {
			foundStackTrace = true
			break
		}
	}

	if !foundStackTrace {
		t.fail(fmt.Sprintf("Expected stack trace and got: "), lines)
	} else {
	}
}

func (t *Test) validateCompileErrors(lines []string) {
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	foundErrors := []string{}
	unexpectedCount := 0
	syntaxErrorPattern := regexp.MustCompile("\\[.*line (\\d+)\\] (Error.+)")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		match := syntaxErrorPattern.FindStringSubmatch(line)
		syntaxErrorPattern.MatchString(line)
		if len(match) > 0 {
			err := fmt.Sprintf("[%s] %s", match[1], match[2])
			containsError := false
			for _, expectedError := range t.expectedErrors {
				if expectedError == err {
					containsError = true
				}
			}
			if containsError {
				foundErrors = append(foundErrors, err)
			} else {
				if unexpectedCount < 10 {
					t.fail("Unexpected output on std err", []string{})
					t.fail(line, []string{})
				}
				unexpectedCount++
			}
		} else if line != "" {
			if unexpectedCount < 10 {
				t.fail("Unexpected output on std err", []string{})
				t.fail(line, []string{})
			}
			unexpectedCount++
		}
		if unexpectedCount > 10 {
			t.fail("(truncated ${unexpectedCount - 10} more...)", []string{})
		}
	}

	for _, err := range difference(t.expectedErrors, foundErrors) {
		t.fail(fmt.Sprintf("Missing expected error %s", err), []string{})
	}
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
