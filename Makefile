EXE=satisfied.exe
OUTDIR=./bin

run:
	go run .

build:
	go build -o $(OUTDIR)/debug/$(EXE)

release:
	CGO_CPPFLAGS="-O3 -DNDEBUG" go build -ldflags="-s -w -H=windowsgui" -o $(OUTDIR)/release/$(EXE)

clean:
	rm -rf ./bin
	go clean -cache

tidy:
	go mod tidy
