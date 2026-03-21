package i2c

import "errors"

type Config struct {
	Enable  bool
	Devices map[string]*Device
}

type Device struct {
	Enable bool
	Name   string
	Speed  uint32
	Pins   *Pins
	bus    *machine.I2C
}

type Pins struct {
	SCL machine.Pin
	SDA machine.Pin
}

func (d *Device) Tx(addr uint16, w, r []uint8) error { return d.bus.Tx(addr, w, r) }
func (d *Device) String() string                     { return d.Name }
func (d *Device) IsNil() bool                        { return d == nil }
func (d *Device) Close()                             { d.bus = nil; d.Pins.SCL = machine.NoPin; d.Pins.SDA = machine.NoPin }

func New(device *Device) (*Device, error) {
	if device == nil {
		return nil, errors.New("Bus state improper; device is nil")
	}
	if device.Enable == false {
		return nil, errors.New("Bus disabled in the config")
	}

	var bus *machine.I2C
	var scl, sda machine.Pin

	if device.Pins != nil {
		scl = device.Pins.SCL
		sda = device.Pins.SDA
	}

	switch device.Name {
	case "i2c0":
		bus = machine.I2C0
		if device.Pins == nil {
			scl = machine.I2C0_SCL_PIN
			sda = machine.I2C0_SDA_PIN
		}
	case "i2c1":
		bus = machine.I2C1
		if device.Pins == nil {
			scl = machine.I2C1_SCL_PIN
			sda = machine.I2C1_SDA_PIN
		}
	default:
		return nil, errors.New("Unknown I2C bus")
	}

	err := bus.Configure(machine.I2CConfig{
		Frequency: device.Speed,
		SDA:       sda,
		SCL:       scl,
	})

	if err != nil {
		return nil, err
	}

	device.bus = bus
	return device, nil
}

func Setup(config *Config) (map[string]*Device, func(), error) {
	if config == nil {
		return nil, func() {}, errors.New("Bus state improper; config is nil")
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
