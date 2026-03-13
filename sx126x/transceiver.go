package sx126x

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
)

func (d *Device) EnqueueTx(payload []uint8) error {
	select {
	case d.Queue.Tx <- payload:
		return nil // All ok
	default:
		return fmt.Errorf("TX queue full - packet dropped")
	}
}

func (d *Device) DequeueRx(timeout time.Duration) ([]uint8, error) {
	select {
	case payload := <-d.Queue.Rx:
		return payload, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("RX timeout - no data in RX queue")
	}
}

func (d *Device) WaitForIRQ(timeout time.Duration) bool {
	return d.gpio.dio.WaitForEdge(timeout)
}

func (d *Device) isr() {
	log := d.log.With("func", "Device.isr()", "params", "(-)", "return", "(-)", "lib", "sx1262")
	log.Debug("[ SX126X ] Handle SX126x IRQs")

	irq, err := d.GetIrqStatus()
	if err != nil {
		log.Warn("Could not get SX126x IRQ status; possible hardware/SPI error", "error", err)
		if err := d.ClearIrqStatus(IrqAll); err != nil {
			log.Error("Could not clear SX126x IRQ status: ", "error", err)
		}
		return
	}

	if (irq & (uint16(IrqCrcErr) | uint16(IrqHeaderErr))) > 0 {
		log.Warn("Damaged packet received; dropping it...")
		if err := d.ClearIrqStatus(IrqAll); err != nil {
			log.Error("Could not clear SX126x IRQ status: ", "error", err)
		}
		return
	}

	if (irq & uint16(IrqRxDone)) > 0 {
		log.Debug("[ SX126X ] RX done")
		status, err := d.GetRxBufferStatus()
		if err != nil {
			log.Error("Could not read SX126x RX buffer status; possible hardware/SPI error", "error", err)
			if err := d.ClearIrqStatus(IrqAll); err != nil {
				log.Warn("Could not clear SX126x IRQ status: ", "error", err)
			}
			return
		}

		payload := make([]uint8, status.RXPayloadLength)
		_, err = d.ReadBuffer(status.RXStartPointer, payload)

		if err != nil {
			log.Warn("Could not read SX126x RX buffer; possible hardware/SPI error", "error", err)
		} else if len(payload) > 0 {
			log.Debug("[ SX126X ] SX126x data received")
			select {
			case d.Queue.Rx <- payload:
			default:
				log.Warn("RX channel queue is full")
			}
		}
	}

	if (irq & uint16(IrqTxDone)) > 0 {
		log.Debug("[ SX126X ] TX done")
		if d.gpio.txEn != nil {
			if err := d.gpio.txEn.Out(gpio.Low); err != nil {
				log.Error("Failed to set TxEn pin state to LOW", "error", err)
			}
		}
		if d.gpio.rxEn != nil {
			if err := d.gpio.rxEn.Out(gpio.High); err != nil {
				log.Error("Failed to set RxEn pin state to HIGH", "error", err)
			}
		}

		if err := d.SetRx(int32(RxContinuous)); err != nil {
			log.Error("Could not enable SX126x RX mode", "mode", RxContinuous, "error", err)
		}
	}

	if err := d.ClearIrqStatus(IrqAll); err != nil {
		log.Warn("Could not clear SX126x IRQ status: ", "error", err)
	}
}

func (d *Device) transmit(data []uint8, timeout int32) {
	log := d.log.With("func", "Device.transmit()", "params", "([]uint8, int32)", "return", "(-)", "lib", "sx1262")
	log.Debug("[ SX126X ] Transmit data")

	if d.gpio.txEn != nil {
		if err := d.gpio.txEn.Out(gpio.High); err != nil {
			log.Error("Failed to set TxEn pin state to HIGH", "error", err)
		}
	}
	if d.gpio.rxEn != nil {
		if err := d.gpio.rxEn.Out(gpio.Low); err != nil {
			log.Error("Failed to set RxEn pin state to LOW", "error", err)
		}
	}

	stringToStandby := map[string]StandbyMode{
		"rc":   StandbyRc,
		"xosc": StandbyXosc,
	}

	standby, ok := stringToStandby[d.Config.StandbyMode]
	if !ok {
		standby = StandbyRc
		log.Warn("Unknown standby mode", "mode", d.Config.StandbyMode)
		log.Warn("Limiting standby mode to RC")
	}

	if err := d.SetStandby(standby); err != nil {
		log.Error("Could not set SX126x stanby mode", "mode", standby, "error", err)
		return
	}

	if err := d.SetPacketParams(d.PacketPayLen(len(data))); err != nil {
		log.Error("Could not set SX126x payload length", "payloadLength", len(data), "error", err)
		return
	}

	if _, err := d.WriteBuffer(d.Config.TxBufferAddress, data); err != nil {
		log.Error("Could not write data to Tx buffer")
		return
	}

	if err := d.SetTx(timeout); err != nil {
		log.Error("Failed to transmit data", "error", err)
		return
	}
}
