.PHONY: build

build:
	@if [ ! -d "build" ]; then mkdir "build"; fi

	@if [ ! -d "build/windows" ]; then mkdir "build/windows"; fi
	@if [ ! -d "build/windows/amd64" ]; then mkdir "build/windows/amd64"; fi
	GOOS=windows GOARCH=amd64 go build -o build/windows/amd64/multisum.exe .

	@if [ ! -d "build/linux" ]; then mkdir "build/linux"; fi
	@if [ ! -d "build/linux/amd64" ]; then mkdir "build/linux/amd64"; fi
	GOOS=linux GOARCH=amd64 go build -o build/linux/amd64/multisum .

test:
	@if [ -d "testdata/tmp" ]; then rm -r "testdata/tmp"; fi; mkdir "testdata/tmp"
	go build .
	./multisum --md5 --sha512 --target ./testdata/tmp ~/go
