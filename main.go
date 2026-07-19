package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("--> ")

	cleanEnv := flag.Bool("clean", false, "use minimal env vars (home, user, path, term, pwd, lang)")
	flag.Parse()

	commandArguments := flag.Args()
	if len(commandArguments) == 0 {
		fmt.Printf("usage: %s <command>\n", os.Args[0])
		os.Exit(1)
	}

	environment := make([]string, 0, 5)
	if *cleanEnv {
		defaultEnvs := []string{"HOME", "USER", "PWD", "TERM", "PATH", "LANG"}
		for _, de := range defaultEnvs {
			environment = append(environment, de+"="+os.Getenv(de))
		}
	} else {
		environment = os.Environ()
	}
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
		environment = append(environment, key+"="+val)
		envKeys = append(envKeys, key)
	}
	if bf.Err() != nil {
		log.Fatalf("env parse failed: %s", err)
	}
	environment, err = dedupEnv(environment)
	if err != nil {
		log.Fatalf("environment dedup failure: %s", err)
	}
	log.Printf(`read %d variables from ".env": %s`, len(envKeys), strings.Join(envKeys, ", "))

	// finally resolve command from PATH and Exec()
	lp, err := exec.LookPath(commandArguments[0])
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("executing app: %s", lp)
	if err := syscall.Exec(lp, commandArguments, environment); err != nil {
		log.Fatalf("syscall.Exec() failed: %s", err)
	}
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
