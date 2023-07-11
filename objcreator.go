package main

import (
	"encoding/json"
	"time"
)

// ObjectCreator holds metadata about a creator
type ObjectCreator struct {
	Genres      []string         `json:"genres,omitempty"`      //The genres known to define this creator
	Albums      []*Object        `json:"albums,omitempty"`      //The albums from this creator
	Provider    string           `json:"provider,omitempty"`
	URI         string           `json:"uri,omitempty"`         //The URI that refers to this creator object
	Name        string           `json:"name,omitempty"`        //The name of this creator
	Description string           `json:"description,omitempty"` //The description or biography of this creator
	Artworks    []*ObjectArtwork `json:"artworks,omitempty"`    //The artworks for this creator
	DateTime    *time.Time       `json:"datetime,omitempty"`    //The debut date of this creator
	TopStreams  []*Object        `json:"topStreams,omitempty"`  //The top X streams from this creator
	Appearances []*Object        `json:"appearances,omitempty"` //The albums this creator appears on
	Singles     []*Object        `json:"singles,omitempty"`     //The single streams from this creator
	Playlists   []*Object        `json:"playlists,omitempty"`   //The playlists from this creator
	Related     []*Object        `json:"related,omitempty"`     //The creators related to this creator
}

func (obj *ObjectCreator) JSON() []byte {
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return objJSON
}

func (obj *ObjectCreator) IsEmpty() bool {
	return len(obj.Albums) == 0 && len(obj.TopStreams) == 0 && len(obj.Appearances) == 0 && len(obj.Singles) == 0 && len(obj.Playlists) == 0 && len(obj.Related) == 0
}