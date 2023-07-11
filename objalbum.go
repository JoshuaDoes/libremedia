package main

import (
	"encoding/json"
)

// ObjectAlbum holds metadata about an album
type ObjectAlbum struct {
	Discs       []*ObjectDisc    `json:"discs,omitempty"`       //The discs in this album
	Copyrights  []string         `json:"copyrights,omitempty"`  //The copyrights that apply to this album
	Label       string           `json:"label,omitempty"`       //The record label or studio that released this album
	Provider    string           `json:"provider,omitempty"`
	URI         string           `json:"uri,omitempty"`         //The URI that refers to this album object
	Name        string           `json:"name,omitempty"`        //The name of this album
	Description string           `json:"description,omitempty"` //The description of this album
	Artworks    []*ObjectArtwork `json:"artworks,omitempty"`    //The artworks for this album
	DateTime    string           `json:"datetime,omitempty"`    //The release date of this album
	Creators    []*Object        `json:"creators,omitempty"`    //The creators of this album
}

func (obj *ObjectAlbum) JSON() []byte {
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return objJSON
}

func (obj *ObjectAlbum) IsEmpty() bool {
	if len(obj.Discs) == 0 {
		return true
	}
	for i := 0; i < len(obj.Discs); i++ {
		if len(obj.Discs[i].Streams) > 0 {
			return false
		}
	}
	return true
}