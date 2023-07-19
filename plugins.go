package main

import (
	//"encoding/json"
	"fmt"
	//"os/exec"
	"strings"
)

type Packet struct {
	Channel string   //The channel for this packet, should be unique per request as response must match
	Opcode  PacketOp //The operation to call
	Data    []byte   //The data to use for this operation, if necessary
}

/* Example packet order: (CHANNEL|OP|DATA), CHANNEL can be anything but must be unique so a counter is used in this example
-> 01|Ping|
<- 01|Pong|1234567890
-> 02|AuthCheck|
<- 02|AuthResp|true
-> 03|ObjectGet|libremedia:stream:asdf1234
<- 03|ObjectResp|{URI:"libremedia:stream:asdf1234",Type:"stream",Provider:"libremedia",Object:{...}}
-> 04|StreamReq|libremedia:stream:asdf1234
<- 04|StreamResp|{audio data}
<- 04|StreamResp|{audio data}
<- 04|...
<- 04|StreamResp|
-> 05|Terminate|
<- 05|Terminate|
*/

// PacketOp is the operation being performed
type PacketOp int
const (
	//Pings the plugin with expectation of a pong response, plugin will be restarted/reconnected on timeout
	PacketOpPing PacketOp = iota
	//Pongs back in response to a ping, must hold current Unix epoch
	PacketOpPong
	//Requests to check if authentication succeeded, plugin session stalls until responded to or timeout
	PacketOpAuthCheck
	//Responds true or false for authentication, plugin is closed if false
	PacketOpAuthResp
	//Requests an object from the plugin, must hold URI string
	PacketOpObjectGet
	//Responds with an object encoded in JSON
	PacketOpObjectResp
	//Attempts to request the raw data of a stream format from the plugin, must hold PluginStreamRequest encoded in JSON
	PacketOpStreamReq
	//Responds sequentially to a stream request, must send response with empty data to terminate streaming session
	PacketOpStreamResp
	//Terminates the plugin, can be sent to plugin to request for it to shut down, or can be sent by plugin to let libremedia know it will no longer respond and must be closed
	PacketOpTerminate
)

// PluginStreamRequest holds a request to start a streaming session on the transport channel
type PluginStreamRequest struct {
	URI string `json:"uri"`    //The URI of the streamable object
	Format int `json:"format"` //The format number to stream, according to the ordered list of formats in the stream object
}

var (
	channels map[string][]*Packet
)

type Plugin struct {
	Active bool   `json:"active"` //Whether or not this plugin should be loaded
	Path   string `json:"path"`   //Path to plugin
	Method string `json:"method"` //Method for loading the plugin (TCP,BIN)

	closed    bool        `json:"-"` //If Close() was called
	transport interface{} `json:"-"` //Loaded transport handler (TCP socket, OS process, etc)
}

func (p *Plugin) Load() error {
	switch strings.ToLower(p.Method) {
	case "bin":
		//TODO: Execute p.Path as OS process
		//TODO: Store OS process in p.transport
		p.closed = false
		return nil
	case "tcp":
		//TODO: Connect to p.Path as TCP address
		//TODO: Store TCP socket in p.transport
		p.closed = false
		return nil
	}
	return fmt.Errorf("plugin: Invalid method %s", p.Method)
}

func (p *Plugin) Close() error {
	switch strings.ToLower(p.Method) {
	case "bin":
		//TODO: Interpret as OS process and Close()
	case "tcp":
		//TODO: Interpret as TCP socket and Close()
	}
	p.transport = nil
	p.closed = true
	return nil
}

//Receive will read and store all incoming packets from the plugin's transport until either are closed
func (p *Plugin) Receive() error {
	if p.closed {
		return fmt.Errorf("plugin: Cannot receive on closed plugin")
	}
	for {
		if p.closed {
			break
		}
		switch strings.ToLower(p.Method) {
		case "bin":
			//TODO: Interpret as OS process and Read()
			//TODO: Call p.Store(b) when CRLF is reached
		case "tcp":
			//TODO: Interpret as TCP socket and Read()
			//TODO: Call p.Store(b) when CRLF is reached
		}
	}
	return nil
}

//Send will write a request packet to the plugin's transport
func (p *Plugin) Send(op string, b []byte) error {
	if p.closed {
		return fmt.Errorf("plugin: Cannot send on closed plugin")
	}
	//packet := NewPacket(op, b)
	switch strings.ToLower(p.Method) {
	case "bin":
		//TODO: Interpret as OS process and Write(packet)
	case "tcp":
		//TODO: Interpret as TCP socket and Write(packet)
	}
	return nil
}

//Store will parse the given packet and store it in the appropriate channel
func (p *Plugin) Store(b []byte) error {
	var packet *Packet
	//TODO: Decode b into *Packet
	if _, exists := channels[packet.Channel]; !exists {
		channels[packet.Channel] = make([]*Packet, 0)
	}
	channels[packet.Channel] = append(channels[packet.Channel], packet)
	return nil
}

//IsPacketAvailable checks if a packet is available on the specified channel
func (p *Plugin) IsPacketAvailable(channel string) bool {
	if packets, exists := channels[channel]; exists {
		if len(packets) > 0 {
			return true
		}
	}
	return false
}

//ReadPacket reads the next packet from the given channel
func (p *Plugin) ReadPacket(channel string) *Packet {
	if p.IsPacketAvailable(channel) {
		packet := channels[channel][0]
		//TODO: Remove index 0 from slice stored in channels[channel] map
		return packet
	}
	return nil
}