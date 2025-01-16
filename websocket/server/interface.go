package server

import (
	"github.com/TerraDharitri/drt-go-chain-communication/websocket"
)

type transceiversAndConnHandler interface {
	addTransceiverAndConn(transceiver Transceiver, conn websocket.WSConClient)
	remove(id string)
	getAll() map[string]tupleTransceiverAndConn
}

// Transceiver defines what a WebSocket transceiver should be able to do
type Transceiver interface {
	Send(payload []byte, topic string, connection websocket.WSConClient) error
	SetPayloadHandler(handler websocket.PayloadHandler) error
	Listen(connection websocket.WSConClient) (closed bool)
	Close() error
}
