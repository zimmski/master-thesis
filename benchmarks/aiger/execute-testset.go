package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

const aigerToolsetPath = "./aiger/"
const testExtension = ".test"

type command struct {
	Name    string
	Command []string
}

type Coverage struct {
	Total      string
	Missed     string
	Percentage string
}

type Result struct {
	Lines   Coverage
	Regions Coverage

	TestCases         int
	AddressSanitizer  int
	ExitStatusNotZero int
	FileNotApplicable int
	InvalidFile       int
}

// results[Command][Testset]
var results = map[string]map[string]Result{}

func main() {
	executeTestSet("./testset-aigfuzz-random")
	executeTestSet("./testset-tavor-random")
	// TODO // For testset-tavor-AllPermutations

	fmt.Println("command;testset;test-cases;lines-total;lines-missed;lines-percentage;regions-total;regions-missed;regions-percentage;AddressSanitizer;ExitStatusNotZero;FileNotApplicable;InvalidFile")
	for c, ts := range results {
		for t, r := range ts {
			fmt.Printf("%s;%s;%d;%s;%s;%s;%s;%s;%s;%d;%d;%d;%d\n", c, t, r.TestCases, r.Lines.Total, r.Lines.Missed, r.Lines.Percentage, r.Regions.Total, r.Regions.Missed, r.Regions.Percentage, r.AddressSanitizer, r.ExitStatusNotZero, r.FileNotApplicable, r.InvalidFile)
		}
	}
}

var currentCount int
var overallCount int

func executeTestSet(testSetPath string) {
	d, err := os.Open(testSetPath)
	if err != nil {
		panic(err)
	}

	fs, err := d.Readdirnames(-1)
	if err != nil {
		panic(err)
	}

	testSetTestCounter := 0

	// Cleanup old files
	for _, f := range fs {
		if strings.HasSuffix(f, testExtension) {
			testSetTestCounter++
		}

		if !strings.HasSuffix(f, ".profdata") && !strings.HasSuffix(f, ".profraw") && // Coverage files
			!strings.HasSuffix(f, ".aig") { // Generated files
			continue
		}

		err := os.Remove(testSetPath + "/" + f)
		if err != nil {
			panic(err)
		}
	}

	cmds := []command{
		command{
			Name:    "aigand",
			Command: []string{aigerToolsetPath + "aigand", "$TESTFILE"},
		},
		command{
			Name:    "aigbmc",
			Command: []string{aigerToolsetPath + "aigbmc", "-m", "$TESTFILE"},
		},
		command{
			Name:    "aigflip",
			Command: []string{aigerToolsetPath + "aigflip", "$TESTFILE"},
		},
		command{
			Name:    "aiginfo",
			Command: []string{aigerToolsetPath + "aiginfo", "$TESTFILE"},
		},
		// command{ TODO infinity loop?
		// 	Name:          "aigjoin",
		// 	Command:       []string{aigerToolsetPath + "aigjoin", "$TESTFILE"},
		// },
		command{
			Name:    "aigmiter",
			Command: []string{aigerToolsetPath + "aigmiter", "$TESTFILE", "$TESTFILE"},
		},
		command{
			Name:    "aigmove", // TODO does not work without the -r argument
			Command: []string{aigerToolsetPath + "aigmove", "-r", "$TESTFILE"},
		},
		command{
			Name:    "aignm",
			Command: []string{aigerToolsetPath + "aignm", "$TESTFILE"},
		},
		command{
			Name:    "aigor",
			Command: []string{aigerToolsetPath + "aigor", "$TESTFILE"},
		},
		command{
			Name:    "aigsim",
			Command: []string{aigerToolsetPath + "aigsim", "-r", "5", "$TESTFILE"}, // random stimulus of 5 input vectors
		},
		command{
			Name:    "aigsplit",
			Command: []string{aigerToolsetPath + "aigsplit", "-f", "$TESTFILE"},
		},
		// command{
		// 	Name:            "aigstrip", Changes the input file so we cannot safely use the command for this benchmark.
		// },
		command{
			Name:    "aigtoaig",
			Command: []string{aigerToolsetPath + "aigtoaig", "$TESTFILE"},
		},
		command{
			Name:    "aigtoblif",
			Command: []string{aigerToolsetPath + "aigtoblif", "$TESTFILE"},
		},
		command{
			Name:    "aigtocnf",
			Command: []string{aigerToolsetPath + "aigtocnf", "$TESTFILE"},
		},
		command{
			Name:    "aigtodot",
			Command: []string{aigerToolsetPath + "aigtodot", "$TESTFILE"},
		},
		command{
			Name:    "aigtosmv",
			Command: []string{aigerToolsetPath + "aigtosmv", "$TESTFILE"},
		},
		command{
			Name:    "aigunroll",
			Command: []string{aigerToolsetPath + "aigunroll", "5", "$TESTFILE"}, // bound of 5 unrollings
		},
	}

	currentCount = 0
	overallCount = len(cmds) * testSetTestCounter

	for _, c := range cmds {
		executeCommandOnTestSet(testSetPath, fs, c)
	}
}

