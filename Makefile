win-reg-sensor.exe: Makefile *.go models/*.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags=no_cgo -ldflags="-s -w" .

module.tar.gz: win-reg-sensor.exe meta.json
	echo VIAM_TARGET_OS $(VIAM_TARGET_OS)
	rm -f $@
	tar czf $@ $^
