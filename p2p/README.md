# drt-chain-go p2p components

The `Messenger` interface with its implementation are 
used to define the way to communicate between Dharitri nodes. 

There are 2 ways to send data to the other peers:
1. Broadcasting messages on a `pubsub` using topics;
2. Direct sending messages to the connected peers.

The first type is used to send messages that have to reach every node 
(from corresponding shard, metachain, consensus group, etc.). The second type is
used to resolve requests coming from directly connected peers. 
