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
	log.SetFlags(0)
	log.SetPrefix("--> ")
	if len(os.Args) < 2 {
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
	envKeys := []string{}
	bf := bufio.NewScanner(bytes.NewReader(f))
	for bf.Scan() {
		s := bf.Text()

		key, val, ok := getKeyVal(s)
		if !ok {
			continue
		}
		cmd.Env = append(cmd.Env, key+"="+val)
		envKeys = append(envKeys, key)
	}
	if bf.Err() != nil {
		log.Fatalf("env parse failed: %s", err)
	}
	log.Printf("read %d .env variables: %s", len(envKeys), strings.Join(envKeys, ", "))
	log.Printf("running prog %s...", cmd.Path)
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

func getKeyVal(line string) (string, string, bool) {
	key, val, ok := strings.Cut(line, "=")
	key = strings.TrimSpace(key)
	val = strings.TrimSpace(val)

	// skip empty keys, values and comments
	if len(key) == 0 || len(val) == 0 || strings.HasPrefix(key, "#") {
		return key, val, false
	}

	if strings.HasPrefix(val, `"`) || strings.HasSuffix(val, `"`) {
		panic(".env file contains quoted values, please remove them")
	}

	return key, val, ok
}
