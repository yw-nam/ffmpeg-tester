package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var inputFile string = "input.mp4"
var inputDir string = "./testdata"

func swOptFmt(bitrate int) string {
	return fmt.Sprintf("-y -i %%s -c:v libx264 -b:v %dM -maxrate:v %dM -minrate:v %dM -bufsize %dM -x264-params nal_hrd=cbr -c:a aac -b:a 96k  -movflags faststart %%s", bitrate, bitrate, bitrate, bitrate*2)
}

func hwOptFmt(bitrate int, gpuId int) string {
	return fmt.Sprintf("-y -i %%s -c:v h264_nvenc -b:v %dM -maxrate:v %dM -bufsize %dM -preset p6 -rc cbr -gpu %d -c:a aac -b:a 96k  -movflags faststart  %%s", bitrate, bitrate, bitrate*2, gpuId)
}

type testCase struct {
	gpuNum  int
	bitrate int
	setNum  int
}

func (tc *testCase) getTitle() string {
	lib := "libx264"
	if tc.gpuNum > 0 {
		lib = "h264_nvenc"
	}
	return fmt.Sprintf("%s_%dM_%02dea", lib, tc.bitrate, tc.setNum)
}

func makeCases() []testCase {
	result := []testCase{}
	for _, bitrate := range []int{1, 2, 4} {
		for _, gpuNum := range []int{0, 1, 2} {
			setNums := []int{}
			switch gpuNum {
			case 0:
				setNums = []int{1, 2, 4}
			case 1:
				setNums = []int{1, 2, 4, 8, 16}
			case 2:
				setNums = []int{8, 16}
			}
			for _, setNum := range setNums {
				tc := testCase{
					bitrate: bitrate,
					gpuNum:  gpuNum,
					setNum:  setNum,
				}
				result = append(result, tc)
			}
		}
	}
	return result
}

func main() {
	readFlag()

	if err := os.Chdir(inputDir); err != nil {
		log.Fatal("Error changing directory:", err)
	}
	os.Mkdir("output", 0777)

	fmt.Printf(">> start with %s/%s\n", inputDir, inputFile)
	originalStdOut := os.Stdout

	if resultFile, err := changeStdOut(); err != nil {
		defer resultFile.Close()
	}

	testCases := makeCases()
	for _, cases := range testCases {
		runCases(cases)
	}

	fmt.Println("file end")
	os.Stdout = originalStdOut
	fmt.Println(">> done")
}

func runCases(tc testCase) {
	fmt.Printf("\n--- setNum:%d / bitRate:%d / gpuNum:%d ----------------------------\n", tc.setNum, tc.bitrate, tc.gpuNum)

	caseName := tc.getTitle()
	optFmts := map[string]string{}
	for i := 0; i < tc.setNum; i++ {
		if tc.gpuNum == 0 {
			optFmt := swOptFmt(tc.bitrate)
			outputFile := fmt.Sprintf("./output/%s/%02d.mp4", caseName, i)
			optFmts[outputFile] = fmt.Sprintf(optFmt, inputFile, outputFile)
		} else {
			gpuId := i % tc.gpuNum
			optFmt := hwOptFmt(tc.bitrate, gpuId)
			outputFile := fmt.Sprintf("./output/%s_%dgpu/%02d.mp4", caseName, tc.gpuNum+1, i)
			optFmts[outputFile] = fmt.Sprintf(optFmt, inputFile, outputFile)
		}
	}

	var wg sync.WaitGroup
	for outputFile, option := range optFmts {
		wg.Add(1)
		go encoding(option, outputFile, &wg)
	}
	wg.Wait()
}

func encoding(option string, outputFile string, wg *sync.WaitGroup) {
	defer wg.Done()

	dirName := getOutputDir(outputFile)
	removeAll(dirName)
	os.Mkdir(dirName, 0777)

	start := time.Now()

	cmd := exec.Command("ffmpeg", strings.Split(option, " ")...)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("done[%s]: %v\n", outputFile, time.Since(start))
}

func readFlag() {
	inputDirFlag := flag.String("input_dir", "./testdata", "-input_dir=./testdata")
	inputFileFlag := flag.String("input_file", "input.mp4", "-input_file=input.mp4")
	flag.Parse()
	inputFile = *inputFileFlag
	inputDir = *inputDirFlag
}

func changeStdOut() (*os.File, error) {
	resultFilePath := fmt.Sprintf("./output/result_%s.txt", time.Now().Format("0102_150405"))
	if resultFile, err := os.Create(resultFilePath); err != nil {
		return nil, err
	} else {
		os.Stdout = resultFile
		return resultFile, nil
	}
}

func getOutputDir(fullPath string) string {
	re := regexp.MustCompile(`^(.*)\/\d+.mp4$`)
	match := re.FindStringSubmatch(fullPath)
	return match[1]
}

func removeAll(path string) error {
	files, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			return err
		}
	}

	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}
