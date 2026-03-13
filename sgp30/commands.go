package sgp30

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func (d *Device) IaqInit() error {
	log := d.log.With("func", "Device.IaqInit()", "params", "(-)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Sensor IAQ init", "sensor", d.Config.Name)

	command := []uint8{uint8(CmdIaqInitMSB), uint8(CmdIaqInitLSB)}
	if err := d.I2C.Tx(uint16(d.Config.Address), command, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send init command to SGP30 sensor '" + d.Config.Name + "': " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(10 * time.Millisecond)

	log.Info("[ SGP30 ] Init command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdIaqInit)
	return nil
}

func (d *Device) MeasureIaq(data []uint8) error {
	log := d.log.With("func", "Device.MeasureIaq()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Measure eCO2 and TVOC values", "sensor", d.Config.Name)

	dataFrameLength := 6
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	command := []uint8{uint8(CmdMeasureIaqMSB), uint8(CmdMeasureIaqLSB)}
	if err := d.I2C.Tx(uint16(d.Config.Address), command, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send measure command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(12 * time.Millisecond)

	if err := d.I2C.Tx(uint16(d.Config.Address), nil, data); err != nil {
		return errors.New("[ SGP30 ] Could not read measured values from '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Function validates CRC after it was received to `data` buffer (duh)
	// So it is possible to ignore returned error and look directly at raw data
	if err := validateCRC(data[0:3]); err != nil {
		return errors.New("[ SGP30 ] eCO2 CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}
	if err := validateCRC(data[3:6]); err != nil {
		return errors.New("[ SGP30 ] TVOC CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	log.Info("[ SGP30 ] Measure command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdMeasureIaq)
	return nil
}

// 6 bytes: (eCO2_MSB, eCO2_LSB, eCO2_CRC, TVOC_MSB, TVOC_LSB, TVOC_CRC)
func (d *Device) GetIaqBaseline(data []uint8) error {
	log := d.log.With("func", "Device.GetIaqBaseline()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Get IAQ baseline calibration value", "sensor", d.Config.Name)

	dataFrameLength := 6
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	command := []uint8{uint8(CmdGetIaqBaselineMSB), uint8(CmdGetIaqBaselineLSB)}
	if err := d.I2C.Tx(uint16(d.Config.Address), command, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send get IAQ baseline command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(10 * time.Millisecond)

	if err := d.I2C.Tx(uint16(d.Config.Address), nil, data); err != nil {
		return errors.New("[ SGP30 ] Could not get IAQ baseline values from '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Function validates CRC after it was received to `data` buffer (duh)
	// So it is possible to ignore returned error and look directly at raw data
	if err := validateCRC(data[0:3]); err != nil {
		return errors.New("[ SGP30 ] eCO2 CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}
	if err := validateCRC(data[3:6]); err != nil {
		return errors.New("[ SGP30 ] TVOC CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	log.Info("[ SGP30 ] Get IAQ baseline command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdGetIaqBaseline)
	return nil
}

// 6 bytes: (eCO2_MSB, eCO2_LSB, eCO2_CRC, TVOC_MSB, TVOC_LSB, TVOC_CRC)
func (d *Device) SetIaqBaseline(data []uint8) error {
	log := d.log.With("func", "Device.SetIaqBaseline()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Set IAQ baseline calibration value", "sensor", d.Config.Name)

	dataFrameLength := 6
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	if err := validateCRC(data[0:3]); err != nil {
		return errors.New("[ SGP30 ] eCO2 CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}
	if err := validateCRC(data[3:6]); err != nil {
		return errors.New("[ SGP30 ] TVOC CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	command := []uint8{uint8(CmdSetIaqBaselineMSB), uint8(CmdSetIaqBaselineLSB)}
	baseline := append(command, data...)
	if err := d.I2C.Tx(uint16(d.Config.Address), baseline, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send set IAQ baseline command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(10 * time.Millisecond)

	log.Info("[ SGP30 ] IAQ set baseline command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdSetIaqBaseline)
	return nil
}

func (d *Device) SetAbsoluteHumidity(data []uint8) error {
	log := d.log.With("func", "Device.SetAbsoluteHumidity()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Set absolute humidity calibration value", "sensor", d.Config.Name)

	dataFrameLength := 3
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	if err := validateCRC(data); err != nil {
		return errors.New("[ SGP30 ] CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	command := []uint8{uint8(CmdSetAbsoluteHumidityMSB), uint8(CmdSetAbsoluteHumidityLSB)}
	data = append(command, data...)
	if err := d.I2C.Tx(uint16(d.Config.Address), data, nil); err != nil {
		return errors.New("[ SGP30 ] Could not set absolute humidity calibration values for '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(10 * time.Millisecond)

	log.Info("[ SGP30 ] Absolute humidity calibration command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdSetAbsoluteHumidity)
	return nil
}

// Page 10 Measure Test
func (d *Device) MeasureTest(data []uint8) error {
	log := d.log.With("func", "Device.MeasureTest()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Sensor self-test", "sensor", d.Config.Name)

	dataFrameLength := 3
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	command := []uint8{uint8(CmdMeasureIaqMSB), uint8(CmdMeasureTestLSB)}
	if err := d.I2C.Tx(uint16(d.Config.Address), command, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send measure test command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(220 * time.Millisecond)

	if err := d.I2C.Tx(uint16(d.Config.Address), nil, data); err != nil {
		return errors.New("[ SGP30 ] Could not read measure test values from '" + d.Config.Name + "' sensor: " + err.Error())
	}

	testValue := uint16(data[0])<<8 | uint16(data[1])
	if testValue != uint16(MeasureTest) {
		return errors.New("[ SGP30 ] Measure test on '" + d.Config.Name + "' returned unexpected value [ " + Hex16(testValue) + " ]")
	}

	if err := validateCRC(data); err != nil {
		return errors.New("[ SGP30 ] CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	log.Info("[ SGP30 ] Measure test command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdMeasureTest)
	return nil
}

func (d *Device) GetFeatureSet(data []uint8) error {
	log := d.log.With("func", "Device.GetFeatureSet()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Get feature set", "sensor", d.Config.Name)

	dataFrameLength := 3
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	command := []uint8{uint8(CmdGetFeatureSetMSB), uint8(CmdGetFeatureSetLSB)}
	if err := d.I2C.Tx(uint16(d.Config.Address), command, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send get feature set command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(10 * time.Millisecond)

	if err := d.I2C.Tx(uint16(d.Config.Address), nil, data); err != nil {
		return errors.New("[ SGP30 ] Could not read get feature set data from '" + d.Config.Name + "' sensor: " + err.Error())
	}

	testValue := uint16(data[0])<<8 | uint16(data[1])
	if testValue != uint16(FeatureSet) {
		return errors.New("[ SGP30 ] Feature set test returned unexpected value [ " + Hex16(testValue) + " ]")
	}

	if err := validateCRC(data); err != nil {
		return errors.New("[ SGP30 ] CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	log.Info("[ SGP30 ] Feature set command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdGetFeatureSet)
	return nil
}

func (d *Device) MeasureRaw(data []uint8) error {
	log := d.log.With("func", "Device.MeasureRaw()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Measure raw sensor values", "sensor", d.Config.Name)

	dataFrameLength := 6
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	command := []uint8{uint8(CmdMeasureRawMSB), uint8(CmdMeasureRawLSB)}
	if err := d.I2C.Tx(uint16(d.Config.Address), command, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send measure raw command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(25 * time.Millisecond)

	if err := d.I2C.Tx(uint16(d.Config.Address), nil, data); err != nil {
		return errors.New("[ SGP30 ] Could not read raw data from '" + d.Config.Name + "' sensor: " + err.Error())
	}

	if err := validateCRC(data[0:3]); err != nil {
		return errors.New("[ SGP30 ] H2 CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}
	if err := validateCRC(data[3:6]); err != nil {
		return errors.New("[ SGP30 ] C2H6O CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	log.Info("[ SGP30 ] Measure raw values command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdMeasureRaw)
	return nil
}

func (d *Device) GetTvocInceptiveBaseline(data []uint8) error {
	log := d.log.With("func", "Device.GetTvocInceptiveBaseline()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Inceptive baseline sensor values", "sensor", d.Config.Name)

	dataFrameLength := 3
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	command := []uint8{uint8(CmdGetTvocInceptiveBaselineMSB), uint8(CmdGetTvocInceptiveBaselineLSB)}
	if err := d.I2C.Tx(uint16(d.Config.Address), command, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send inceptive baseline set command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(10 * time.Millisecond)

	if err := d.I2C.Tx(uint16(d.Config.Address), nil, data); err != nil {
		return errors.New("[ SGP30 ] Could not read inceptive baseline value from '" + d.Config.Name + "' sensor: " + err.Error())
	}

	if err := validateCRC(data); err != nil {
		return errors.New("[ SGP30 ] CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	log.Info("[ SGP30 ] Inceptive baseline set command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdGetTvocInceptiveBaseline)
	return nil
}

func (d *Device) SetTvocBaseline(data []uint8) error {
	log := d.log.With("func", "Device.SetTvocBaseline()", "params", "([]uint8)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Set TVOC baseline calibration value", "sensor", d.Config.Name)

	dataFrameLength := 2
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	crc, err := calculateCRC(data)
	if err != nil {
		return errors.New("[ SGP30 ] CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	command := []uint8{uint8(CmdSetTvocBaselineMSB), uint8(CmdSetTvocBaselineLSB)}
	data = append(data, crc)
	data = append(command, data...)

	if err := d.I2C.Tx(uint16(d.Config.Address), data, nil); err != nil {
		return errors.New("[ SGP30 ] Could not set TVOC baseline calibration values for '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(10 * time.Millisecond)

	log.Info("[ SGP30 ] TVOC baseline calibration command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdSetTvocBaseline)
	return nil
}

func (d *Device) SoftReset() error {
	log := d.log.With("func", "Device.SoftReset()", "params", "(-)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Sensor soft reset", "sensor", d.Config.Name)

	if err := d.I2C.Tx(uint16(CmdSoftResetMSB), []uint8{uint8(CmdSoftResetLSB)}, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send soft reset command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	log.Info("[ SGP30 ] Soft reset command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdSoftReset)
	return nil
}

func (d *Device) GetSerialId(data []uint8) error {
	log := d.log.With("func", "Device.GetSerialId()", "params", "(-)", "return", "(error)", "lib", "sgp30")
	log.Debug("[ SGP30 ] Get sensor serial ID", "sensor", d.Config.Name)

	dataFrameLength := 9
	if len(data) != dataFrameLength {
		return errors.New("[ SGP30 ] Data frame length invalid.\n\rExpected: [ " + strconv.Itoa(dataFrameLength) + " ]\n\rGot:      [ " + strconv.Itoa(len(data)) + " ]")
	}

	command := []uint8{uint8(CmdGetSerialIdMSB), uint8(CmdGetSerialIdLSB)}
	if err := d.I2C.Tx(uint16(d.Config.Address), command, nil); err != nil {
		return errors.New("[ SGP30 ] Could not send serial ID command to '" + d.Config.Name + "' sensor: " + err.Error())
	}

	// Table 10 Measurement commands
	time.Sleep(1 * time.Millisecond)

	if err := d.I2C.Tx(uint16(d.Config.Address), nil, data); err != nil {
		return errors.New("[ SGP30 ] Could not read serial ID from '" + d.Config.Name + "' sensor: " + err.Error())
	}

	if err := validateCRC(data[0:3]); err != nil {
		return errors.New("[ SGP30 ] CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}
	if err := validateCRC(data[3:6]); err != nil {
		return errors.New("[ SGP30 ] CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}
	if err := validateCRC(data[6:9]); err != nil {
		return errors.New("[ SGP30 ] CRC validation on '" + d.Config.Name + "' error: " + err.Error())
	}

	log.Info("[ SGP30 ] Get serial ID command send to sensor", "bus", d.I2C.String(), "address", Hex8(d.Config.Address), "command", CmdGetSerialId)
	return nil
}

// 6.6 Checksum Calculation
func calculateCRC(data []uint8) (uint8, error) {
	dataFrameLength := 2
	if len(data) == 0 || len(data) > dataFrameLength {
		return 0, fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	var crc uint8 = uint8(CrcBase)  // 0b11111111
	var mask uint8 = uint8(CrcMask) // 0b00110001

	for _, v := range data[0:2] { // <0,2)
		crc = crc ^ v // XOR
		for range 8 {
			msbit := crc & uint8(CrcMsbit) // Most significant BIT
			crc = crc << 1
			if msbit != 0 {
				crc = crc ^ mask // XOR
			}
		}
	}

	return crc, nil
}

// 6.6 Checksum Calculation - TODO: dynamic validation, so it detects if we sent 3, 6, 9 or more bytes
// Dataframe: (DATA0, DATA1, CRC)
func validateCRC(data []uint8) error {
	dataFrameLength := 3
	if len(data) == 0 || len(data) > dataFrameLength {
		return fmt.Errorf("Data frame length invalid.\nExpected: [ %v ]\nGot:      [ %v ]", dataFrameLength, len(data))
	}

	crc, err := calculateCRC(data[0:2])
	if err != nil {
		return err
	}

	if crc != data[2] {
		return fmt.Errorf("Checksum invalid.\nExpected: [%# x ]\nGot:      [%# x ]", data[2], crc)
	}

	return nil
}
