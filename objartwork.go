package main

import (
	"encoding/json"
)

// ObjectArtwork holds metadata about an artwork
type ObjectArtwork struct {
	Provider string `json:"provider,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	URL    string `json:"url,omitempty"`
	Type   string `json:"type,omitempty"` //The file type, i.e. mp4 or jpg
}

func (obj *ObjectArtwork) JSON() []byte {
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return objJSON
}

func NewObjArtwork(provider, fileType, url string, width, height int) *ObjectArtwork {
	return &ObjectArtwork{provider, width, height, url, fileType}
}