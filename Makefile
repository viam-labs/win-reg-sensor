win-reg-sensor.exe: *.go models/*.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" .
