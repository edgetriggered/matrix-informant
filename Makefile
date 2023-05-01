all: matrix-informant

clean:
	go clean
	rm -f matrix-informant

format:
	gofmt -s -w .

DATE=$(shell TZ=GMT date --rfc-3339="seconds")
HASH=$(shell printf "r%s.%s" "$(shell git rev-list --count HEAD)" "$(shell git describe --always --abbrev=8 --dirty --exclude '*')")

matrix-informant: $(shell find pkg -name \*.go) cmd/matrix-informant/*
	go build \
		-ldflags "-s -w -X 'main.buildDate=$(DATE)' -X 'main.buildHash=$(HASH)'" \
		-o matrix-informant cmd/matrix-informant/main.go