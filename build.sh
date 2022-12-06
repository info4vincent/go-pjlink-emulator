git_head=$(git rev-parse HEAD)
version="-X main.Version=$git_head"
echo $version
go build -ldflags "$version" PJLinkEmulator.go