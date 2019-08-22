default: test install

test:
	go test ./...

install:
	go install

.PHONY: clean
clean:
	-rm -rf release
	mkdir release

.PHONY: release
release: clean darwin64 linux64 windows64
	cd release/darwin64 && zip -9 ../darwin64.zip imgdedup
	cd release/linux64 && zip -9 ../linux64.zip imgdedup
	cd release/windows64 && zip -9 ../windows64.zip imgdedup.exe

darwin64:
	env GOOS=darwin GOARCH=amd64 go build -o release/darwin64/imgdedup .

linux64:
	env GOOS=linux GOARCH=amd64 go build -o release/linux64/imgdedup .

windows64:
	env GOOS=windows GOARCH=amd64 go build -o release/windows64/imgdedup.exe .
