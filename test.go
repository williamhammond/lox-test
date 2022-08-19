package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type TestParseError struct {
	path string
}

func (e *TestParseError) Error() string {
	return fmt.Sprintf("%s failed to parse", e.path)
}

type ExpectedOutput struct {
	line   int
	output string
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

func (t *Test) run(lox *exec.Cmd) []string {
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

	stdoutSring := strings.ReplaceAll(string(stdout.Bytes()), "\r", "")
	outputLines := strings.Split(stdoutSring, "\n")

	stderrSring := strings.ReplaceAll(string(stderr.Bytes()), "\r", "")
	errorLines := strings.Split(stderrSring, "\n")

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

func (t *Test) parse(language string) error {
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
		if err != nil {
			return &TestParseError{t.path}
		}
		if nonTestPattern.MatchString(line) {
			return nil
		}

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

		match = expectedErrorPattern.FindStringSubmatch(line)
		if len(match) > 0 {
			expectedError := fmt.Sprintf("[%d] %s", lineNum, match[1])
			t.expectedErrors = append(t.expectedErrors, expectedError)

			t.expectedExitCode = compileErrorCode
			t.expectations++
			lineNum++
			continue
		}

		match = errorLinePattern.FindStringSubmatch(line)
		if len(match) > 0 {
			errorLanguage := match[2]
			if errorLanguage == "" || errorLanguage == language {
				t.expectedErrors = append(t.expectedErrors, fmt.Sprintf("[%s] %s", match[3], match[4]))
				t.expectedExitCode = compileErrorCode
				t.expectations++
			}
		}

		match = expectedRuntimeErrorPattern.FindStringSubmatch(line)
		if len(match) > 0 {
			t.runtimeErrorLine = lineNum
			t.expectedRuntimeError = match[1]
			t.expectedExitCode = runtimeErrorCode
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
