CMD = ./cmd/anasm

BIN     = ./bin
OUT     = $(BIN)/app
INSTALL = /usr/bin/anasm

GO = go

$(OUT):
	$(GO) build -o $(OUT) $(CMD)

$(BIN):
	mkdir -p $(BIN)

run:
	$(GO) run $(CMD)

install: $(OUT)
	cp $(OUT) $(INSTALL)

clean:
	rm -r $(BIN)/*

all:
	@echo compile, run, install, clean
