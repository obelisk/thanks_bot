# Thanks Bot Makefile

.PHONY:
BUILD_DIR=build

thanks_bot:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-linkmode external -extldflags -static" -o $(BUILD_DIR)/thanks_bot -a cmd/*.go

format:
	gofmt -w ./

clean:
	rm -rf $(BUILD_DIR)

