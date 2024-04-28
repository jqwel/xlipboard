mode=release
version=1.0.0

cd "$( dirname "${BASH_SOURCE[0]}" )"
ROOT=`pwd`
RELEASE_DIR="$ROOT/release"
mkdir -p $RELEASE_DIR

#CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui -X 'main.mode=$mode' -X 'main.version=$version'" -o $RELEASE_DIR/xlipboard_windows.exe ./src/cmd/main.go
#CGO_ENABLED=1 GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w -X 'main.mode=$mode' -X 'main.version=$version'" -o $RELEASE_DIR/xlipboard_ubuntu ./src/cmd/main.go
#CGO_ENABLED=1 GOOS=darwin               go build -ldflags="-s -w -X 'main.mode=$mode' -X 'main.version=$version'" -o $RELEASE_DIR/xlipboard_macos ./src/cmd/main.go
