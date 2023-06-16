package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var inputFile string = "input.mp4"
var inputDir string = "./testdata"

var caseSample map[string]string = map[string]string{
	"sample_01": "-i %s -c:v libx264 -preset fast -an %s",
	"sample_02": "-i %s -c:v libx264 -preset fast -an %s",
}

func swOptFmt(bitrate int) string {
	return fmt.Sprintf("-y -i %%s -c:v libx264 -b:v %dM -maxrate:v %dM -minrate:v %dM -bufsize %dM -x264-params nal_hrd=cbr -c:a aac -b:a 96k  -movflags faststart %%s", bitrate, bitrate, bitrate, bitrate*2)
}

func hwOptFmt(bitrate int, gpuId int) string {
	return fmt.Sprintf("-y -i %%s -c:v h264_nvenc -b:v %dM -maxrate:v %dM -bufsize %dM -preset p6 -rc cbr -gpu %d -c:a aac -b:a 96k  -movflags faststart  %%s", bitrate, bitrate, bitrate*2, gpuId)
}

type testCase struct {
	gpuNum  int
	bitrate int
	howMany int
}

func (tc *testCase) getTitle() string {
	lib := "libx264"
	if tc.gpuNum > 0 {
		lib = "h264_nvenc"
	}
	return fmt.Sprintf("%s_%dM_%dea", lib, tc.bitrate, tc.howMany)
}

func makeCases() []testCase {
	result := []testCase{}
	for _, bitrate := range []int{1, 2, 4} {
		for _, gpuNum := range []int{0, 1, 2} {
			howManyArr := []int{}
			switch gpuNum {
			case 0:
				howManyArr = []int{1, 2, 4}
			case 1:
				howManyArr = []int{1, 2, 4, 8, 16}
			case 2:
				howManyArr = []int{8, 16}
			}
			for _, howMany := range howManyArr {
				tc := testCase{
					bitrate: bitrate,
					gpuNum:  gpuNum,
					howMany: howMany,
				}
				result = append(result, tc)
			}
		}
	}
	return result
}

func main() {
	readFlag()
	fmt.Printf(">> start with %s/%s\n", inputDir, inputFile)
	originalStdOut := os.Stdout
	makeOutputDir()
	// changeStdOut()

	testCases := makeCases()
	for _, cases := range testCases {
		runCases(cases)
	}

	fmt.Println("file end")
	os.Stdout = originalStdOut
	fmt.Println(">> done")
}

func runCases(tc testCase) {
	var wg sync.WaitGroup
	caseName := tc.getTitle()
	optFmts := []string{}

	for i := 0; i < tc.howMany; i++ {
		if tc.gpuNum == 0 {
			optFmts = append(optFmts, swOptFmt(tc.bitrate))
		} else {
			gpuId := i % 2
			optFmts = append(optFmts, hwOptFmt(tc.bitrate, gpuId))
		}
	}

	for _, optFmt := range optFmts {
		wg.Add(1)
		outputFile := fmt.Sprintf("./output/output_%s.mp4", caseName)
		option := fmt.Sprintf(optFmt, inputFile, outputFile)
		go encoding(option, outputFile, &wg)
	}
	wg.Wait()
}

func encoding(option string, outputFile string, wg *sync.WaitGroup) {
	defer wg.Done()
	os.Remove(outputFile)

	start := time.Now()

	// TODO 일단 출력만!!!!!!>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> 파일 이름이나 경로 바꿔줘야 함!!!!!!
	fmt.Println(option)
	// cmd := exec.Command("ffmpeg", strings.Split(option, " ")...)
	// if err := cmd.Run(); err != nil {
	// 	log.Fatal(err)
	// }
	fmt.Printf("done[%s]: %v\n", outputFile, time.Since(start))
}

func readFlag() {
	inputDirFlag := flag.String("input_dir", "./testdata", "-input_dir=./testdata")
	inputFileFlag := flag.String("input_file", "input.mp4", "-input_file=input.mp4")
	flag.Parse()
	inputFile = *inputFileFlag
	inputDir = *inputDirFlag
}

func makeOutputDir() {
	if err := os.Chdir(inputDir); err != nil {
		log.Fatal("Error changing directory:", err)
	}
	os.Mkdir("output", 0777)
}

func changeStdOut() {
	resultFilePath := fmt.Sprintf("./output/result_%s.txt", time.Now().Format("0102_150405"))
	if resultFile, err := os.Create(resultFilePath); err != nil {
		fmt.Println(err)
	} else {
		defer resultFile.Close()
		os.Stdout = resultFile
	}
}
