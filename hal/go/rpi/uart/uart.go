package uart

import (
	"errors"

	pconn "periph.io/x/conn/v3"
	pphysic "periph.io/x/conn/v3/physic"
	puart "periph.io/x/conn/v3/uart"
	puartreg "periph.io/x/conn/v3/uart/uartreg"
	phost "periph.io/x/host/v3"
)

type Config struct {
	Enable  bool              `yaml:"enable" env:"UART_ENABLE" env-default:"false"`
	Devices map[string]Device `yaml:"devices"`
}

type Device struct {
	Enable     bool       `yaml:"enable" env:"UART_ENABLE" env-default:"false"`
	Name       string     `yaml:"name" env:"UART_DEVICE" env-default:"0" env-separator:","`
	Speed      uint64     `yaml:"speed" env:"UART_SPEED" env-default:"9600"`
	DataLength int        `yaml:"data_length" env:"UART_DATA_LENGTH" env-default:"8"`
	Parity     string     `yaml:"parity_bit" env:"UART_PARITY_BIT" env-default:"N"`
	StopBit    puart.Stop `yaml:"stop_bit" env:"UART_STOP_BIT" env-default:"1"`
	DataFlow   uint8      `yaml:"data_flow" env:"UART_DATA_FLOW" env-default:"0"`
}

var byteToFlow = map[uint8]puart.Flow{
	0: puart.NoFlow,
	1: puart.XOnXOff,
	2: puart.RTSCTS,
}

var stringToParity = map[string]puart.Parity{
	"N": puart.NoParity,
	"O": puart.Odd,
	"E": puart.Even,
	"M": puart.Mark,
	"S": puart.Space,
}

func New(device string) (puart.PortCloser, error) {
	// RPi drivers
	if _, err := phost.Init(); err != nil {
		return nil, err
	}

	port, err := puartreg.Open(device)
	if err != nil {
		return nil, err
	}

	return port, nil
}

func Setup(config *Config) (map[string]pconn.Conn, func(), error) {
	if config == nil {
		return nil, func() {}, errors.New("UART device state improper; config is nil")
	}
	if config.Enable == false {
		return nil, func() {}, errors.New("UART device disabled in the config")
	}

	connections := make(map[string]pconn.Conn)
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
			return nil, func() {}, err
		}
		closers = append(closers, port.Close)

		flow, ok := byteToFlow[device.DataFlow]
		if !ok {
			flow = puart.NoFlow
			println("Unknown flow option; limiting to NoFlow")
		}

		parity, ok := stringToParity[device.Parity]
		if !ok {
			parity = puart.NoParity
			println("Unknown parity option; limiting to NoParity")
		}

		connection, err := port.Connect(pphysic.Frequency(device.Speed), device.StopBit, parity, flow, device.DataLength)
		if err != nil {
			cleanup()
			return nil, func() {}, errors.Join(err, closeErrors)
		}
		connections[key] = connection
	}

	return connections, cleanup, nil
}

func SetupSingle(config *Device) (pconn.Conn, puart.PortCloser, error) {
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

	flow, ok := byteToFlow[config.DataFlow]
	if !ok {
		flow = puart.NoFlow
		println("Unknown flow option; limiting to NoFlow")
	}

	parity, ok := stringToParity[config.Parity]
	if !ok {
		parity = puart.NoParity
		println("Unknown parity option; limiting to NoParity")
	}

	connection, err := port.Connect(pphysic.Frequency(config.Speed), config.StopBit, parity, flow, config.DataLength)
	if err != nil {
		return nil, nil, err
	}

	return connection, port, nil
}
