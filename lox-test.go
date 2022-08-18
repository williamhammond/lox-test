package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
)

func main() {
	var path = flag.String("path", "", "Path to lox executable")
	flag.Parse()
	if _, err := os.Stat(*path); errors.Is(err, os.ErrNotExist) {
		log.Panicln(err)
	}

	log.Printf("running: %s", *path)
	lox := exec.Command(*path)
	bytes, err := lox.Output()
	if err != nil {
		log.Panicln(err)
	}
	result := string(bytes)
	log.Println(result)
}
