package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// Object holds a metadata object
type Object struct {
	URI       string           `json:"uri,omitempty"`       //The URI that matches this object
	Type      string           `json:"type,omitempty"`      //search, stream, creator, album
	Provider  string           `json:"provider,omitempty"`  //The service that provides this object
	Expires   *time.Time       `json:"expires,omitempty"`   //When this object should expire by
	LastMod   *time.Time       `json:"lastMod,omitempty"`   //When this object was last altered
	Object    *json.RawMessage `json:"object,omitempty"`    //Holds either the raw object or a string containing the object's reference URI
	Expanding bool             `json:"expanding,omitempty"` //Whether or not this object is in the process of internal expansion
	Expanded  bool             `json:"expanded,omitempty"`  //Whether or not this object has been expanded internally
}

// JSON returns this object as serialized JSON
func (obj *Object) JSON() ([]byte, error) {
	return json.Marshal(obj)
}

func (obj *Object) SearchResults() *ObjectSearchResults {
	if obj.Object == nil {
		return nil
	}
	switch obj.Type {
	case "search":
		ret := &ObjectSearchResults{}
		if objJSON, err := obj.Object.MarshalJSON(); err == nil {
			if err := json.Unmarshal(objJSON, &ret); err == nil {
				return ret
			}
		}
	}
	return nil
}

func (obj *Object) Creator() *ObjectCreator {
	if obj.Object == nil {
		return nil
	}
	switch obj.Type {
	case "artist", "creator", "user", "channel", "chan", "streamer":
		ret := &ObjectCreator{}
		if objJSON, err := obj.Object.MarshalJSON(); err == nil {
			if err := json.Unmarshal(objJSON, &ret); err == nil {
				return ret
			}
		}
	}
	return nil
}

func (obj *Object) Album() *ObjectAlbum {
	if obj.Object == nil {
		return nil
	}
	switch obj.Type {
	case "album":
		ret := &ObjectAlbum{}
		if objJSON, err := obj.Object.MarshalJSON(); err == nil {
			if err := json.Unmarshal(objJSON, &ret); err == nil {
				return ret
			}
		}
	}
	return nil
}

func (obj *Object) Stream() *ObjectStream {
	if obj.Object == nil {
		return nil
	}
	switch obj.Type {
	case "track", "song", "video", "audio", "stream":
		ret := &ObjectStream{}
		if objJSON, err := obj.Object.MarshalJSON(); err == nil {
			if err := json.Unmarshal(objJSON, &ret); err == nil {
				return ret
			}
		}
	}
	return nil
}

// Sync writes this object to the cache
func (obj *Object) Sync() {
	if obj.URI == "" {
		return
	}
	expiryTime := time.Now()
	switch obj.Type {
	case "search": //2 hours
		objSearch := obj.SearchResults()
		if objSearch != nil && objSearch.IsEmpty() {
			return
		}
		expiryTime = expiryTime.Add(time.Hour * 2)
	case "artist", "creator", "user", "channel", "chan", "streamer": //12 hours
		objCreator := obj.Creator()
		if objCreator != nil && objCreator.IsEmpty() {
			return
		}
		expiryTime = expiryTime.Add(time.Hour * 12)
	case "album", "track", "song", "video", "audio", "stream": //30 days
		objAlbum := obj.Album()
		if objAlbum != nil && objAlbum.IsEmpty() {
			return
		}
		expiryTime = expiryTime.Add(time.Hour * (24 * 30))
	}
	obj.LastMod = &expiryTime
	obj.Expires = &expiryTime

	objData, err := obj.JSON()
	if err != nil {
		return
	}
	splitURI := strings.Split(obj.URI, ":")
	pathURL := "cache/"
	fileName := ""
	for i := 0; i < len(splitURI); i++ {
		if i == len(splitURI)-1 {
			fileName = splitURI[i] + ".json"
			break
		}
		pathURL += splitURI[i] + "/"
	}
	os.MkdirAll(pathURL, 0777)
	pathURL += fileName
	ioutil.WriteFile(pathURL, objData, 0777)
}

