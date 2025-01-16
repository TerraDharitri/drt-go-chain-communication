//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/Dharitri/protobuf/protobuf  --gogoslick_out=. topicMessage.proto
package data
