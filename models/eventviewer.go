//go:build windows

package models

import (
	"context"
	"fmt"
	"syscall"

	"github.com/google/winops/winlog"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils/rpc"
	"golang.org/x/sys/windows"
)

var (
	EventViewer = resource.NewModel("viam", "win-reg-sensor", "event-viewer")
)

func init() {
	resource.RegisterComponent(sensor.API, EventViewer,
		resource.Registration[sensor.Sensor, *EVConfig]{
			Constructor: newWinRegSensorEventViewer,
		},
	)
}

type EVConfig struct {
	resource.TriviallyValidateConfig
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `EVConfig` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *EVConfig) Validate(path string) ([]string, error) {
	return nil, nil
}

type winRegSensorEventViewer struct {
	name resource.Name

	logger logging.Logger
	cfg    *EVConfig

	cancelCtx  context.Context
	cancelFunc func()

	resource.TriviallyReconfigurable

	resource.TriviallyCloseable
}

func newWinRegSensorEventViewer(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*EVConfig](rawConf)
	if err != nil {
		return nil, err
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &winRegSensorEventViewer{
		name:       rawConf.ResourceName(),
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *winRegSensorEventViewer) Name() resource.Name {
	return s.name
}

// func (s *winRegSensorEventViewer) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
// 	myConf, err := resource.NativeConfig[*Config](conf)
// 	if err != nil {
// 		return err
// 	}
// 	s.cfg = myConf
// 	return nil
// }

func (s *winRegSensorEventViewer) NewClientFromConn(ctx context.Context, conn rpc.ClientConn, remoteName string, name resource.Name, logger logging.Logger) (sensor.Sensor, error) {
	panic("not implemented")
}

func (s *winRegSensorEventViewer) Readings(ctx context.Context, extra map[string]any) (map[string]any, error) {
	pubcache := make(map[string]windows.Handle)
	defer func() {
		for _, h := range pubcache {
			winlog.Close(h)
		}
	}()
	config, err := winlog.DefaultSubscribeConfig()
	if err != nil {
		return nil, err
	}
	sub, err := winlog.Subscribe(config)
	if err != nil {
		return nil, err
	}
	defer winlog.Close(sub)
	status, err := windows.WaitForSingleObject(config.SignalEvent, 1000)
	if err != nil {
		return nil, err
	}
	if status == syscall.WAIT_OBJECT_0 {
		rendered, err := winlog.GetRenderedEvents(config, pubcache, sub, 100, 1033)
		if err != nil {
			return nil, err
		}
		s.logger.Info("about to return %+v\n", rendered)
		var e0 string
		if len(rendered) > 0 {
			e0 = rendered[0]
		}
		return map[string]any{
			"nevents": len(rendered),
			"first":   e0,
		}, nil
	} else {
		return nil, fmt.Errorf("unexpected status %d", status)
	}
}

func (s *winRegSensorEventViewer) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	return nil, errUnimplemented
}

func (s *winRegSensorEventViewer) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
