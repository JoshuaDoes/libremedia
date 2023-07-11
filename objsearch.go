package main

import (
	"encoding/json"
)

// ObjectSearchResults holds the results for a given search query
type ObjectSearchResults struct {
	Query     string    `json:"query,omitempty"`     //The query that generated these results
	Streams   []*Object `json:"streams,omitempty"`   //The stream results for this query
	Creators  []*Object `json:"creators,omitempty"`  //The creator results for this query
	Albums    []*Object `json:"albums,omitempty"`    //The album results for this query
	Provider string `json:"provider,omitempty"`
}

func (obj *ObjectSearchResults) JSON() []byte {
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return objJSON
}

func (obj *ObjectSearchResults) IsEmpty() bool {
	return len(obj.Streams) == 0 && len(obj.Creators) == 0 && len(obj.Albums) == 0
}