package factory

import (
	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/client"
	outportData "github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-communication-go/websocket/driver"
	"github.com/multiversx/mx-chain-communication-go/websocket/server"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// ArgsWebSocketDriverFactory holds the arguments needed for creating a webSocketsDriverFactory
type ArgsWebSocketDriverFactory struct {
	WebSocketConfig outportData.WebSocketConfig
	Marshaller      marshal.Marshalizer
	Log             core.Logger
}

// NewWebSocketDriver will handle the creation of all the components needed to create an outport driver that sends data over WebSocket
func NewWebSocketDriver(args ArgsWebSocketDriverFactory) (websocket.Driver, error) {
	var host websocket.HostWebSocket
	var err error
	if args.WebSocketConfig.IsServer {
		host, err = createWebSocketServer(args)
	} else {
		host, err = createWebSocketClient(args)
	}

	if err != nil {
		return nil, err
	}

	host.Start()

	return driver.NewWebsocketDriver(
		driver.ArgsWebSocketDriver{
			Marshaller:      args.Marshaller,
			WebsocketSender: host,
			Log:             args.Log,
		},
	)
}

// TODO merge the ArgsWebSocketClient and ArgsWebSocketServer as they look the same and remove the duplicated arguments build
func createWebSocketClient(args ArgsWebSocketDriverFactory) (websocket.HostWebSocket, error) {
	payloadConverter, err := websocket.NewWebSocketPayloadConverter(args.Marshaller)
	if err != nil {
		return nil, err
	}

	return client.NewWebSocketClient(client.ArgsWebSocketClient{
		RetryDurationInSeconds: args.WebSocketConfig.RetryDurationInSec,
		WithAcknowledge:        args.WebSocketConfig.WithAcknowledge,
		URL:                    args.WebSocketConfig.URL,
		PayloadConverter:       payloadConverter,
		Log:                    args.Log,
		BlockingAckOnError:     args.WebSocketConfig.BlockingAckOnError,
	})
}

func createWebSocketServer(args ArgsWebSocketDriverFactory) (websocket.HostWebSocket, error) {
	payloadConverter, err := websocket.NewWebSocketPayloadConverter(args.Marshaller)
	if err != nil {
		return nil, err
	}

	host, err := server.NewWebSocketServer(server.ArgsWebSocketServer{
		RetryDurationInSeconds: args.WebSocketConfig.RetryDurationInSec,
		WithAcknowledge:        args.WebSocketConfig.WithAcknowledge,
		URL:                    args.WebSocketConfig.URL,
		PayloadConverter:       payloadConverter,
		Log:                    args.Log,
		BlockingAckOnError:     args.WebSocketConfig.BlockingAckOnError,
	})
	if err != nil {
		return nil, err
	}

	return host, nil
}
