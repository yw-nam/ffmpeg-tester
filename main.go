package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// const inputDir string = "./testdata"
// const inputFile string = "input.mp4"

var testCases map[string]string = map[string]string{
	"sample_01": "-i %s -c:v libx264 -preset fast -an %s",
	"sample_02": "-i %s -c:v libx264 -preset fast -an %s",

	// "sw_1M_0": "-y -i %s -c:v libx264 -b:v 1M -maxrate:v 1M -minrate:v 1M -bufsize 2M -x264-params nal_hrd=cbr -c:a aac -b:a 96k  -movflags faststart %s",
	// "sw_1M_1": "-y -i %s -c:v libx264 -b:v 1M -maxrate:v 1M -minrate:v 1M -bufsize 2M -x264-params nal_hrd=cbr -c:a aac -b:a 96k  -movflags faststart %s",
	// "sw_2M_0": "-y -i %s -c:v libx264 -b:v 2M -maxrate:v 2M -minrate:v 2M -bufsize 4M -x264-params nal_hrd=cbr -c:a aac -b:a 96k  -movflags faststart %s",
	// "sw_2M_1": "-y -i %s -c:v libx264 -b:v 2M -maxrate:v 2M -minrate:v 2M -bufsize 4M -x264-params nal_hrd=cbr -c:a aac -b:a 96k  -movflags faststart %s",
	// "sw_4M_0": "-y -i %s -c:v libx264 -b:v 4M -maxrate:v 4M -minrate:v 4M -bufsize 8M -x264-params nal_hrd=cbr -c:a aac -b:a 96k  -movflags faststart %s",
	// "sw_4M_1": "-y -i %s -c:v libx264 -b:v 4M -maxrate:v 4M -minrate:v 4M -bufsize 8M -x264-params nal_hrd=cbr -c:a aac -b:a 96k  -movflags faststart %s",

	// "hw_one_gpu_1M_0": "-y -i %s -c:v h264_nvenc -b:v 1M -maxrate:v 1M -bufsize 2M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_one_gpu_1M_1": "-y -i %s -c:v h264_nvenc -b:v 1M -maxrate:v 1M -bufsize 2M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_one_gpu_2M_0": "-y -i %s -c:v h264_nvenc -b:v 2M -maxrate:v 2M -bufsize 4M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_one_gpu_2M_1": "-y -i %s -c:v h264_nvenc -b:v 2M -maxrate:v 2M -bufsize 4M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_one_gpu_4M_0": "-y -i %s -c:v h264_nvenc -b:v 4M -maxrate:v 4M -bufsize 8M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_one_gpu_4M_1": "-y -i %s -c:v h264_nvenc -b:v 4M -maxrate:v 4M -bufsize 8M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",

	// "hw_two_gpu_1M_0": "-y -i %s -c:v h264_nvenc -b:v 1M -maxrate:v 1M -bufsize 2M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_two_gpu_1M_1": "-y -i %s -c:v h264_nvenc -b:v 1M -maxrate:v 1M -bufsize 2M -preset p6 -rc cbr -gpu 1 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_two_gpu_2M_0": "-y -i %s -c:v h264_nvenc -b:v 2M -maxrate:v 2M -bufsize 4M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_two_gpu_2M_1": "-y -i %s -c:v h264_nvenc -b:v 2M -maxrate:v 2M -bufsize 4M -preset p6 -rc cbr -gpu 1 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_two_gpu_4M_0": "-y -i %s -c:v h264_nvenc -b:v 4M -maxrate:v 4M -bufsize 8M -preset p6 -rc cbr -gpu 0 -c:a aac -b:a 96k  -movflags faststart  %s",
	// "hw_two_gpu_4M_1": "-y -i %s -c:v h264_nvenc -b:v 4M -maxrate:v 4M -bufsize 8M -preset p6 -rc cbr -gpu 1 -c:a aac -b:a 96k  -movflags faststart  %s",
}

func main() {
	inputDir := flag.String("input_dir", "./testdata", "-input_dir=./testdata")
	inputFile := flag.String("input_file", "input.mp4", "-input_file=input.mp4")
	flag.Parse()

	fmt.Printf(">> start with %s/%s\n", *inputDir, *inputFile)
	originalStdOut := os.Stdout
	if err := os.Chdir(*inputDir); err != nil {
		log.Fatal("Error changing directory:", err)
	}
	os.Mkdir("output", 0777)

	resultFilePath := fmt.Sprintf("./output/result_%s.txt", time.Now().Format("0102_150405"))
	if resultFile, err := os.Create(resultFilePath); err != nil {
		fmt.Println(err)
	} else {
		defer resultFile.Close()
		os.Stdout = resultFile
	}

	var wg sync.WaitGroup
	for caseName, optFmt := range testCases {
		wg.Add(1)
		outputFile := fmt.Sprintf("./output/output_%s.mp4", caseName)
		option := fmt.Sprintf(optFmt, *inputFile, outputFile)
		go encoding(option, outputFile, &wg)
	}
	wg.Wait()
	fmt.Println("file end")
	os.Stdout = originalStdOut
	fmt.Println(">> done")
}

func encoding(option string, outputFile string, wg *sync.WaitGroup) {
	defer wg.Done()
	os.Remove(outputFile)

	start := time.Now()
	cmd := exec.Command("ffmpeg", strings.Split(option, " ")...)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("done[%s]: %v\n", outputFile, time.Since(start))
}
