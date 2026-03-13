package sgp30

import (
	"errors"
	"os"
	"reflect"
)

func New(i2c Bus, cfg *Config, opts ...Option) (*Device, error) {
	if cfg == nil {
		return nil, errors.New("[ SGP30 ] Sensor state improper; cfg is nil")
	}
	if i2c == nil || reflect.ValueOf(i2c).IsNil() {
		return nil, errors.New("[ SGP30 ] Sensor state improper; i2c is nil")
	}

	dev := &Device{
		I2C:    i2c,
		Config: cfg,
		log:    noplog{}, // Default empty logger
	}

	for _, opt := range opts {
		opt(dev)
	}

	log := dev.log.With("func", "New()", "params", "(Bus, *cfg, ...Option)", "return", "(*Device, error)", "lib", "sgp30")
	log.Info("[ SGP30 ] Single sensor constructor", "name", cfg.Name)

	return dev, nil
}

// TODO full initilization path with optional settings
func Setup(buses map[string]Bus, cfg *Group, logger Logger) (map[string]*Device, func(), error) {
	if logger == nil {
		logger = noplog{}
	}

	log := logger.With("func", "Setup()", "params", "(*cfg, map[string]Bus, Logger)", "return", "(map[string]*Device, func(), error)", "lib", "sgp30")
	log.Info("[ SGP30 ] All sensors setup")

	if cfg.Enable == false {
		return nil, func() {}, errors.New("[ SGP30 ] All sensors disabled in the config file")
	}

	sensors := make(map[string]*Device)
	var closers []func() error

	cleanup := func() {
		for i, c := range closers {
			log.Debug("[ SGP30 ] Closing SGP30 sensor and saving baseline values...", "sensor", i)
			_ = c()
		}
	}

	for key, dev := range cfg.Devices {
		if dev.Enable == false {
			log.Debug("[ SGP30 ] Sensor disabled in the config file;", "name [", dev.Name, "] bus [", dev.BusName, "] address [", Hex8(dev.Address), "]")
			continue
		}

		bus, ok := buses[dev.BusName]
		if !ok {
			cleanup()
			return nil, func() {}, errors.New("[ SGP30 ] I2C bus '" + dev.BusName + "' not found for SGP30 sensor '" + dev.Name + "' with address [ " + Hex8(dev.Address) + " ]")
		}

		sensor, err := New(bus, &dev, WithLogger(logger))
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}

		if err := sensor.IaqInit(); err != nil {
			cleanup()
			return nil, func() {}, errors.New("[ SGP30 ] SGP30 sensor '" + dev.Name + "' on bus '" + dev.BusName + "' with address [ " + Hex8(dev.Address) + " ] unresponsive: " + err.Error())
		}

		// temporary
		baseline := []uint8{0xA1, 0xAF, 0x58, 0xA5, 0x2A, 0x54}
		if err := sensor.SetIaqBaseline(baseline); err != nil {
			cleanup()
			return nil, func() {}, errors.New("[ SGP30 ] SGP30 sensor '" + dev.Name + "' on bus '" + dev.BusName + "' with address [ " + Hex8(dev.Address) + " ] unresponsive: " + err.Error())
		}

		closers = append(closers, func() error {
			baseline := make([]uint8, 6)
			if err := sensor.GetIaqBaseline(baseline); err != nil {
				return err
			}

			filename := "sgp30_baseline" + stringSanitize(key) + ".bin"
			if err := os.WriteFile(filename, baseline, 0644); err != nil {
				return err
			}

			return nil
		})

		sensors[key] = sensor
		log.Debug("[ SGP30 ] Sensor configured and initilized;", "name [", dev.Name, "] bus [", dev.BusName, "] address [", Hex8(dev.Address), "]")
	}

	return sensors, cleanup, nil
}
