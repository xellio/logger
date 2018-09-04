GO      = go
TARGET  = logger

BINDIR = ./bin/

GLIDE_VERSION := $(shell glide --version 2>/dev/null)
UPX := $(shell upx --version 2>/dev/null)

all: $(TARGET)

$(TARGET): build
ifdef UPX
	upx --brute $(BINDIR)$@
endif

build: vendor clean $(BINDIR)
	$(GO) build -ldflags="-s -w" -o $(BINDIR)$(TARGET) ./cmd/cli/main.go

vendor:
ifdef GLIDE_VERSION
	glide install
else
	go get .
endif

clean:
	rm -f $(BINDIR)*

$(BINDIR):
	mkdir -p $(BINDIR)