package main

import (
	"encoding/json"
	"strings"

	"github.com/rhnvrm/lyric-api-go"
)

// ObjectStream holds metadata about a stream and the available formats to stream
type ObjectStream struct {
	Track      int               `json:"track,omitempty"`      //The track number of the disc/album that holds this stream
	Name       string            `json:"name,omitempty"`       //The name of this file
	Visual     bool              `json:"visual,omitempty"`     //Set to true if designed to be streamed as video
	Explicit   bool              `json:"explicit,omitempty"`   //Whether or not this stream is explicit
	Duration   int64             `json:"duration,omitempty"`   //The duration of this file in seconds
	Formats    []*ObjectFormat   `json:"formats,omitempty"`    //The stream formats available for this file
	Language   string            `json:"language,omitempty"`   //Ex: en, english, es, spanish, etc - something unified for purpose
	Transcript *ObjectTranscript `json:"transcript,omitempty"` //Ex: lyrics for a song, closed captions or transcript of a video/recording, etc
	Artworks   []*ObjectArtwork  `json:"artworks,omitempty"`   //The artworks for this file
	Creators   []*Object         `json:"creators,omitempty"`   //The creators of this file
	Album      *Object           `json:"album,omitempty"`      //The album that holds this file
	DateTime   string            `json:"datetime,omitempty"`   //The release date of this file
	Provider   string            `json:"provider,omitempty"`
	URI        string            `json:"uri,omitempty"`     //The URI that refers to this stream object
	ID         string            `json:"id,omitempty"`      //The ID that refers to this stream object
	Extdata    map[string]string `json:"extdata,omitempty"` //To store additional service-specific and file-defined metadata
}

func (obj *ObjectStream) JSON() []byte {
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	return objJSON
}

func (obj *ObjectStream) FileName() string {
	creatorName := obj.Creators[0].Creator().Name
	album := obj.Album.Album()
	albumName := album.Name
	albumDate := album.DateTime
	trackName := obj.Name

	fileName := creatorName + " - " + albumName
	if albumDate != "" {
		fileName += " " + albumDate
	}
	fileName += " - " + trackName + "." + obj.Formats[0].Format
	return fileName
}

func (obj *ObjectStream) GetFormat(format int) *ObjectFormat {
	for i := 0; i < len(obj.Formats); i++ {
		if obj.Formats[i].ID == format {
			return obj.Formats[i]
		}
	}
	return nil
}

func (obj *ObjectStream) Transcribe() {
	if obj.Transcript != nil && len(obj.Transcript.Lines) > 0 {
		return //You should expire the object if you want to resync it
	}

	//Try the source of the object first
	//TODO: Check all providers when providers can be an array
	if handler, exists := handlers[obj.Provider]; exists {
		if err := handler.Transcribe(obj); err == nil {
			return
		}
	}

	//Make sure we at least know the creator and stream names first
	if len(obj.Creators) > 0 && obj.Name != "" {
		l := lyrics.New()

		//We want the first creator that matches a result
		for i := 0; i < len(obj.Creators); i++ {
			creator := obj.Creators[i].Creator()
			if creator != nil {
				transcript, err := l.Search(creator.Name, obj.Name)
				if err == nil {
					obj.Transcript = &ObjectTranscript{
						Provider:         "libremedia",
						ProviderLyricsID: obj.ID,
						ProviderTrackID:  obj.ID,
						Lines:            make([]*ObjectTranscriptLine, 0),
					}
					lines := strings.Split(transcript, "\n")
					for j := 0; j < len(lines); j++ {
						line := lines[j]
						obj.Transcript.Lines = append(obj.Transcript.Lines, &ObjectTranscriptLine{Text: line})
					}
					return
				}
			}
		}
	}
}
