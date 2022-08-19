package main

import "log"

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
