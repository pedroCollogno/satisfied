EXE=satisfied.exe
OUTDIR=./bin

run:
	go run .

icon: assets/icon.ico resource.rc
	windres -i resource.rc -o resource.syso

build: icon
	go build -o $(OUTDIR)/debug/$(EXE)

release: icon
	CGO_CPPFLAGS="-O3 -DNDEBUG" go build -ldflags="-s -w -H=windowsgui" -o $(OUTDIR)/release/$(EXE)

clean:
	rm -rf ./bin
	go clean -cache

tidy:
	go mod tidy
