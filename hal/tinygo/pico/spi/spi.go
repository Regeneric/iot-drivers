package spi

import (
	"errors"
	"machine"
)

type Config struct {
	Enable  bool
	Devices map[string]*Device
}

type Device struct {
	Enable   bool
	Name     string
	Speed    uint32
	Mode     uint8
	LSBFirst bool
	Pins     *Pins
	bus      *machine.SPI
}

type Pins struct {
	MISO machine.Pin
	MOSI machine.Pin
	SCK  machine.Pin
}

func (d *Device) Tx(w, r []uint8) error { return d.bus.Tx(w, r) }
func (d *Device) String() string        { return d.Name }
func (d *Device) IsNil() bool           { return d == nil }
func (d *Device) Close() {
	d.bus = nil
	d.Pins = &Pins{
		MISO: machine.NoPin,
		MOSI: machine.NoPin,
		SCK:  machine.NoPin,
	}
}

func New(device *Device) (*Device, error) {
	if device == nil {
		return nil, errors.New("Bus state improper; device is nil")
	}

	var bus *machine.SPI
	var miso, mosi, sck machine.Pin

	if device.Pins != nil {
		miso = device.Pins.MISO
		mosi = device.Pins.MOSI
		sck = device.Pins.SCK
	}

	switch device.Name {
	case "spi0":
		bus = machine.SPI0
		if device.Pins == nil {
			miso = machine.SPI0_SDI_PIN
			mosi = machine.SPI0_SDO_PIN
			sck = machine.SPI0_SCK_PIN
		}
	case "spi1":
		bus = machine.SPI1
		if device.Pins == nil {
			miso = machine.SPI1_SDI_PIN
			mosi = machine.SPI1_SDO_PIN
			sck = machine.SPI1_SCK_PIN
		}
	default:
		return nil, errors.New("Unknown SPI bus")
	}

	err := bus.Configure(
		machine.SPIConfig{
			Frequency: device.Speed,
			SDI:       miso,
			SDO:       mosi,
			SCK:       sck,
			LSBFirst:  device.LSBFirst,
			Mode:      device.Mode,
		})

	if err != nil {
		return nil, errors.New("Could not configure SPI bus")
	}

	device.bus = bus
	return device, nil
}

func Setup(config *Config) (map[string]*Device, func(), error) {
	if config == nil {
		return nil, nil, errors.New("Bus state improper; config is nil")
	}

	if config.Enable == false {
		return nil, func() {}, errors.New("Bus disabled in the config")
	}

	connections := make(map[string]*Device)
	var closers []func()

	cleanup := func() {
		for _, close := range closers {
			close()
		}
	}

	for key, device := range config.Devices {
		if device.Enable == false {
			continue
		}

		bus, err := New(device)
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}
		closers = append(closers, bus.Close)

		connections[key] = bus
	}

	return connections, cleanup, nil
}
