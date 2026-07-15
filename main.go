package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Printf("usage: %s <command>\n", os.Args[0])
		os.Exit(1)
	}

	// init command struct
	c := os.Args[1]
	args := os.Args[2:]
	cmd := exec.Command(c, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	currentEnv := os.Environ()
	cmd.Env = currentEnv

	// read .env file and append to command struct Env field
	f, err := os.ReadFile(".env")
	if err != nil {
		log.Fatalf("failed: %s", err)
	}
	bf := bufio.NewScanner(bytes.NewReader(f))
	for bf.Scan() {
		s := bf.Text()

		key, val, ok := strings.Cut(s, "=")
		if !ok {
			continue
		}
		cmd.Env = append(cmd.Env, key+"="+val)
	}
	if bf.Err() != nil {
		log.Fatalf("env parse failed: %s", err)
	}

	// start program and wait
	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd run failed: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		ee := err.(*exec.ExitError)
		log.Printf("process exit code %d", ee.ExitCode())
		os.Exit(ee.ExitCode())
	}
	os.Exit(0)
}
