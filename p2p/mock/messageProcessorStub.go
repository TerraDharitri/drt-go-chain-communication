package mock

import (
	"github.com/TerraDharitri/drt-go-chain-communication/p2p"
	"github.com/TerraDharitri/drt-go-chain-core/core"
)

// MessageProcessorStub -
type MessageProcessorStub struct {
	ProcessMessageCalled func(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error
}

// ProcessReceivedMessage -
func (mps *MessageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error {
	if mps.ProcessMessageCalled != nil {
		return mps.ProcessMessageCalled(message, fromConnectedPeer, source)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mps *MessageProcessorStub) IsInterfaceNil() bool {
	return mps == nil
}
