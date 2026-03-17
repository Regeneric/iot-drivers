```go
var config = config.Config{
	I2C: *i2c.Config{
		Enable: true,
		Devices: map[string]*i2c.Device{
			"i2c0": {
				Enable: true,
				Name:   "i2c0",
				Speed:  100 * machine.KHz,
                // Pins: use default mapping if nil
			},
			"i2c1": {
				Enable: true,
				Name:   "i2c1",
				Speed:  400 * machine.KHz,
				Pins: &i2c.Pins{
					SCL: machine.GP27,
					SDA: machine.GP26,
				},
			},
		},
	},
}


i2cConnections, i2cClose, err := hi2c.Setup(config.I2C)
if err != nil {
    println("Critical I2C init failure:", err.Error())
} else {
    defer i2cClose()
}

i2c0, ok := i2cConnections["i2c0"]
if !ok {
    println("Missing I2C bus configuration;", "name", "i2c0")
}
```