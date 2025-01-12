//go:build windows

package main

import (
	"context"
	"win-reg-sensor/models"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"
)

func main() {
	utils.ContextualMain(mainWithArgs, module.NewLoggerFromArgs("win-reg-sensor"))
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) error {
	winRegSensor, err := module.NewModuleFromArgs(ctx)
	if err != nil {
		return err
	}

	if err = winRegSensor.AddModelFromRegistry(ctx, sensor.API, models.Registry); err != nil {
		return err
	}

	err = winRegSensor.Start(ctx)
	defer winRegSensor.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
