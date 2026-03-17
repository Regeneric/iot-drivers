```go
// ************************************************************************
// = 1-Wire ===
// ------------------------------------------------------------------------
owConnections, owClose, err := onewire.Setup(&config.OneWire)
if err != nil {
    slog.Error("Critical 1-Wire init failure", "error", err)
} else {
    defer owClose()
}

ow0, ok = owConnections["ow0"]
if !ok {
    slog.Error("Missing 1-Wire device configuration", "name", "ow0")
}
// ------------------------------------------------------------------------
```