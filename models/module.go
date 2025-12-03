//go:build windows

package models

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	errw "github.com/pkg/errors"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils/rpc"

	"golang.org/x/sys/windows/registry"
)

var (
	Registry         = resource.NewModel("viam", "win-reg-sensor", "registry")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterComponent(sensor.API, Registry,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: newWinRegSensorRegistry,
		},
	)
}

type Config struct {
	Keys     []string `json:"keys"`
	Programs []string `json:"programs"`
	resource.TriviallyValidateConfig
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	return nil, nil, nil
}

type winRegSensorRegistry struct {
	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()

	resource.TriviallyReconfigurable

	resource.TriviallyCloseable
}

func newWinRegSensorRegistry(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &winRegSensorRegistry{
		name:       rawConf.ResourceName(),
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *winRegSensorRegistry) Name() resource.Name {
	return s.name
}

func (s *winRegSensorRegistry) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	myConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return err
	}
	s.cfg = myConf
	return nil
}

func (s *winRegSensorRegistry) NewClientFromConn(ctx context.Context, conn rpc.ClientConn, remoteName string, name resource.Name, logger logging.Logger) (sensor.Sensor, error) {
	return nil, errUnimplemented
}

func (s *winRegSensorRegistry) Readings(ctx context.Context, extra map[string]any) (map[string]any, error) {
	ret := make(map[string]any)
	s.logger.Debugf("reading %d programs", len(s.cfg.Programs))
	for _, programName := range s.cfg.Programs {
		version, err := getWindowsProgramVersion(programName)
		if err != nil {
			// Not installed or not found
			s.logger.Warnf("%v", err)
			version = "Not installed"
		}

		s.logger.Infof("%s version details: %s", programName, version)
		ret[programName] = version
	}

	s.logger.Debugf("reading %d keys", len(s.cfg.Keys))
	for _, fullKey := range s.cfg.Keys {
		subMap := make(map[string]any)
		// note: in colon/subKey mode, we set a single value instead of the submap.
		// we do this because triggers can't access nested maps.
		key, subKey, hasColon := strings.Cut(fullKey, ":")
		if !hasColon {
			ret[fullKey] = subMap
		}
		err := func() error {
			s.logger.Debugf("opening key %s", key)
			k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.QUERY_VALUE)
			if err != nil {
				return errw.Wrap(err, key)
			}
			defer k.Close()
			names, err := k.ReadValueNames(0)
			if err != nil {
				return errw.Wrap(err, key)
			}
			for _, name := range names {
				val, _, err := k.GetStringValue(name)
				if err == registry.ErrUnexpectedType {
					// todo: handle some non-string types
					val = "non_string"
				} else if err != nil {
					return errw.Wrap(err, key)
				}
				subMap[name] = val
			}
			if hasColon {
				ret[fullKey] = subMap[subKey]
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (s *winRegSensorRegistry) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	s.logger.Infof("DoCommand not implemented")
	return nil, errUnimplemented
}

func (s *winRegSensorRegistry) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}

func getWindowsProgramVersion(programName string) (string, error) {
	var subkey string

	// Attempt to find the program's uninstall information in the registry.
	if programName != "" {
		subkey = `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`

		k, err := registry.OpenKey(registry.LOCAL_MACHINE, subkey, registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
		if err != nil {
			return "", err
		}
		defer k.Close()

		subkeys, err := k.ReadSubKeyNames(-1)
		if err != nil {
			return "", err
		}

		for _, name := range subkeys {
			appKeyPath := filepath.Join(subkey, name)
			appKey, err := registry.OpenKey(registry.LOCAL_MACHINE, appKeyPath, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			defer appKey.Close()

			displayName, _, err := appKey.GetStringValue("DisplayName")
			if err != nil {
				continue
			}

			if strings.Contains(displayName, programName) {
				version, _, err := appKey.GetStringValue("DisplayVersion")
				if err != nil {
					return "", err
				}
				return version, nil
			}
		}
	}
	return "", fmt.Errorf("program '%s' not found or version information unavailable", programName)
}
