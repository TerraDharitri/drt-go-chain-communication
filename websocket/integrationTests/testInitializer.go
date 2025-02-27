package integrationTests

import (
	"fmt"
	"net"

	"github.com/TerraDharitri/drt-go-chain-communication/websocket"
	"github.com/TerraDharitri/drt-go-chain-communication/websocket/client"
	hostFactory "github.com/TerraDharitri/drt-go-chain-communication/websocket/factory"
	"github.com/TerraDharitri/drt-go-chain-communication/websocket/server"
	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/marshal/factory"
)

const retryDurationInSeconds = 1

var (
	marshaller, _       = factory.NewMarshalizer("gogo protobuf")
	payloadConverter, _ = websocket.NewWebSocketPayloadConverter(marshaller)
)

func createClient(url string, log core.Logger) (hostFactory.FullDuplexHost, error) {
	return client.NewWebSocketClient(client.ArgsWebSocketClient{
		RetryDurationInSeconds:     retryDurationInSeconds,
		WithAcknowledge:            true,
		URL:                        url,
		PayloadConverter:           payloadConverter,
		Log:                        log,
		DropMessagesIfNoConnection: false,
		AckTimeoutInSeconds:        retryDurationInSeconds,
		PayloadVersion:             1,
	})
}

func createServer(url string, log core.Logger) (hostFactory.FullDuplexHost, error) {
	return server.NewWebSocketServer(server.ArgsWebSocketServer{
		RetryDurationInSeconds:     retryDurationInSeconds,
		WithAcknowledge:            true,
		URL:                        url,
		PayloadConverter:           payloadConverter,
		Log:                        log,
		DropMessagesIfNoConnection: false,
		AckTimeoutInSeconds:        retryDurationInSeconds,
		PayloadVersion:             1,
	})
}

func getFreePort() string {
	// Listen on port 0 to get a free port
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = l.Close()
	}()

	// Get the port number that was assigned
	addr := l.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port)
}