// Expand fills in all top-level object arrays with completed objects
func (src *Object) Expand() {
	if src.URI == "" {
		return
	}
	//Check if object is being expanded right now
	if src.Expanding {
		//Sleep and try again
		return
	}
	src.Expanding = true
	src.Expanded = false
	src.Sync()
	Trace.Println("Expanding " + src.URI)
	switch src.Type {
	case "search":
		if search := src.SearchResults(); search != nil {
			syncSearch := func() {
				searchJSON, err := json.Marshal(search)
				if err == nil {
					src.Object = &json.RawMessage{}
					src.Object.UnmarshalJSON(searchJSON)
					src.Sync()
				}
			}
			for i := 0; i < len(search.Streams); i++ {
				if search.Streams[i].URI == "" {
					continue
				}
				search.Streams[i] = GetObject(search.Streams[i].URI)
				syncSearch()
			}
			for i := 0; i < len(search.Creators); i++ {
				if search.Creators[i].URI == "" {
					continue
				}
				search.Creators[i] = GetObject(search.Creators[i].URI)
				syncSearch()
			}
			for i := 0; i < len(search.Albums); i++ {
				if search.Albums[i].URI == "" {
					continue
				}
				search.Albums[i] = GetObject(search.Albums[i].URI)
				syncSearch()
			}
		}
	case "artist", "creator", "user", "channel", "chan", "streamer":
		if creator := src.Creator(); creator != nil {
			syncCreator := func() {
				creatorJSON, err := json.Marshal(creator)
				if err == nil {
					src.Object = &json.RawMessage{}
					src.Object.UnmarshalJSON(creatorJSON)
					src.Sync()
				}
			}
			for i := 0; i < len(creator.TopStreams); i++ {
				if creator.TopStreams[i].URI == "" {
					continue
				}
				creator.TopStreams[i] = GetObject(creator.TopStreams[i].URI)
				syncCreator()
			}
			for i := 0; i < len(creator.Albums); i++ {
				if creator.Albums[i].URI == "" {
					continue
				}
				creator.Albums[i] = GetObject(creator.Albums[i].URI)
				syncCreator()
			}
			for i := 0; i < len(creator.Appearances); i++ {
				if creator.Appearances[i].URI == "" {
					continue
				}
				creator.Appearances[i] = GetObject(creator.Appearances[i].URI)
				syncCreator()
			}
			for i := 0; i < len(creator.Singles); i++ {
				if creator.Singles[i].URI == "" {
					continue
				}
				creator.Singles[i] = GetObject(creator.Singles[i].URI)
				syncCreator()
			}
			for i := 0; i < len(creator.Related); i++ {
				if creator.Related[i].URI == "" {
					continue
				}
				creator.Related[i] = GetObject(creator.Related[i].URI)
				syncCreator()
			}
		}
	case "album":
		if album := src.Album(); album != nil {
			syncAlbum := func() {
				albumJSON, err := json.Marshal(album)
				if err == nil {
					src.Object = &json.RawMessage{}
					src.Object.UnmarshalJSON(albumJSON)
					src.Sync()
				}
			}
			for i := 0; i < len(album.Creators); i++ {
				if album.Creators[i].URI == "" {
					continue
				}
				album.Creators[i] = GetObject(album.Creators[i].URI)
				syncAlbum()
			}
			for i := 0; i < len(album.Discs); i++ {
				for j := 0; j < len(album.Discs[i].Streams); j++ {
					if album.Discs[i].Streams[j].URI == "" {
						continue
					}
					album.Discs[i].Streams[j] = GetObject(album.Discs[i].Streams[j].URI)
					syncAlbum()
				}
			}
		}
	case "track", "song", "video", "audio", "stream":
		if stream := src.Stream(); stream != nil {
			syncStream := func() {
				streamJSON, err := json.Marshal(stream)
				if err == nil {
					src.Object = &json.RawMessage{}
					src.Object.UnmarshalJSON(streamJSON)
					src.Sync()
				}
			}
			stream.Album = GetObject(stream.Album.URI)
			syncStream()
			for i := 0; i < len(stream.Creators); i++ {
				if stream.Creators[i].URI == "" {
					continue
				}
				stream.Creators[i] = GetObject(stream.Creators[i].URI)
				syncStream()
			}
		}
	}
	if src.Object != nil {
		src.Expanding = false
		src.Expanded = true
		Trace.Println("Finished expanding " + src.URI)
	} else {
		src.Expanding = false
		Trace.Println("Failed to expand " + src.URI)
	}
	src.Sync()
}

// GetObjectCached returns a new object from the cache that links to a given URI
func GetObjectCached(uri string) (obj *Object) {
	Trace.Println("Retrieving " + uri + " from the cache")

	splitURI := strings.Split(uri, ":")
	pathURL := "cache/"
	for i := 0; i < len(splitURI); i++ {
		if i == len(splitURI)-1 {
			break
		}
		pathURL += splitURI[i] + "/"
	}
	pathURL += splitURI[len(splitURI)-1] + ".json"

	//Check if the object exists
	info, err := os.Stat(pathURL)
	if os.IsNotExist(err) {
		//Trace.Println("Object " + uri + " does not exist in cache")
		return nil
	}
	if info.IsDir() {
		Warning.Println("Object " + uri + " points to a directory")
		return nil
	}

	//Try reading the object from disk
	objData, err := ioutil.ReadFile(pathURL)
	if err != nil {
		Error.Println("Object " + uri + " failed to read from cache, garbage collecting it instead")
		os.Remove(pathURL)
		return nil
	}

	//Map the object into memory, or invalidate it to be resynced if that fails
	obj = &Object{Object: &json.RawMessage{}}
	err = json.Unmarshal(objData, obj)
	if err != nil {
		Error.Println("Object " + uri + " failed to map into memory, garbage collecting it instead")
		os.Remove(pathURL)
		return nil
	}

	//Check if object expired and was missed during cleanup
	if obj.Expires != nil && time.Now().After(*obj.Expires) {
		Error.Println("Object " + uri + " expired, garbage collecting")
		os.Remove(pathURL)
		return nil
	}

	return obj
}

// NewObjError returns an error object
func NewObjError(msg string) (obj *Object) {
	obj = &Object{
		Type: "error",
		Provider: "libremedia",
		Object: &json.RawMessage{},
	}
	errJSON, err := json.Marshal(&exporterr{Error: msg})
	if err == nil {
		obj.Object.UnmarshalJSON(errJSON)
	}
	return obj
}
