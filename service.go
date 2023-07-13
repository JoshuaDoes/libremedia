package main

import (
	"fmt"
	"net/http"
	"time"
)

var (
	handlers = map[string]Handler{
		"tidal":   &TidalClient{},
		"spotify": &SpotifyClient{},
	}
	providers = make([]string, 0)
)

type Handler interface {
	Provider() string                             //Used for service identification
	SetService(*Service)                          //Provides the handler access to the libremedia service
	Authenticate(*HandlerConfig) (Handler, error) //Attempts to authenticate with the given configuration
	Creator(id string) (*ObjectCreator, error)    //Returns the matching creator object for metadata
	Album(id string) (*ObjectAlbum, error)        //Returns the matching album object for metadata
	Stream(id string) (*ObjectStream, error)      //Returns the matching stream object for metadata
	StreamFormat(w http.ResponseWriter, r *http.Request, stream *ObjectStream, format int) error
	FormatList() []*ObjectFormat                       //Returns all the possible formats as templates ordered from best to worst
	Search(query string) (*ObjectSearchResults, error) //Returns all the available search results that match the query
	Transcribe(obj *ObjectStream) error                //Fills in the stream's transcript with lyrics, closed captioning, subtitles, etc
	ReplaceURI(text string) string                     //Replaces all instances of a URI with a libremedia-acceptable URI, for dynamic hyperlinking
}

type HandlerConfig struct {
	Active     bool   `json:"active"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	DeviceName string `json:"deviceName"`
	BlobPath   string `json:"blobPath"`
}

type Service struct {
	AccessKeys []string                  `json:"accessKeys"`
	BaseURL    string                    `json:"baseURL"`
	Handlers   map[string]*HandlerConfig `json:"handlers"`
	HostAddr   string                    `json:"httpAddr"`

	Grants map[string]*ServiceUser `json:"-"`
}

func (s *Service) Login() error {
	if s.BaseURL[len(s.BaseURL)-1] != '/' {
		s.BaseURL += "/"
	}
	for provider, config := range s.Handlers {
		if handler, exists := handlers[provider]; exists {
			if config.Active {
				newHandler, err := handler.Authenticate(config)
				if err != nil {
					Error.Println("Failed to authenticate " + provider + ": ", err)
					return err
				}
				newHandler.SetService(s)
				handlers[provider] = newHandler
				providers = append(providers, provider)
			} else {
				Trace.Println("Skipping authenticating " + provider)
				delete(handlers, provider)
			}
		}
	}
	return nil
}

func (s *Service) Auth(accessKey string) (*ServiceUser, error) {
	allow := false
	for i := 0; i < len(s.AccessKeys); i++ {
		if s.AccessKeys[i] == accessKey {
			allow = true
			break
		}
	}
	if !allow {
		return nil, fmt.Errorf("invalid accessKey")
	}
	return nil, nil
}

func (s *Service) Stream(w http.ResponseWriter, r *http.Request, stream *ObjectStream, format int) error {
	if stream == nil {
		return fmt.Errorf("stream is nil")
	}
	if stream.Provider == "" {
		return fmt.Errorf("provider not specified")
	}
/*	if stream.Formats == nil || len(stream.Formats) <= format {
		objStream := GetObject(stream.URI, false)
		if objStream != nil {
			stream = objStream.Stream()
			if stream.Formats == nil || len(stream.Formats) <= format {
				return fmt.Errorf("format not available to stream")
			}
		} else {
			return fmt.Errorf("stream not available right now")
		}
	}
*/	if handler, exists := handlers[stream.Provider]; exists {
		return handler.StreamFormat(w, r, stream, format)
	}
	return fmt.Errorf("no handler for provider " + stream.Provider)
}

func (s *Service) Download(w http.ResponseWriter, r *http.Request, stream *ObjectStream, format int) error {
	if stream.Provider == "" {
		return fmt.Errorf("provider not specified")
	}
	if handler, exists := handlers[stream.Provider]; exists {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+stream.FileName()+"\"")
		return handler.StreamFormat(w, r, stream, format)
	}
	return fmt.Errorf("no handler for provider " + stream.Provider)
}

type ServiceUser struct {
	Expires time.Time
}
