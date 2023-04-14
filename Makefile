.PHONY: all clean

BUILD_DIRECTORY=build
BINARY_NAME=ProxiTokScraper

all: linux windows arm mac

clean:
	rm -rf $(BUILD_DIRECTORY)

linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIRECTORY)/$(BINARY_NAME)-linux-amd64

windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIRECTORY)/$(BINARY_NAME)-windows-amd64.exe

arm:
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(BUILD_DIRECTORY)/$(BINARY_NAME)-linux-arm

mac:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIRECTORY)/$(BINARY_NAME)-darwin-amd64
