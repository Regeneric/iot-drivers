package onewire

import (
	"errors"

	ponewire "periph.io/x/conn/v3/onewire"
	ponewirereg "periph.io/x/conn/v3/onewire/onewirereg"
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

func New(device string) (ponewire.BusCloser, error) {
	// RPi drivers
	if _, err := phost.Init(); err != nil {
		return nil, err
	}

	bus, err := ponewirereg.Open(device)
	if err != nil {
		return nil, err
	}

	return bus, nil
}

func Setup(config *Config) (map[string]ponewire.BusCloser, func(), error) {
	if config == nil {
		return nil, func() {}, errors.New("Bus state improper; config is nil")
	}

	if config.Enable == false {
		return nil, func() {}, errors.New("Bus disabled in the config")
	}

	connections := make(map[string]ponewire.BusCloser)
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

func SetupSingle(config *Device) (ponewire.BusCloser, error) {
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
