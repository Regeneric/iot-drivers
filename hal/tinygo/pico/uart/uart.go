package uart

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
	BaudRate uint32
	DataBits uint8
	StopBits uint8
	Parity   machine.UARTParity
	Pins     *Pins
	bus      *machine.UART
}

type Pins struct {
	RX  machine.Pin
	TX  machine.Pin
	RTS machine.Pin
	CTS machine.Pin
}

func (d *Device) Write(data []uint8) (int, error) { return d.bus.Write(data) }
func (d *Device) WriteByte(data uint8) error      { return d.bus.WriteByte(data) }
func (d *Device) Read(data []uint8) (int, error)  { return d.bus.Read(data) }
func (d *Device) ReadByte() (uint8, error)        { return d.bus.ReadByte() }
func (d *Device) String() string                  { return d.Name }
func (d *Device) IsNil() bool                     { return d == nil }
func (d *Device) Close() {
	d.bus = nil
	d.Pins = &Pins{
		RX:  machine.NoPin,
		TX:  machine.NoPin,
		RTS: machine.NoPin,
		CTS: machine.NoPin,
	}
}

func New(device *Device) (*Device, error) {
	if device == nil {
		return nil, errors.New("UART device state improper; device is nil")
	}
	if device.Enable == false {
		return nil, errors.New("Bus disabled in the config")
	}

	var bus *machine.UART // I know it's not a bus
	var rx, tx machine.Pin

	if device.Pins != nil {
		rx = device.Pins.RX
		tx = device.Pins.TX
	}

	switch device.Name {
	case "uart0":
		bus = machine.UART0
		if device.Pins == nil {
			rx = machine.UART0_RX_PIN
			tx = machine.UART0_TX_PIN
		}
	case "uart1":
		bus = machine.UART1
		if device.Pins == nil {
			rx = machine.UART1_RX_PIN
			tx = machine.UART1_TX_PIN
		}
	default:
		return nil, errors.New("Unknown UART device")
	}

	err := bus.Configure(
		machine.UARTConfig{
			BaudRate: device.BaudRate,
			TX:       tx,
			RX:       rx,
			RTS:      device.Pins.RTS,
			CTS:      device.Pins.CTS,
		},
	)
	if err != nil {
		return nil, err
	}

	err = bus.SetFormat(device.DataBits, device.StopBits, device.Parity)
	if err != nil {
		return nil, err
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
