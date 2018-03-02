package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

const TestCaseCount = 5

func main() {
	generateTestSet("testset-aigfuzz-random", []string{"/home/zimmski/master-thesis/benchmarks/aiger/aiger/aigfuzz", "-a"})
	generateTestSet("testset-tavor-random", []string{"tavor", "--format-file", "/home/zimmski/master-thesis/benchmarks/aiger/aag.tavor", "fuzz", "--max-repeat", "10"})
	// TODO // For testset-tavor-AllPermutations
}

func generateTestSet(path string, cmd []string) {
	then := time.Now()

	hashs := map[string]bool{}

	err := os.Mkdir(path, 0700)
	if err != nil {
		panic(err)
	}

	for len(hashs) < TestCaseCount {
		data, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			fmt.Printf("ERROR exec: %s\n", err.Error())

			continue
		}

		hash := fmt.Sprintf("%x", md5.Sum(data))
		if _, ok := hashs[hash]; ok {
			fmt.Printf("Generated duplicate %s\n", hash)

			continue
		}
		hashs[hash] = true

		filepath := path + "/" + hash + ".test"

		err = ioutil.WriteFile(filepath, data, 0600)
		if err != nil {
			fmt.Printf("ERROR write test case: %s\n", err.Error())

			continue
		}

		fmt.Printf("Wrote test case %s\n", filepath)
	}

	now := time.Now()

	fmt.Printf("Generated %s in %s\n", path, now.Sub(then).String())
}
