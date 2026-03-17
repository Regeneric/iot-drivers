```go
// ************************************************************************
// = I2C ===
// ------------------------------------------------------------------------
i2cConnections, i2cClose, err := i2c.Setup(&config.I2C)
if err != nil {
    slog.Error("Critical I2C init failure", "error", err)
} else {
    defer i2cClose()
}

i2c0, ok = i2cConnections["i2c0"]
if !ok {
    slog.Error("Missing I2C device configuration", "name", "i2c")
}
// ------------------------------------------------------------------------
```