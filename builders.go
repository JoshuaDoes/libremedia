package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GetObject returns an object, either from the cache, or live if possible
func GetObject(uri string) (obj *Object) {
	//Try the cache first, it will update frequently during a live expansion
	obj = GetObjectCached(uri)
	if obj != nil {
		return obj
	}

	//Fetch the object live
	obj = GetObjectLive(uri)
	if obj == nil {
		return
	}

	obj.Sync()
	return
}

// GetObjectLive returns a live object from a given URI
func GetObjectLive(mediaURI string) (obj *Object) {
	if mediaURI == "" {
		Error.Println("Cannot get object with empty mediaURI")
		return nil
	}

	obj = &Object{URI: mediaURI, Object: &json.RawMessage{}}
	Trace.Println("Fetching " + mediaURI + " live")

	splitURI := strings.Split(mediaURI, ":")
	switch splitURI[0] {
	case "bestmatch": //Returns an object that best matches the given search query
		if len(splitURI) < 2 {
			return NewObjError("bestmatch: need query")
		}
		query := strings.ReplaceAll(splitURI[1], "+", " ")
		query = strings.ReplaceAll(query, "%20", " ")
		query = strings.ToLower(query)
		searchResultsObj := GetObjectLive("search:" + splitURI[1])
		if searchResultsObj.Type != "search" {
			return searchResultsObj
		}
		searchResults := searchResultsObj.SearchResults()
		if len(searchResults.Streams) > 0 {
			return GetObjectLive(searchResults.Streams[0].Stream().URI)
		}
		if len(searchResults.Creators) > 0 {
			return GetObjectLive(searchResults.Creators[0].Creator().URI)
		}
		if len(searchResults.Albums) > 0 {
			return GetObjectLive(searchResults.Albums[0].Album().URI)
		}
		return NewObjError("bestmatch: try a better query")
	case "search": //Main search handler
		if len(splitURI) < 2 {
			return NewObjError("search: need query")
		}
		query := strings.ReplaceAll(splitURI[1], "+", " ")
		query = strings.ReplaceAll(query, "%20", " ")
		query = strings.ToLower(query)
		obj.URI = "search:" + query
		results := &ObjectSearchResults{Query: query}
		for i := 0; i < len(providers); i++ {
			Trace.Println("Searching for '" + query + "' on " + providers[i])
			handler := handlers[providers[i]]
			res, err := handler.Search(query)
			if err != nil {
				Error.Printf("Error searching on %s: %v", providers[i], err)
				continue
			}
			if len(res.Creators) > 0 {
				results.Creators = append(results.Creators, res.Creators...)
			}
			if len(res.Albums) > 0 {
				results.Albums = append(results.Albums, res.Albums...)
			}
			if len(res.Streams) > 0 {
				results.Streams = append(results.Streams, res.Streams...)
			}
		}
		obj.Type = "search"
		obj.Provider = "libremedia"
		resultsJSON, err := json.Marshal(results)
		if err != nil {
			Error.Printf("Unable to marshal search results: %v\n", err)
			return NewObjError(fmt.Sprintf("invalid search %s: %v", query, err))
		}

		if err := obj.Object.UnmarshalJSON(resultsJSON); err != nil {
			Error.Printf("Unable to unmarshal search results: %v\n", err)
			return NewObjError(fmt.Sprintf("invalid search %s: %v", query, err))
		}
		return
	}

	if handler, exists := handlers[splitURI[0]]; exists {
		obj.Provider = splitURI[0]
		if len(splitURI) > 2 {
			id := splitURI[2]
			switch splitURI[1] {
			case "artist", "creator", "user", "channel", "chan", "streamer":
				creator, err := handler.Creator(id)
				if err != nil {
					Error.Printf("Invalid creator %s: %v\n", id, err)
					return NewObjError(fmt.Sprintf("invalid creator %s: %v", id, err))
				}
				obj.Type = "creator"
				creatorJSON, err := json.Marshal(creator)
				if err != nil {
					Error.Printf("Unable to marshal creator: %v\n", err)
					return NewObjError(fmt.Sprintf("invalid creator %s: %v", id, err))
				}
				if err := obj.Object.UnmarshalJSON(creatorJSON); err != nil {
					Error.Printf("Unable to unmarshal creator: %v\n", err)
					return NewObjError(fmt.Sprintf("invalid creator %s: %v", id, err))
				}
			case "album":
				Trace.Printf("Searching for album %s\n", mediaURI)
				album, err := handler.Album(id)
				if err != nil {
					Error.Printf("Invalid album %s: %v\n", id, err)
					return NewObjError(fmt.Sprintf("invalid album %s: %v", id, err))
				}
				Trace.Printf("Found album %s\n", mediaURI)
				obj.Type = "album"
				albumJSON, err := json.Marshal(album)
				if err != nil {
					Error.Printf("Unable to marshal album: %v\n", err)
					return NewObjError(fmt.Sprintf("invalid album %s: %v", id, err))
				}
				if err := obj.Object.UnmarshalJSON(albumJSON); err != nil {
					Error.Printf("Unable to unmarshal album: %v\n", err)
					return NewObjError(fmt.Sprintf("invalid album %s: %v", id, err))
				}
				Trace.Printf("Successfully loaded album %s\n", mediaURI)
			case "track", "song", "video", "audio", "stream":
				stream, err := handler.Stream(id)
				if err != nil {
					Error.Printf("Invalid stream %s: %v\n", id, err)
					return NewObjError(fmt.Sprintf("invalid stream %s: %v", id, err))
				}
				obj.Type = "stream"
				stream.Transcribe()
				streamJSON, err := json.Marshal(stream)
				if err != nil {
					Error.Printf("Unable to marshal stream: %v\n", err)
					return NewObjError(fmt.Sprintf("invalid stream %s: %v", id, err))
				}
				if err := obj.Object.UnmarshalJSON(streamJSON); err != nil {
					Error.Printf("Unable to unmarshal stream: %v\n", err)
					return NewObjError(fmt.Sprintf("invalid stream %s: %v", id, err))
				}
			}

			Info.Printf("Successfully found live object for %s\n", mediaURI)
			return obj
		}
	}

	Error.Printf("Failed to find live object for %s\n", mediaURI)
	return nil
}