func executeCommandOnTestSet(testSetPath string, fs []string, c command) {
	firstProfile := true
	combinedProfileFilePath := testSetPath + "/combined-" + c.Name + ".profdata"

	var r Result

	for _, f := range fs {
		if !strings.HasSuffix(f, testExtension) {
			continue
		}

		hash := strings.TrimSuffix(f, testExtension)

		currentCount++
		fmt.Printf("%d/%d: Execute %q on test case %q\n", currentCount, overallCount, c.Name, f)

		cmd := append([]string(nil), c.Command...)

		for i := range cmd {
			if cmd[i] == "$TESTFILE" {
				cmd[i] = testSetPath + "/" + f
			}
		}

		// Execute test case
		e := exec.Command(cmd[0], cmd[1:]...)
		profRawFilePath := testSetPath + "/" + hash + "." + c.Name + ".profraw"
		e.Env = []string{
			"LLVM_PROFILE_FILE=" + profRawFilePath,
		}

		out, err := e.CombinedOutput()
		sout := string(out)

		if strings.Contains(sout, "AddressSanitizer") {
			r.AddressSanitizer++
		}
		r.TestCases++

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					fmt.Printf("\tExit status is not 0, was %d\n", status.ExitStatus())

					r.ExitStatusNotZero++

					if strings.Contains(sout, "cyclic definition for and gate") {
						r.InvalidFile++
					} else if strings.Contains(sout, "more than one output") || strings.Contains(sout, "can not handle latches") || strings.Contains(sout, "without outputs") || strings.Contains(sout, "no output") {
						r.FileNotApplicable++
					} else if strings.Contains(sout, "AddressSanitizer") && (c.Name == "aigflip") {
						// Ignore
					} else {
						fmt.Printf("%s\n\n", sout)

						panic("unknown exit status")
					}
				}
			} else {
				panic(err)
			}
		}

		// Merge coverage profiles and index them
		if firstProfile {
			_, err := exec.Command("llvm-profdata-4.0.1", "merge", "-sparse", profRawFilePath, "-o", combinedProfileFilePath).CombinedOutput()
			if err != nil {
				panic(err)
			}

			firstProfile = false
		} else {
			_, err := exec.Command("llvm-profdata-4.0.1", "merge", "-sparse", profRawFilePath, combinedProfileFilePath, "-o", combinedProfileFilePath).CombinedOutput()
			if err != nil {
				panic(err)
			}
		}

		err = os.Remove(profRawFilePath)
		if err != nil {
			panic(err)
		}

		// Cleanup aig files
		d, err := os.Open(testSetPath)
		if err != nil {
			panic(err)
		}

		aigs, err := d.Readdirnames(-1)
		if err != nil {
			panic(err)
		}

		for _, f := range aigs {
			if !strings.HasSuffix(f, ".aig") {
				continue
			}

			err := os.Remove(testSetPath + "/" + f)
			if err != nil {
				panic(err)
			}
		}
	}

	coverageOutput, err := exec.Command("llvm-cov-4.0.1", "report", c.Command[0], "-instr-profile="+combinedProfileFilePath).CombinedOutput()
	if err != nil {
		panic(err)
	}

	matches := regexp.MustCompile(`TOTAL\s+(\d+)\s+(\d+)\s+(\d+\.\d+%)\s+(\d+)\s+(\d+)\s+(\d+\.\d+%)\s+(\d+)\s+(\d+)\s+(\d+\.\d+%)\s+(\d+)\s+(\d+)\s+(\d+\.\d+%)`).FindStringSubmatch(string(coverageOutput))

	r.Regions = Coverage{
		Total:      matches[1],
		Missed:     matches[2],
		Percentage: matches[3],
	}
	r.Lines = Coverage{
		Total:      matches[10],
		Missed:     matches[11],
		Percentage: matches[12],
	}

	if _, ok := results[c.Name]; !ok {
		results[c.Name] = map[string]Result{}
	}
	results[c.Name][testSetPath] = r
}
