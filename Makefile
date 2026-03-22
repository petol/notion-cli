BIN := notion
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: build install clean

build:
	CGO_ENABLED=0 go build -o $(BIN) .

install: build
	install -Dm755 $(BIN) $(INSTALL_DIR)/$(BIN)

clean:
	rm -f $(BIN)
