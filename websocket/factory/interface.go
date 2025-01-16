package factory

import "github.com/TerraDharitri/drt-go-chain-communication/websocket"

// FullDuplexHost defines what a full duplex host should be able to do
type FullDuplexHost interface {
	Send(payload []byte, topic string) error
	SetPayloadHandler(handler websocket.PayloadHandler) error
	Close() error
	IsInterfaceNil() bool
}
