# [속도비교] libx264(sw), h264_nvenc(hw-gpu1), h264_nvenc(hw-gpu2)

## 사용 방법

```sh
    go run main.go -input_dir=./testdata -input_file=input.mp4
    # >> start with <입력한 경로>
    # >> done
    # 결과는 ./testdat/output/result_<날짜_시간>.txt 에 기록됨
```

### args

- input_dir: 테스트 파일이 있는 경로, 작업 폴더 (def: ./testdata)
- input_file: 테스트 대상 파일 (def: input.mp4)

## 비교 항목

```text
    속도
        - 시간측정
        - sh: 맨 앞에 time 쓰면 실행시간이 출력된다.
    동시성
        - 1개/ 3개/ 5개 동시에 진행했을 때 걸리는 시간 측정
        - cf) 현재는 3-7개.. (sw 기준)
```
