package main

import "regexp"

var expectedOutputPattern = regexp.MustCompile("// expect: ?(.*)")
var expectedErrorPattern = regexp.MustCompile("// (Error.*)")
var errorLinePattern = regexp.MustCompile("// \\[((java|c) )?line (\\d+)\\] (Error.*)")
var expectedRuntimeErrorPattern = regexp.MustCompile("// expect runtime error: (.+)")
var syntaxErrorPattern = regexp.MustCompile("\\[.*line (\\d+)\\] (Error.+)")
var stackTracePattern = regexp.MustCompile("[line (d+)]")
var nonTestPattern = regexp.MustCompile("// nontest")
