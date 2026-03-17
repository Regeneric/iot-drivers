package i2c

import (
	"errors"

	pi2c "periph.io/x/conn/v3/i2c"
	pi2creg "periph.io/x/conn/v3/i2c/i2creg"
	phost "periph.io/x/host/v3"
)

type Config struct {
	Enable  bool
	Devices map[string]Device
}

type Device struct {
	Enable bool
	Name   string
}

func New(device string) (pi2c.BusCloser, error) {
	// RPi drivers
	if _, err := phost.Init(); err != nil {
		return nil, err
	}

	// /dev/i2c-0 ; /dev/i2c-1 etc.
	bus, err := pi2creg.Open(device)
	if err != nil {
		return nil, err
	}

	return bus, nil
}

func Setup(config *Config) (map[string]pi2c.BusCloser, func(), error) {
	if config == nil {
		return nil, func() {}, errors.New("Bus state improper; config is nil")
	}

	if config.Enable == false {
		return nil, func() {}, errors.New("Bus disabled in the config")
	}

	connections := make(map[string]pi2c.BusCloser)
	var closers []func() error
	var closeErrors error

	cleanup := func() {
		for _, close := range closers {
			if err := close(); err != nil {
				closeErrors = errors.Join(closeErrors, err)
			}
		}
	}

	for key, device := range config.Devices {
		if device.Enable == false {
			continue
		}

		bus, err := New(device.Name)
		if err != nil {
			cleanup()
			return nil, func() {}, errors.Join(err, closeErrors)
		}
		closers = append(closers, bus.Close)

		connections[key] = bus
	}

	return connections, cleanup, nil
}

func SetupSingle(config *Device) (pi2c.BusCloser, error) {
	if config == nil {
		return nil, errors.New("Device state improper; config is nil")
	}

	if config.Enable == false {
		return nil, errors.New("Device disabled in the config")
	}

	bus, err := New(config.Name)
	if err != nil {
		return nil, err
	}

	return bus, nil
}
