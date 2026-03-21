package lora

type Node struct {
	hw  Transceiver
	cfg *sx126x.Config
}
