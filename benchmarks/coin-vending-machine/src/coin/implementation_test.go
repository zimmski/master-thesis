package coin

import (
	"os"
	"strings"
	"testing"

	"github.com/zimmski/tavor/executor/keydriven"
)

func TestImplementation(t *testing.T) {
	testsetPath := os.Getenv("TESTSET")
	if testsetPath == "" {
		testsetPath = "."
	}

	d, err := os.Open(testsetPath)
	if err != nil {
		panic(err)
	}

	fs, err := d.Readdirnames(-1)
	if err != nil {
		panic(err)
	}

	testCount := 0
	for _, f := range fs {
		if !strings.HasSuffix(f, ".test") {
			continue
		}

		cmds, err := keydriven.ReadKeyDrivenFile(testsetPath + "/" + f)
		if err != nil {
			t.Fatalf("Error: %v\n", err)
		}

		executor := initExecutor()

		if err := executor.Execute(cmds); err != nil {
			t.Fatalf("Failed: %v\n", err)
		}

		testCount++
	}

	if testCount == 0 {
		t.Fatalf("No test files where found in %q", testsetPath)
	}
}
