package spi

import (
	"errors"

	pphysic "periph.io/x/conn/v3/physic"
	pspi "periph.io/x/conn/v3/spi"
	pspireg "periph.io/x/conn/v3/spi/spireg"
	phost "periph.io/x/host/v3"
)

type Config struct {
	Enable  bool
	Devices map[string]Device
}

type Device struct {
	Enable      bool
	Name        string
	Speed       uint64
	Mode        pspi.Mode
	BitsPerWord int
}

func New(device string) (pspi.PortCloser, error) {
	// RPi drivers
	if _, err := phost.Init(); err != nil {
		return nil, err
	}

	// SPI0.0 ; SPI0.1 etc.
	bus, err := pspireg.Open(device)
	if err != nil {
		return nil, err
	}

	return bus, nil
}

func Setup(config *Config) (map[string]pspi.Conn, func(), error) {
	if config == nil {
		return nil, func() {}, errors.New("Bus state improper; config is nil")
	}

	if config.Enable == false {
		return nil, func() {}, errors.New("Bus disabled in the config")
	}

	connections := make(map[string]pspi.Conn)
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

		port, err := New(device.Name)
		if err != nil {
			cleanup()
			return nil, func() {}, errors.Join(err, closeErrors)
		}
		closers = append(closers, port.Close)

		connection, err := port.Connect(pphysic.Frequency(device.Speed*uint64(pphysic.Hertz)), device.Mode, device.BitsPerWord)
		if err != nil {
			cleanup()
			return nil, func() {}, errors.Join(err, closeErrors)
		}
		connections[key] = connection
	}

	return connections, cleanup, nil
}

func SetupSingle(config *Device) (pspi.Conn, pspi.PortCloser, error) {
	if config == nil {
		return nil, nil, errors.New("Device state improper; config is nil")
	}

	if config.Enable == false {
		return nil, nil, errors.New("Device disabled in the config")
	}

	port, err := New(config.Name)
	if err != nil {
		return nil, nil, err
	}

	connection, err := port.Connect(pphysic.Frequency(config.Speed*uint64(pphysic.Hertz)), config.Mode, config.BitsPerWord)
	if err != nil {
		return nil, nil, err
	}

	return connection, port, nil
}
