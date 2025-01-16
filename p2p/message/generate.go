//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/Dharitri/protobuf/protobuf  --gogoslick_out=. peerShardMessage.proto

package message
