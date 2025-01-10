win-reg-sensor.exe: *.go models/*.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" .

module.tar.gz: win-reg-sensor.exe meta.json
	rm -f $@
	tar czf $@ $^
