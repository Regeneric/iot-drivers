```go
// ************************************************************************
// = SPI ===
// ------------------------------------------------------------------------
spiConnections, spiClose, err := spi.Setup(&config.SPI)
if err != nil {
    slog.Error("Critical SPI init failure", "error", err)
} else {
    defer spiClose()
}

spi0, ok := spiConnections["spi0"]
if !ok {
    slog.Error("Missing SPI device configuration", "name", "spi0")
}
// ------------------------------------------------------------------------
```