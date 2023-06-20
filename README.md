# [속도비교] libx264(sw), h264_nvenc(hw-gpu1), h264_nvenc(hw-gpu2)

## 사용 방법

```sh
    go run main.go -input_dir=./testdata
    # >> start with <입력한 경로>
    # >> done
    # 결과는 ./testdata/output_<test_case>/result_<날짜_시간>.txt 에 기록됨

    # background 로 하려면
    nohup go run main.go &
```

### args

- input_dir: 테스트 파일이 있는 경로, 작업 폴더 (def: ./testdata)

## 비교 항목

```text
    속도
        - 시간측정
    동시성
        - 1개/ 2개/ 4개/ ... 동시에 진행했을 때 걸리는 시간 측정 (setNum)
        - cf) 현재는 3-7개.. (sw 기준)
    bitrate별 차이
        - 현재 bitrate별 파일명 고정되어 있음..
```

## todo

- 파일명 유동성 추가
