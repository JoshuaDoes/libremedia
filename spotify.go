package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/eolso/librespot-golang/Spotify"
	"github.com/eolso/librespot-golang/librespot"
	"github.com/eolso/librespot-golang/librespot/core"
	"github.com/eolso/librespot-golang/librespot/mercury"
	"github.com/eolso/librespot-golang/librespot/utils"
)

/*
	- GetPlaylist(playlistID string) (*ObjectPlaylist, error)
	- GetUserPlaylist(userID, playlistID string) (*ObjectPlaylist, error)
	- GetSuggest(query string) ([]string, error)
*/

var spoturire = regexp.MustCompile(`<a href="spotify:(.*?):(.*?)">(.*?)<\/a>`)

// SpotifyClient holds a Spotify client
type SpotifyClient struct {
	sync.Mutex

	Session *core.Session
	Service *Service
}

func (s *SpotifyClient) Provider() string {
	return "spotify"
}

func (s *SpotifyClient) SetService(service *Service) {
	s.Service = service
}

func (s *SpotifyClient) Authenticate(cfg *HandlerConfig) (handler Handler, err error) {
	if cfg.Username == "" || cfg.Password == "" || cfg.BlobPath == "" {
		return nil, fmt.Errorf("spotify: must provide username, password, and path to file for storing auth blob")
	}
	username := cfg.Username
	password := cfg.Password
	deviceName := cfg.DeviceName
	blobPath := cfg.BlobPath

	if _, err := os.Stat(blobPath); !os.IsNotExist(err) {
		blobBytes, err := ioutil.ReadFile(blobPath)
		if err != nil {
			return nil, err
		}
		session, err := librespot.LoginSaved(username, blobBytes, deviceName)
		if err != nil {
			return nil, err
		}
		s.Session = session
		return s, nil
	}

	session, err := librespot.Login(username, password, deviceName)
	if err != nil {
		return nil, err
	}
	s.Session = session

	err = ioutil.WriteFile(blobPath, session.ReusableAuthBlob(), 0600)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// NewSpotify authenticates to Spotify and returns a Spotify session
func NewSpotify(username, password, deviceName, blobPath string) (*SpotifyClient, error) {
	if username == "" || password == "" || blobPath == "" {
		return nil, fmt.Errorf("must provide username, password, and filepath to auth blob")
	}

	//Don't return unless blob authentication is successful, we'll reauth if it fails and only return an error then
	if _, err := os.Stat(blobPath); !os.IsNotExist(err) { //File exists
		blobBytes, err := ioutil.ReadFile(blobPath)
		if err == nil {
			session, err := librespot.LoginSaved(username, blobBytes, deviceName)
			if err == nil {
				return &SpotifyClient{Session: session}, nil
			}
		}
	}

	session, err := librespot.Login(username, password, deviceName)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(blobPath, session.ReusableAuthBlob(), 0600)
	if err != nil {
		return nil, err
	}

	return &SpotifyClient{Session: session}, nil
}

func (s *SpotifyClient) mercuryGet(url string) []byte {
	m := s.Session.Mercury()
	done := make(chan []byte)
	go m.Request(mercury.Request{
		Method:  "GET",
		Uri:     url,
		Payload: [][]byte{},
	}, func(res mercury.Response) {
		done <- res.CombinePayload()
	})

	result := <-done
	return result
}

func (s *SpotifyClient) mercuryGetJson(url string, result interface{}) (err error) {
	data := s.mercuryGet(url)
	//Trace.Printf("Spotify Mercury JSON: %s\n", data)
	err = json.Unmarshal(data, result)
	return
}

// Creator gets an artist object from Spotify
func (s *SpotifyClient) Creator(creatorID string) (creator *ObjectCreator, err error) {
	s.Lock()
	defer s.Unlock()

	spotCreator, err := s.Session.Mercury().GetArtist(utils.Base62ToHex(creatorID))
	if err != nil {
		return nil, err
	}
	biography := ""
	if len(spotCreator.Biography) > 0 {
		spotBios := spotCreator.Biography
		for i := 0; i < len(spotBios); i++ {
			if spotBios[i] != nil && spotBios[i].Text != nil {
				biography += *spotBios[i].Text
				if i == len(spotBios)-1 {
					biography += "\n\n"
				}
			}
		}
		if biography != "" {
			biography = s.ReplaceURI(biography)
		}
	}
	topTracks := make([]*Object, 0)
	for i := 0; i < len(spotCreator.TopTrack); i++ {
		for j := 0; j < len(spotCreator.TopTrack[i].Track); j++ {
			topTrack := spotCreator.TopTrack[i].Track[j]
			objStream := &ObjectStream{URI: "spotify:track:" + gid2Id(topTrack.Gid)}
			obj := &Object{
				URI: "spotify:track:" + gid2Id(topTrack.Gid),
				Type:     "stream",
				Provider: "spotify",
				Object: &json.RawMessage{},
			}
			obj.Object.UnmarshalJSON(objStream.JSON())
			topTracks = append(topTracks, obj)
		}
	}
	albums := make([]*Object, 0)
	for i := 0; i < len(spotCreator.AlbumGroup); i++ {
		for j := 0; j < len(spotCreator.AlbumGroup[i].Album); j++ {
			spotAlbum := spotCreator.AlbumGroup[i].Album[j]
			objAlbum := &ObjectAlbum{URI: "spotify:album:" + gid2Id(spotAlbum.Gid)}
			obj := &Object{
				URI: "spotify:album:" + gid2Id(spotAlbum.Gid),
				Type:     "album",
				Provider: "spotify",
				Object: &json.RawMessage{},
			}
			obj.Object.UnmarshalJSON(objAlbum.JSON())
			albums = append(albums, obj)
		}
	}
	appearances := make([]*Object, 0)
	for i := 0; i < len(spotCreator.AppearsOnGroup); i++ {
		for j := 0; j < len(spotCreator.AppearsOnGroup[i].Album); j++ {
			spotAlbum := spotCreator.AppearsOnGroup[i].Album[j]
			objAlbum := &ObjectAlbum{URI: "spotify:album:" + gid2Id(spotAlbum.Gid)}
			obj := &Object{
				URI: "spotify:album:" + gid2Id(spotAlbum.Gid),
				Type:     "album",
				Provider: "spotify",
				Object: &json.RawMessage{},
			}
			obj.Object.UnmarshalJSON(objAlbum.JSON())
			appearances = append(appearances, obj)
		}
	}
	singles := make([]*Object, 0)
	for i := 0; i < len(spotCreator.SingleGroup); i++ {
		for j := 0; j < len(spotCreator.SingleGroup[i].Album); j++ {
			spotAlbum := spotCreator.SingleGroup[i].Album[j]
			objAlbum := &ObjectAlbum{URI: "spotify:album:" + gid2Id(spotAlbum.Gid)}
			obj := &Object{
				URI: "spotify:album:" + gid2Id(spotAlbum.Gid),
				Type:     "album",
				Provider: "spotify",
				Object: &json.RawMessage{},
			}
			obj.Object.UnmarshalJSON(objAlbum.JSON())
			singles = append(singles, obj)
		}
	}
	related := make([]*Object, len(spotCreator.Related))
	for i := 0; i < len(related); i++ {
		relatedCreator := spotCreator.Related[i]
		objCreator := &ObjectCreator{
			URI:  "spotify:artist:" + gid2Id(relatedCreator.Gid),
			Name: *relatedCreator.Name,
		}
		related[i] = &Object{
			URI:  "spotify:artist:" + gid2Id(relatedCreator.Gid),
			Type:     "creator",
			Provider: "spotify",
			Object: &json.RawMessage{},
		}
		related[i].Object.UnmarshalJSON(objCreator.JSON())
	}
	creator = &ObjectCreator{
		URI:         "spotify:artist:" + creatorID,
		Name:        *spotCreator.Name,
		Description: biography,
		Genres:      spotCreator.Genre,
		TopStreams:  topTracks,
		Albums:      albums,
		Appearances: appearances,
		Singles:     singles,
		Related:     related,
	}
	return
}

// Album gets an album object from Spotify
func (s *SpotifyClient) Album(albumID string) (album *ObjectAlbum, err error) {
	s.Lock()
	defer s.Unlock()

	spotAlbum, err := s.Session.Mercury().GetAlbum(utils.Base62ToHex(albumID))
	if err != nil {
		return nil, err
	}
	creators := make([]*Object, len(spotAlbum.Artist))
	for i := 0; i < len(creators); i++ {
		objCreator := &ObjectCreator{
			URI:  "spotify:artist:" + gid2Id(spotAlbum.Artist[i].Gid),
			Name: *spotAlbum.Artist[i].Name,
		}
		creators[i] = &Object{
			URI:  "spotify:artist:" + gid2Id(spotAlbum.Artist[i].Gid),
			Type:     "creator",
			Provider: "spotify",
			Object: &json.RawMessage{},
		}
		creators[i].Object.UnmarshalJSON(objCreator.JSON())
	}
	discs := make([]*ObjectDisc, len(spotAlbum.Disc))
	for i := 0; i < len(discs); i++ {
		discStreams := make([]*Object, len(spotAlbum.Disc[i].Track))
		for j := 0; j < len(spotAlbum.Disc[i].Track); j++ {
			spotTrack := spotAlbum.Disc[i].Track[j]
			objStream := &ObjectStream{
				URI: "spotify:track:" + gid2Id(spotTrack.Gid),
				ID:  gid2Id(spotTrack.Gid),
			}
			discStreams[j] = &Object{
				URI: "spotify:track:" + gid2Id(spotTrack.Gid),
				Type:     "stream",
				Provider: "spotify",
				Object: &json.RawMessage{},
			}
			discStreams[j].Object.UnmarshalJSON(objStream.JSON())
		}
		discs[i] = &ObjectDisc{
			Streams: discStreams,
		}
		if spotAlbum.Disc[i].Number != nil {
			discs[i].Disc = int(*spotAlbum.Disc[i].Number)
		}
		if spotAlbum.Disc[i].Name != nil {
			discs[i].Name = *spotAlbum.Disc[i].Name
		}
	}
	copyrights := make([]string, len(spotAlbum.Copyright))
	for i := 0; i < len(copyrights); i++ {
		copyrights[i] = *spotAlbum.Copyright[i].Text
	}
	album = &ObjectAlbum{
		URI:        "spotify:album:" + albumID,
		Name:       *spotAlbum.Name,
		Creators:   creators,
		Discs:      discs,
		Copyrights: copyrights,
	}
	if spotAlbum.Label != nil {
		album.Label = *spotAlbum.Label
	}
	if spotAlbum.Date != nil {
		date := *spotAlbum.Date
		dateTime := ""
		if date.Year != nil && date.Month != nil {
			dateTime = fmt.Sprintf("%d-%d", *date.Year, *date.Month)
			if date.Day != nil {
				dateTime += fmt.Sprintf("-%d", *date.Day)
			}
		}
		if date.Hour != nil && date.Minute != nil {
			if dateTime != "" {
				dateTime += " "
			}
			dateTime += fmt.Sprintf("%d:%d", *date.Hour, *date.Minute)
		}
		album.DateTime = dateTime
	}
	return
}

// Stream gets a stream object from Spotify
func (s *SpotifyClient) Stream(trackID string) (stream *ObjectStream, err error) {
	s.Lock()
	defer s.Unlock()

	spotTrack, err := s.Session.Mercury().GetTrack(utils.Base62ToHex(trackID))
	if err != nil {
		return nil, err
	}
	creators := make([]*Object, len(spotTrack.Artist))
	for i := 0; i < len(creators); i++ {
		creator := spotTrack.Artist[i]
		objCreator := &ObjectCreator{
			Name: *creator.Name,
			URI:  "spotify:artist:" + gid2Id(creator.Gid),
		}
		creators[i] = &Object{
			URI:  "spotify:artist:" + gid2Id(creator.Gid),
			Type:     "creator",
			Provider: "spotify",
			Object: &json.RawMessage{},
		}
		creators[i].Object.UnmarshalJSON(objCreator.JSON())
	}
	formats := make([]*ObjectFormat, 0)
	formatList := s.FormatList()
	for i := 0; i < len(formatList); i++ {
		for j := 0; j < len(spotTrack.File); j++ {
			switch i {
			case 0:
				if *spotTrack.File[j].Format == Spotify.AudioFile_OGG_VORBIS_320 {
					formats = append(formats, formatList[i])
					formats[len(formats)-1].File = spotTrack.File[j]
					break
				}
			case 1:
				if *spotTrack.File[j].Format == Spotify.AudioFile_OGG_VORBIS_320 {
					formats = append(formats, formatList[i])
					formats[len(formats)-1].File = spotTrack.File[j]
					break
				}
			case 2:
				if *spotTrack.File[j].Format == Spotify.AudioFile_OGG_VORBIS_320 {
					formats = append(formats, formatList[i])
					formats[len(formats)-1].File = spotTrack.File[j]
					break
				}
			case 3:
				if *spotTrack.File[j].Format == Spotify.AudioFile_OGG_VORBIS_320 {
					formats = append(formats, formatList[i])
					formats[len(formats)-1].File = spotTrack.File[j]
					break
				}
			case 4:
				if *spotTrack.File[j].Format == Spotify.AudioFile_OGG_VORBIS_320 {
					formats = append(formats, formatList[i])
					formats[len(formats)-1].File = spotTrack.File[j]
					break
				}
			case 5:
				if *spotTrack.File[j].Format == Spotify.AudioFile_OGG_VORBIS_320 {
					formats = append(formats, formatList[i])
					formats[len(formats)-1].File = spotTrack.File[j]
					break
				}
			case 6:
				if *spotTrack.File[j].Format == Spotify.AudioFile_OGG_VORBIS_320 {
					formats = append(formats, formatList[i])
					formats[len(formats)-1].File = spotTrack.File[j]
					break
				}
			}
		}
	}
	stream = &ObjectStream{
		Provider: s.Provider(),
		URI:      "spotify:track:" + trackID,
		ID:       trackID,
		Name:     *spotTrack.Name,
		Creators: creators,
		Duration: int64(*spotTrack.Duration) / 1000,
		Formats:  formats,
	}
	if spotTrack.Album != nil {
		album := &Object{
			URI:  "spotify:album:" + gid2Id(spotTrack.Album.Gid),
			Type:     "album",
			Provider: "spotify",
			Object: &json.RawMessage{},
		}
		stream.Album = album
		albumObj := &ObjectAlbum{
			Name: *spotTrack.Album.Name,
			URI:  "spotify:album:" + gid2Id(spotTrack.Album.Gid),
		}
		stream.Album.Object.UnmarshalJSON(albumObj.JSON())
	}
	return
}

// Format gets a format object from a Spotify stream object
func (s *SpotifyClient) StreamFormat(w http.ResponseWriter, r *http.Request, stream *ObjectStream, format int) (err error) {
	objFormat := stream.GetFormat(format)
	if objFormat == nil {
		return fmt.Errorf("spotify: unknown format %d for stream %s", format, stream.ID)
	}
	file, ok := objFormat.File.(*Spotify.AudioFile)
	if !ok || file == nil {
		stream, err = s.Stream(stream.ID)
		if err != nil {
			return fmt.Errorf("spotify: failed to get stream %s: %v", stream.ID, err)
		}
		objFormat = stream.GetFormat(format)
		if objFormat == nil {
			return fmt.Errorf("spotify: unknown format %d for stream %s after resyncing", format, stream.ID)
		}
		file = objFormat.File.(*Spotify.AudioFile)
		if file == nil {
			return fmt.Errorf("spotify: unknown file for format %d from stream %s after resyncing", format, stream.ID)
		}
	}
	streamer, err := s.Session.Player().LoadTrackWithIdAndFormat(file.FileId, file.GetFormat(), id2Gid(stream.ID))
	if err != nil {
		return fmt.Errorf("spotify: failed to load track for stream %s: %v", stream.ID, err)
	}
	w.Header().Set("Content-Type", "audio/ogg")
	http.ServeContent(w, r, stream.ID, time.Time{}, streamer)
	return nil
}

// FormatList returns all the possible formats as templates ordered from best to worst
func (s *SpotifyClient) FormatList() (formats []*ObjectFormat) {
	formats = []*ObjectFormat{
		&ObjectFormat{
			ID: 0,
			Name: "Very High OGG",
			Format:     "ogg",
			Codec:      "vorbis",
			BitRate:    320000,
			BitDepth:   16,
			SampleRate: 44100,
		},
		&ObjectFormat{
			ID: 1,
			Name: "Very High MP3",
			Format:     "mp3",
			Codec:      "mp3",
			BitRate:    320000,
			BitDepth:   16,
			SampleRate: 44100,
		},
		&ObjectFormat{
			ID: 2,
			Name: "High MP3",
			Format:     "mp3",
			Codec:      "mp3",
			BitRate:    256000,
			BitDepth:   16,
			SampleRate: 44100,
		},
		&ObjectFormat{
			ID: 3,
			Name: "Normal OGG",
			Format:     "ogg",
			Codec:      "vorbis",
			BitRate:    160000,
			BitDepth:   16,
			SampleRate: 44100,
		},
		&ObjectFormat{
			ID: 4,
			Name: "Normal MP3",
			Format:     "mp3",
			Codec:      "mp3",
			BitRate:    160000,
			BitDepth:   16,
			SampleRate: 44100,
		},
		&ObjectFormat{
			ID: 5,
			Name: "Low OGG",
			Format:     "ogg",
			Codec:      "vorbis",
			BitRate:    96000,
			BitDepth:   16,
			SampleRate: 44100,
		},
		&ObjectFormat{
			ID: 6,
			Name: "Low MP3",
			Format:     "mp3",
			Codec:      "mp3",
			BitRate:    96000,
			BitDepth:   16,
			SampleRate: 44100,
		},
	}
	return
}

// Search returns the results matching a given query
func (s *SpotifyClient) Search(query string) (results *ObjectSearchResults, err error) {
	s.Lock()
	defer s.Unlock()

	searchResponse, err := s.Session.Mercury().Search(query, 10, s.Session.Country(), s.Session.Username())
	if err != nil {
		return nil, err
	}

	results = &ObjectSearchResults{}
	results.Query = query
	if searchResponse.Results.Artists.Total > 0 {
		artists := searchResponse.Results.Artists.Hits
		for i := 0; i < len(artists); i++ {
			creator := &ObjectCreator{
				Name: artists[i].Name,
				URI:  artists[i].Uri,
			}
			obj := &Object{URI: creator.URI, Type: "creator", Provider: "spotify", Object: &json.RawMessage{}}
			obj.Object.UnmarshalJSON(creator.JSON())

			results.Creators = append(results.Creators, obj)
		}
	}
	if searchResponse.Results.Albums.Total > 0 {
		albums := searchResponse.Results.Albums.Hits
		for i := 0; i < len(albums); i++ {
			album := &ObjectAlbum{
				Name: albums[i].Name,
				URI:  albums[i].Uri,
			}
			obj := &Object{URI: album.URI, Type: "album", Provider: "spotify", Object: &json.RawMessage{}}
			obj.Object.UnmarshalJSON(album.JSON())

			results.Albums = append(results.Albums, obj)
		}
	}
	if searchResponse.Results.Tracks.Total > 0 {
		tracks := searchResponse.Results.Tracks.Hits
		for i := 0; i < len(tracks); i++ {
			stream := &ObjectStream{Name: tracks[i].Name}
			stream.URI = tracks[i].Uri
			for _, artist := range tracks[i].Artists {
				objCreator := &ObjectCreator{Name: artist.Name, URI: artist.Uri}
				obj := &Object{URI: artist.Uri, Type: "creator", Provider: "spotify", Object: &json.RawMessage{}}
				obj.Object.UnmarshalJSON(objCreator.JSON())
				stream.Creators = append(stream.Creators, obj)
			}
			stream.Album = &Object{URI: tracks[i].Album.Uri, Type: "album", Provider: "spotify", Object: &json.RawMessage{}}
			objAlbum := &ObjectAlbum{Name: tracks[i].Album.Name, URI: tracks[i].Album.Uri}
			stream.Album.Object.UnmarshalJSON(objAlbum.JSON())
			stream.Artworks = []*ObjectArtwork{
				&ObjectArtwork{
					URL: tracks[i].Image,
				},
			}
			stream.Duration = int64(tracks[i].Duration) / 1000

			objStream := &Object{URI: stream.URI, Type: "stream", Provider: "spotify", Object: &json.RawMessage{}}
			objStream.Object.UnmarshalJSON(stream.JSON())
			results.Streams = append(results.Streams, objStream)
		}
	}
	/*if searchResponse.Results.Playlists.Total > 0 {
		playlists := searchResponse.Results.Playlists.Hits
		for i := 0; i < len(playlists); i++ {
			playlist := &ObjectPlaylist{
				Name: playlists[i].Name,
				URI:  playlists[i].Uri,
			}

			results.Playlists = append(results.Playlists, &Object{Type: "playlist", Provider: "spotify", Object: playlist})
		}
	}*/

	return
}

// gid2Id converts a given GID to an ID
func gid2Id(gid []byte) (id string) {
	dstId := make([]byte, base64.StdEncoding.DecodedLen(len(gid)))
	_, err := base64.StdEncoding.Decode(dstId, gid)
	if err != nil {
		//Error.Printf("Gid2Id: %v str(%v) err(%s)\n", gid, string(gid), err.Error())
		id = utils.ConvertTo62(gid)
		return
	}
	id = utils.ConvertTo62(dstId)
	return
}

// id2Gid converts a given ID to a GID
func id2Gid(id string) (gid []byte) {
	dstId := make([]byte, base64.StdEncoding.EncodedLen(len(id)))
	base64.StdEncoding.Encode(dstId, []byte(id))
	if len(dstId) > 0 {
		id = string(dstId)
	}
	gid = utils.Convert62(id)
	return
}

type SpotifyLyrics struct {
	Colors *Colors
	HasVocalRemoval bool
	Lyrics *SpotifyLyricsInner
}

type Colors struct {
	Background json.Number
	HighlightText json.Number
	Text json.Number
}

type SpotifyLyricsInner struct {
	FullscreenAction string
	IsDenseTypeface bool
	IsRtlLanguage bool
	Language string
	Lines []*Line
	Provider string
	ProviderDisplayName string
	ProviderLyricsID string
	SyncLyricsURI string
	SyncType json.Number //0=Unsynced,1=LineSynced
}

type Line struct {
	StartTimeMs string
	EndTimeMs string
	Words string
	Syllables []json.RawMessage
}

func (s *SpotifyClient) Transcribe(stream *ObjectStream) (err error) {
	s.Lock()
	defer s.Unlock()

	uri := fmt.Sprintf("hm://color-lyrics/v2/track/%s", stream.URI)
	lyrics := &SpotifyLyrics{}
	err = s.mercuryGetJson(uri, lyrics)
	//Trace.Printf("Spotify Lyrics object: %v\n", lyrics)
	return
}

// ReplaceURI replaces all instances of a URI with a libremedia-acceptable URI
func (s *SpotifyClient) ReplaceURI(text string) string {
	return spoturire.ReplaceAllStringFunc(text, spotifyReplaceURI)
}

func spotifyReplaceURI(link string) string {
	fmt.Println("Testing string: " + link)
	match := spoturire.FindAllStringSubmatch(link, 1)
	if len(match) > 0 {
		typed := match[0][1]
		id := match[0][2]
		name := match[0][3]
		switch typed {
		case "album":
			typed = "album"
		case "artist":
			typed = "creator"
		case "search":
			//TODO: Handle search URIs
			return "[" + name + "](/search?q=" + name + ")"
		}
		return "[" + name + "](/" + typed + "?uri=spotify:" + typed + ":" + id + ")"
	}
	return link
}