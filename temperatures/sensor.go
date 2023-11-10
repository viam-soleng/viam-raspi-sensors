package temperatures

import (
	"context"
	"sync"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/viam-soleng/viam-raspi-sensors/utils"
)

var Model = resource.NewModel("viam-soleng", "raspi", "temperature")
var PrettyName = "Raspberry Pi Temperature Sensor"
var Description = "A sensor that reports the temperatures of the Raspberry Pi."
var Version = utils.Version

type Config struct {
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
}

func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, *ComponentConfig]{Constructor: NewSensor})
}

func NewSensor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	logger.Infof("Starting %s %s", PrettyName, Version)
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	b := Config{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		mu:         sync.RWMutex{},
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Config) Reconfigure(ctx context.Context, _ resource.Dependencies, conf resource.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Debugf("Reconfiguring %s", PrettyName)

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	socTemp, err := getSoCTemperature()
	if err != nil {
		return nil, err
	}
	pmicTemp, err := getPmicTemperature()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"soc_temp":  socTemp,
		"pmic_temp": pmicTemp,
	}, nil
}

func (c *Config) Close(ctx context.Context) error {
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
