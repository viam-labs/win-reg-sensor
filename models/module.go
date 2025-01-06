package models

import (
	"context"
	"errors"

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
	Keys []string `json:"keys"`
	resource.TriviallyValidateConfig
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, error) {
	return nil, nil
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
	// Put reconfigure code here
	return errUnimplemented
}

func (s *winRegSensorRegistry) NewClientFromConn(ctx context.Context, conn rpc.ClientConn, remoteName string, name resource.Name, logger logging.Logger) (sensor.Sensor, error) {
	panic("not implemented")
}

func (s *winRegSensorRegistry) Readings(ctx context.Context, extra map[string]any) (map[string]any, error) {
	ret := make(map[string]any)
	s.logger.Infof("reading %d keys", len(s.cfg.Keys))
	for _, key := range s.cfg.Keys {
		subMap := make(map[string]any)
		ret[key] = subMap
		err := func() error {
			s.logger.Infof("opening key %s", key)
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
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (s *winRegSensorRegistry) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	panic("not implemented")
}

func (s *winRegSensorRegistry) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
