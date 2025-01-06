package models

import (
	"context"
	"errors"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils/rpc"
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
	/*
		Put config attributes here. There should be public/exported fields
		with a `json` parameter at the end of each attribute.

		Example config struct:
			type Config struct {
				Pin   string `json:"pin"`
				Board string `json:"board"`
				MinDeg *float64 `json:"min_angle_deg,omitempty"`
			}

		If your model does not need a config, replace *Config in the init
		function with resource.NoNativeConfig
	*/

	/* Uncomment this if your model does not need to be validated
	   and has no implicit dependecies. */
	// resource.TriviallyValidateConfig
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, error) {
	// Add config validation code here
	return nil, nil
}

type winRegSensorRegistry struct {
	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()

	/* Uncomment this if your model does not need to reconfigure. */
	// resource.TriviallyReconfigurable

	// Uncomment this if the model does not have any goroutines that
	// need to be shut down while closing.
	// resource.TriviallyCloseable

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

func (s *winRegSensorRegistry) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	panic("not implemented")
}

func (s *winRegSensorRegistry) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	panic("not implemented")
}

func (s *winRegSensorRegistry) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
