package message

import (
	"github.com/TerraDharitri/drt-go-chain-communication/p2p"
	"github.com/TerraDharitri/drt-go-chain-core/core"
)

var _ p2p.MessageP2P = (*Message)(nil)

// Message is a data holder struct
type Message struct {
	FromField            []byte
	DataField            []byte
	PayloadField         []byte
	SeqNoField           []byte
	TopicField           string
	SignatureField       []byte
	KeyField             []byte
	PeerField            core.PeerID
	TimestampField       int64
	BroadcastMethodField p2p.BroadcastMethod
}

// From returns the message originator's peer ID
func (m *Message) From() []byte {
	return m.FromField
}

// Data returns the useful message that was actually sent
func (m *Message) Data() []byte {
	return m.DataField
}

// Payload returns the encapsulated message along with metadata such as timestamp
func (m *Message) Payload() []byte {
	return m.PayloadField
}

// SeqNo returns the message sequence number
func (m *Message) SeqNo() []byte {
	return m.SeqNoField
}

// Topic returns the topic on which the message was sent
func (m *Message) Topic() string {
	return m.TopicField
}

// Signature returns the message signature
func (m *Message) Signature() []byte {
	return m.SignatureField
}

// Key returns the message public key (if it can not be recovered from the From field)
func (m *Message) Key() []byte {
	return m.KeyField
}

// Peer returns the peer that originated the message
func (m *Message) Peer() core.PeerID {
	return m.PeerField
}

// Timestamp returns the message timestamp to prevent endless re-processing of the same message
func (m *Message) Timestamp() int64 {
	return m.TimestampField
}

// BroadcastMethod returns the broadcast method of the message
func (m *Message) BroadcastMethod() p2p.BroadcastMethod {
	return m.BroadcastMethodField
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *Message) IsInterfaceNil() bool {
	return m == nil
}
