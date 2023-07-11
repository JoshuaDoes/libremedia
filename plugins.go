package main

import (
	"fmt"
	//"os/exec"
	"strings"
)

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
	if _, exists := channels[packet.Id]; !exists {
		channels[packet.Id] = make([]*Packet, 0)
	}
	channels[packet.Id] = append(channels[packet.Id], packet)
	return nil
}

//IsPacketAvailable checks if a packet is available on the specified channel
func (p *Plugin) IsPacketAvailable(id string) bool {
	if channel, exists := channels[id]; exists {
		if len(channel) > 0 {
			return true
		}
	}
	return false
}

//ReadPacket reads the next packet from the given channel
func (p *Plugin) ReadPacket(id string) *Packet {
	if channel, exists := channels[id]; exists {
		if len(channel) > 0 {
			packet := channel[0]
			//TODO: Remove index 0 from slice stored in channels[id] map
			return packet
		}
	}
	return nil
}

type Packet struct {
	Id   string `json:"id"`   //The channel for this packet, should be unique per request as response will match
	Op	 string `json:"op"`   //The operation to call
	Data []byte `json:"data"` //The data for this operation
}

func NewPacket(op string, b []byte) []byte {
	return nil
}
