package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var inputDir string = "./testdata"

var bitRateToInputFile map[int]string = map[int]string{
	1: "720p-EARTH.mp4",
	2: "1080p-birds.mp4",
	4: "2160p-color.mp4",
}

func swOptFmt(bitrate int) string {
	return fmt.Sprintf("-y -i %%s -c:v libx264 -b:v %dM -maxrate:v %dM -minrate:v %dM -bufsize %dM -x264-params nal_hrd=cbr -c:a aac -b:a 96k -movflags faststart %%s", bitrate, bitrate, bitrate, bitrate*2)
}

func hwOptFmt(bitrate int, gpuId int) string {
	return fmt.Sprintf("-y -i %%s -c:v h264_nvenc -b:v %dM -maxrate:v %dM -bufsize %dM -preset p6 -rc cbr -gpu %d -c:a aac -b:a 96k -movflags faststart %%s", bitrate, bitrate, bitrate*2, gpuId)
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

func getOutPutDir(inputFile string) string {
	fileName := strings.Split(inputFile, ".")
	return fmt.Sprintf("output_%s", fileName[0])
}

func makeOutputDirs() {
	for _, inputFile := range bitRateToInputFile {
		outputDir := getOutPutDir(inputFile)
		removeAll(outputDir)
		os.Mkdir(outputDir, 0777)
	}
}

func main() {
	readFlag()
	if err := os.Chdir(inputDir); err != nil {
		log.Fatal("Error changing directory:", err)
	}
	makeOutputDirs()

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
	inputFile := bitRateToInputFile[tc.bitrate]
	outputDir := getOutPutDir(inputFile)
	optFmts := map[string]string{}
	for i := 0; i < tc.setNum; i++ {
		if tc.gpuNum == 0 {
			optFmt := swOptFmt(tc.bitrate)
			outputFile := fmt.Sprintf("./%s/%s_%02d.mp4", outputDir, caseName, i)
			optFmts[outputFile] = fmt.Sprintf(optFmt, inputFile, outputFile)
		} else {
			gpuId := i % tc.gpuNum
			optFmt := hwOptFmt(tc.bitrate, gpuId)
			outputFile := fmt.Sprintf("./%s/%s_%dgpu_%02d.mp4", outputDir, caseName, tc.gpuNum, i)
			optFmts[outputFile] = fmt.Sprintf(optFmt, inputFile, outputFile)
		}
	}

	var wg sync.WaitGroup
	start := time.Now()
	for outputFile, option := range optFmts {
		wg.Add(1)
		go encoding(option, outputFile, &wg)
	}
	wg.Wait()
	fmt.Printf(">> set done: %s\n", time.Since(start))
}

func encoding(option string, outputFile string, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Println(option)

	start := time.Now()
	cmd := exec.Command("ffmpeg", strings.Split(option, " ")...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command failed with error: %s\n", err)
		log.Fatalf("Output:\n%s", string(output))
	}
	fmt.Printf("done[%s]: %v\n", outputFile, time.Since(start))
}

func readFlag() {
	inputDirFlag := flag.String("input_dir", "./testdata", "-input_dir=./testdata")
	flag.Parse()
	inputDir = *inputDirFlag
}

func changeStdOut() (*os.File, error) {
	resultFilePath := fmt.Sprintf("./result_%s.txt", time.Now().Format("0102_150405"))
	if resultFile, err := os.Create(resultFilePath); err != nil {
		return nil, err
	} else {
		os.Stdout = resultFile
		return resultFile, nil
	}
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
