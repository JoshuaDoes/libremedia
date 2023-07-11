package main

import (
	"encoding/json"
)

// ObjectDisc holds a list of streams
type ObjectDisc struct {
	Provider string           `json:"provider,omitempty"`
	Disc     int              `json:"disc,omitempty"`     //The disc or part number of an album
	Name     string           `json:"name,omitempty"`     //The name of this disc
	Artworks []*ObjectArtwork `json:"artworks,omitempty"` //The artworks for this disc
	Streams  []*Object        `json:"streams,omitempty"`  //The streams on this disc
}

func (obj *ObjectDisc) JSON() []byte {
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return objJSON
}