package main

import (
	"encoding/json"
	"strconv"
)

// ObjectFormat holds the URL to stream this file and its codec and format information
type ObjectFormat struct {
	Provider   string      `json:"provider,omitempty"`
	ID         int         `json:"id,omitempty"`         //The ID of this format's quality; lower is better
	Name       string      `json:"name,omitempty"`       //The title or name of this format type
	URL        string      `json:"url,omitempty"`        //Ex (qualityId would be 2 for OGG 320Kbps if Spotify): /stream/{uri}?quality={qualityId}
	Format     string      `json:"format,omitempty"`     //Ex: ogg, mp4
	Codec      string      `json:"codec,omitempty"`      //Ex: vorbis, h264
	BitRate    int32       `json:"bitrate,omitempty"`    //Ex: 320000, 5500
	BitDepth   int         `json:"bitdepth,omitempty"`   //Ex: 8, 16, 24, 32
	SampleRate int32       `json:"samplerate,omitempty"` //Ex: 96000 or 44100, 30 or 60
	File       interface{} `json:"-"`                    //A place for the live session to store a temporary file
}

func (obj *ObjectFormat) JSON() []byte {
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return objJSON
}

func (obj *ObjectFormat) GenerateURL(uri string) {
	obj.URL = service.BaseURL + "v1/stream/" + uri + "?format=" + strconv.Itoa(obj.ID)
}