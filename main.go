package main

import (
	"win-reg-sensor/models"
	"context"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"
	"go.viam.com/rdk/components/sensor"

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
