BIN=imgdedup

.PHONY: all
all: clean test build install

.PHONY: test
test:
	go test ./...

.PHONY: install
install:
	go install ./cmd/imgdedup

.PHONY: clean
clean:
	-rm -rf release dist
	mkdir release dist

release/darwin_amd64:
	env GOOS=darwin GOARCH=amd64 go build -o release/darwin_amd64/$(BIN) ./cmd/imgdedup

release/darwin_arm64:
	env GOOS=darwin GOARCH=arm64 go build -o release/darwin_arm64/$(BIN) ./cmd/imgdedup

release/darwin_universal: release/darwin_amd64 release/darwin_arm64
	mkdir release/darwin_universal
	lipo -create -output release/darwin_universal/$(BIN) release/darwin_amd64/$(BIN) release/darwin_arm64/$(BIN)

release/linux_amd64:
	env GOOS=linux GOARCH=amd64 go build -o release/linux_amd64/$(BIN) ./cmd/imgdedup

release/freebsd_amd64:
	env GOOS=linux GOARCH=amd64 go build -o release/freebsd_amd64/$(BIN) ./cmd/imgdedup

release/windows_amd64:
	env GOOS=windows GOARCH=amd64 go build -o release/windows_amd64/$(BIN).exe ./cmd/imgdedup

.PHONY: build
build: release/darwin_universal release/linux_amd64 release/freebsd_amd64 release/windows_amd64

.PHONY: release
release: clean build
	zip -9 -j 'dist/$(BIN).darwin_universal.zip'  release/darwin_universal/$(BIN)
	zip -9 -j 'dist/$(BIN).linux_amd64.zip'       release/linux_amd64/$(BIN)
	zip -9 -j 'dist/$(BIN).freebsd_amd64.zip'     release/freebsd_amd64/$(BIN)
	zip -9 -j 'dist/$(BIN).windows_amd64.exe.zip' release/windows_amd64/$(BIN).exe
