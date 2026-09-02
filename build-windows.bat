@echo off
setlocal
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go test ./... || exit /b 1
go vet ./... || exit /b 1
if not exist dist mkdir dist
go build -trimpath -buildvcs=false -ldflags="-s -w -buildid= -H=windowsgui" -o dist\BomberRush.exe ./cmd/bomberrush || exit /b 1
echo Gotowe: dist\BomberRush.exe
