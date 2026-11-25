# windows registry sensor

This is a viam module that reads windows registry keys using the golang [registry package](https://pkg.go.dev/golang.org/x/sys/windows/registry).

## Build

Run `make win-reg-sensor.exe` to cross-compile for Windows amd64, or run Go directly:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags=no_cgo ./...
```

The RDK's `no_cgo` tag excludes its microphone driver, which this registry sensor does not use.

## Example config

```json
{
  "keys": [
    "SOFTWARE\\Viam",
    "SOFTWARE\\Viam:version"
  ],
  "programs": [
    "Google Chrome",
    "Microsoft Edge"
  ]
}
```

This produces output:

```json
Google Chrome    "142.0.7444.176"
Microsoft Edge   "142.0.3595.94"
{
  "SOFTWARE\\Viam": {
    "version": "123",
    "": "hello"
  },
  "SOFTWARE\\Viam:version": "123"
}
```

The `SOFTWARE\\Viam:version` is a special form that reads the `version` value inside the `Viam` key.

## Limitations

This currently only supports string keys.
