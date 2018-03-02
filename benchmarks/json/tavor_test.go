package json

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func testCase(t *testing.T, c []byte) {
	var o interface{}

	err := Unmarshal(c, &o)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
}

func readTestCases(path string) [][]byte {
	var cases [][]byte

	d, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	fs, err := d.Readdirnames(-1)
	if err != nil {
		panic(err)
	}

	for _, f := range fs {
		if strings.HasSuffix(f, ".test") {
			b, err := ioutil.ReadFile(path + "/" + f)
			if err != nil {
				panic(err)
			}

			cases = append(cases, b)
		}
	}

	return cases
}

func TestTavorMaxRepeat1(t *testing.T) {
	for _, c := range readTestCases("/home/symflower/json/json/testset-max-repeat-1/") {
		testCase(t, c)
	}
}

func TestTavorMaxRepeat2(t *testing.T) {
	for _, c := range readTestCases("/home/symflower/json/json/testset-max-repeat-2/") {
		testCase(t, c)
	}
}

func TestTavorMaxRepeat3(t *testing.T) {
	for _, c := range readTestCases("/home/symflower/json/json/testset-max-repeat-3/") {
		testCase(t, c)
	}
}
