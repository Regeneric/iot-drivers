```go
var config = config.Config{
	SPI: *spi.Config{
		Enable: true,
		Devices: map[string]*spi.Device{
			"spi0": {
				Enable:   true,
				Name:     "spi0",
				Speed:    2000000,
				Mode:     0,
				LSBFirst: false,
                // Pins: use default mapping if nil
			},
            "spi1": {
				Enable:   true,
				Name:     "spi1",
				Speed:    10000000,
				Mode:     0,
				LSBFirst: false,
                Pins: &spi.Pins{
                    MISO: machine.GP8,
                    MOSI: machine.GP20,
                    SCK:  machine.GP19,
                }
			},
		},
	},
}

spiConnections, spiClose, err := spi.Setup(config.SPI)
if err != nil {
    logger.Error("Critical SPI init failure:", "error", err)
} else {
    defer spiClose()
}

spi0, ok := spiConnections["spi0"]
if !ok {
    logger.Error("Missing SPI bus configuration;", "name", "spi0")
}
```