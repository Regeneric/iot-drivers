package sx126x

import (
	"context"
	"errors"
	"reflect"
	"time"
)

func New(conn Bus, cfg *Config, opts ...Option) (*Device, error) {
	if cfg == nil {
		return nil, errors.New("[ SX126X ] Modem state improper; cfg is nil")
	}
	if conn == nil || reflect.ValueOf(conn).IsNil() {
		return nil, errors.New("[ SX126X ] SPI bus connection state improper")
	}

	if cfg.Enable == false {
		return nil, errors.New("[ SX126X ] Modem disabled in the config")
	}

	// Empty, explicit init so I will always remember to populate them
	dev := &Device{
		SPI:     conn,
		Config:  cfg,
		Status:  Status{},
		Queue:   Queue{},
		gpioreg: nil,
		gpio:    &pinsDirection{},
		irqChan: make(chan struct{}),
		log:     noplog{},
	}

	for _, opt := range opts {
		opt(dev)
	}

	log := dev.log.With("func", "New()", "params", "(Bus, *Config, ...Option)", "return", "(*Device, error)", "lib", "sx1262")
	log.Info("[ SX126X ] Initializing module")

	// We can define and configure pins in the higher abstraction layer
	if dev.gpioreg != nil {
		log.Info("[ SX1262 ] Pin provider present. Configuring all pins on modem")

		loadPin := func(name string) (PinIO, error) {
			p := dev.gpioreg.ByName(name)
			if p == nil {
				return nil, errors.New("[ SX126X ] Pin not found")
			}
			log.Debug("[ SX126X ] Pin found", "pin", name)
			return p, nil
		}

		var err error
		pins := &pinsDirection{}

		if pins.reset, err = loadPin(cfg.Pins.Reset); err != nil {
			return nil, err
		}
		if pins.busy, err = loadPin(cfg.Pins.Busy); err != nil {
			return nil, err
		}
		if pins.dio, err = loadPin(cfg.Pins.DIO); err != nil {
			return nil, err
		}
		if pins.txEn, err = loadPin(cfg.Pins.TxEn); err != nil {
			return nil, err
		}
		if pins.rxEn, err = loadPin(cfg.Pins.RxEn); err != nil {
			log.Warn(err.Error())
		}
		if cfg.Pins.CS != "" {
			if pins.cs, err = loadPin(cfg.Pins.CS); err != nil {
				return nil, err
			}
		}

		if err := pins.reset.Out(High); err != nil {
			return nil, errors.New("[ SX126X ] Failed to set RESET pin state to HIGH: " + err.Error())
		}
		if err := pins.busy.In(PullNoChange, NoEdge); err != nil {
			return nil, errors.New("[ SX126X ] Failed to set BUSY pin edge detection: " + err.Error())
		}
		if err := pins.dio.In(PullDown, RisingEdge); err != nil {
			return nil, errors.New("[ SX126X ] Failed to set DIO1 pin pull down and edge detection: " + err.Error())
		}
		if pins.txEn != nil {
			if err := pins.txEn.Out(Low); err != nil {
				return nil, errors.New("[ SX126X ] Failed to set TxEn pin state to LOW: " + err.Error())
			}
		}
		if pins.rxEn != nil {
			if err := pins.rxEn.Out(Low); err != nil {
				return nil, errors.New("[ SX126X ] Failed to set RxEn pin state to LOW: " + err.Error())
			}
		}
		if pins.cs != nil {
			if err := pins.cs.Out(High); err != nil {
				return nil, errors.New("[ SX126X ] Failed to set CS pin state to HIGH: " + err.Error())
			}
		}
		dev.gpio = pins
	} else {
		log.Warn("[ SX126X ] No pin provider present. All pins will not be configured during modem setup")
	}

	if cfg.RxQueueSize <= 0 {
		cfg.RxQueueSize = 10
		log.Warn("[ SX126X ] RX queue size cannot be less than 1; resized to 10", "size", cfg.RxQueueSize)
	}

	if cfg.TxQueueSize <= 0 {
		cfg.TxQueueSize = 10
		log.Warn("[ SX126X ] TX queue size cannot be less than 1; resized to 10", "size", cfg.TxQueueSize)
	}

	queue := Queue{
		Rx: make(chan []uint8, cfg.RxQueueSize),
		Tx: make(chan []uint8, cfg.TxQueueSize),
	}
	dev.Queue = queue

	return dev, nil
}

func (d *Device) Close(sleepMode SleepConfig) error {
	log := d.log.With("func", "Device.Close()", "params", "(SleepConfig)", "return", "(error)", "lib", "sx1262")
	log.Info("[ SX126X ] Closing module", "mode", sleepMode)

	var err error = nil
	if err = d.SetSleep(sleepMode); err != nil {
		log.Error("[ SX126X ] Could not set sleep mode", "mode", sleepMode, "error", err)
	}

	if d.gpio.txEn != nil {
		if err = d.gpio.txEn.Out(Low); err != nil {
			log.Error("[ SX126X ] Could not set TxEn pin to LOW", "error", err)
		}
	}

	close(d.Queue.Rx)
	close(d.Queue.Tx)

	return err
}

func (d *Device) Run(ctx context.Context) error {
	log := d.log.With("func", "Device.Run()", "params", "(context.Context)", "return", "(-)", "lib", "sx1262")
	log.Info("[ SX126X ] Modem event loop")

	go func() {
		defer close(d.irqChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if d.WaitForIRQ(1 * time.Second) {
					log.Debug("IRQ!")
					select {
					case d.irqChan <- struct{}{}:
					default:
					}
				}
			}
		}
	}()

	if err := d.SetRx(int32(RxContinuous)); err != nil {
		log.Error("[ SX126X ] Could not enable module RX mode", "mode", RxContinuous, "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-d.irqChan:
			d.isr() // Interrupt Service Routine
		case data := <-d.Queue.Tx:
			d.transmit(data, int32(TxSingle))
		}
	}
}
