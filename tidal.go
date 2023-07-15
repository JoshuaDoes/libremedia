package main

/*
technical_names = {
    'eac3': 'E-AC-3 JOC (Dolby Digital Plus with Dolby Atmos, with 5.1 bed)',
    'mha1': 'MPEG-H 3D Audio (Sony 360 Reality Audio)',
    'ac4': 'AC-4 IMS (Dolby AC-4 with Dolby Atmos immersive stereo)',
    'mqa': 'MQA (Master Quality Authenticated) in FLAC container',
    'flac': 'FLAC (Free Lossless Audio Codec)',
    'alac': 'ALAC (Apple Lossless Audio Codec)',
    'mp4a.40.2': 'AAC 320 (Advanced Audio Coding) with a bitrate of 320kb/s',
    'mp4a.40.5': 'AAC 96 (Advanced Audio Coding) with a bitrate of 96kb/s'
}
*/

import (
	"context"
	jsontwo "encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	//Golang repos
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	//GitHub repos
	"github.com/JoshuaDoes/json"
	"github.com/dsoprea/go-utility/filesystem"
)

const (
	tidalAuth        = "https://auth.tidal.com/v1/oauth2"
	tidalAPI         = "https://api.tidal.com/v1/"
	tidalImgURL      = "https://resources.tidal.com/images/%s/%dx%d.jpg"
	tidalVidURL      = "https://resources.tidal.com/videos/%s/%dx%d.mp4"
	tidalTracksItems = "100" //max 100
	tidalSearchItems = "10"  //max 100
)

var (
	tidalSizesCreator = []int{160, 320, 480, 750}
	tidalSizesAlbum   = []int{80, 160, 320, 640, 1280}

	tidalurire = regexp.MustCompile(`\[wimpLink (.*?)="(.*?)"\](.*?)\[/wimpLink\]`)
)

// TidalError holds an error from Tidal
type TidalError struct {
	Status    int    `json:"status"`
	SubStatus int    `json:"sub_status"`
	ErrorType string `json:"error"`
	ErrorMsg  string `json:"error_description"`
}

// Error returns an error
func (terr *TidalError) Error() error {
	return fmt.Errorf("%d:%d %s: %s", terr.Status, terr.SubStatus, terr.ErrorType, terr.ErrorMsg)
}

// TidalClient holds a Tidal client
type TidalClient struct {
	sync.Mutex

	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`

	HTTP    *oauth2.Transport `json:"-"`
	Auth    *TidalDeviceCode  `json:"auth"`
	Service *Service          `json:"-"`
}

// Provider returns the name of this provider
func (t *TidalClient) Provider() string {
	return "tidal"
}

// SetService sets the global libremedia service for this provider
func (t *TidalClient) SetService(service *Service) {
	t.Service = service
}

// Authenticate authenticates this Tidal session
func (t *TidalClient) Authenticate(cfg *HandlerConfig) (handler Handler, err error) {
	if cfg.Username == "" || cfg.Password == "" || cfg.BlobPath == "" {
		return nil, fmt.Errorf("tidal: must provide username, password, and path to file for storing auth blob")
	}
	id := cfg.Username
	secret := cfg.Password
	blobPath := cfg.BlobPath

	t, err = NewTidalBlob(blobPath)
	if err != nil {
		t, err = NewTidal(id, secret)
		if err != nil {
			return nil, fmt.Errorf("tidal: unable to receive device code for account linking: %v", err)
		}
		os.Stderr.Write([]byte("Please link your Tidal account to continue!\n- https://" + t.Auth.VerificationURIComplete + "\n"))
		err = t.WaitForAuth()
		if err != nil {
			return nil, fmt.Errorf("tidal: unable to receive auth token: %v", err)
		}
	} else {
		if t.NeedsAuth() {
			os.Remove(blobPath)
			err = t.NewDeviceCode()
			if err != nil {
				return nil, fmt.Errorf("tidal: unable to receive device code for account linking: %v", err)
			}
			os.Stderr.Write([]byte("Please link your Tidal account to continue!\n- https://" + t.Auth.VerificationURIComplete + "\n"))
			err = t.WaitForAuth()
			if err != nil {
				return nil, fmt.Errorf("tidal: unable to receive auth token: %v", err)
			}
		}
	}

	t.SaveBlob(blobPath) //Save the new token blob for future restarts
	os.Stderr.Write([]byte("Authenticated to Tidal: Welcome\n"))
	return t, nil
}

// Get attempts to roundtrip an authenticated request to Tidal
func (t *TidalClient) Get(endpoint string, query url.Values) (*http.Response, error) {
	t.Lock()
	defer t.Unlock()

	if query == nil {
		query = url.Values{}
	}
	query.Add("countryCode", t.Auth.CountryCode)
	req, err := http.NewRequest("GET", tidalAPI+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = query.Encode()
	resp, err := t.HTTP.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetJSON gets an authenticated JSON resource from a Tidal endpoint and writes it to a target interface
func (t *TidalClient) GetJSON(endpoint string, query url.Values, target interface{}) error {
	resp, err := t.Get(endpoint, query)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//Trace.Println("Tidal:", resp.Status, endpoint, "\n", string(body))
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}
	if target != nil {
		return json.Unmarshal(body, target)
	}
	return nil
}

// TidalArtist holds a Tidal artist
type TidalArtist struct {
	ID            jsontwo.Number `json:"id"`
	Name          string         `json:"name"`
	ArtistTypes   []string       `json:"artistTypes,omitempty"`
	Albums        []TidalAlbum   `json:"albums,omitempty"`
	EPsAndSingles []TidalAlbum   `json:"epsandsingles,omitempty"`
	Picture       string         `json:"picture,omitempty"`
}

// TidalArtistAlbums holds a Tidal artist's album list
type TidalArtistAlbums struct {
	Limit              int          `json:"limit"`
	Offset             int          `json:"offset"`
	TotalNumberOfItems int          `json:"totalNumberOfItems"`
	Items              []TidalAlbum `json:"items"`
}

// TidalTracks holds a Tidal track list
type TidalTracks struct {
	Limit              int          `json:"limit"`
	Offset             int          `json:"offset"`
	TotalNumberOfItems int          `json:"totalNumberOfItems"`
	Items              []TidalTrack `json:"items"`
}

// TidalBio holds a Tidal artist's biography
type TidalBio struct {
	Text string `json:"text"`
}

// Creator gets an artist object from Tidal
func (t *TidalClient) Creator(creatorID string) (creator *ObjectCreator, err error) {
	tCreator := &TidalArtist{}
	err = t.GetJSON("artists/"+creatorID, nil, &tCreator)
	if err != nil {
		return nil, err
	}
	tBio := &TidalBio{}
	_ = t.GetJSON("artists/"+creatorID+"/bio", nil, &tBio)
	if tBio.Text != "" {
		tBio.Text = t.ReplaceURI(tBio.Text)
	}
	tTopTracks := &TidalTracks{}
	tracksFilter := url.Values{}
	tracksFilter.Set("limit", tidalTracksItems)
	err = t.GetJSON("artists/"+creatorID+"/toptracks", tracksFilter, &tTopTracks)
	if err != nil {
		return nil, err
	}
	topTracks := make([]*Object, len(tTopTracks.Items))
	for i := 0; i < len(topTracks); i++ {
		tTrack := tTopTracks.Items[i]
		tCreators := make([]*Object, len(tTrack.Artists))
		for j := 0; j < len(tCreators); j++ {
			objCreator := &ObjectCreator{
				URI:  "tidal:artist:" + tTrack.Artists[j].ID.String(),
				Name: tTrack.Artists[j].Name,
			}
			tCreators[j] = &Object{
				URI:      "tidal:artist:" + tTrack.Artists[j].ID.String(),
				Type:     "creator",
				Provider: "tidal",
				Object:   &jsontwo.RawMessage{},
			}
			tCreators[j].Object.UnmarshalJSON(objCreator.JSON())
		}
		trackNum, err := tTrack.TrackNumber.Int64()
		if err != nil {
			return nil, err
		}
		duration, err := tTrack.Duration.Int64()
		if err != nil {
			return nil, err
		}
		objStream := &ObjectStream{
			URI:      "tidal:track:" + tTopTracks.Items[i].ID.String(),
			ID:       tTopTracks.Items[i].ID.String(),
			Name:     tTopTracks.Items[i].Title,
			Track:    int(trackNum),
			Duration: duration,
			Creators: tCreators,
			Album: &Object{
				URI:      "tidal:album:" + tTrack.Album.ID.String(),
				Type:     "album",
				Provider: "tidal",
				Object:   &jsontwo.RawMessage{},
			},
		}
		objAlbum := &ObjectAlbum{
			URI:  "tidal:album:" + tTrack.Album.ID.String(),
			Name: tTrack.Album.Title,
		}
		objStream.Album.Object.UnmarshalJSON(objAlbum.JSON())
		topTracks[i] = &Object{
			URI:      "tidal:track:" + tTopTracks.Items[i].ID.String(),
			Type:     "stream",
			Provider: "tidal",
			Object:   &jsontwo.RawMessage{},
		}
		topTracks[i].Object.UnmarshalJSON(objStream.JSON())
	}
	tAlbums := TidalArtistAlbums{}
	albumFilter := url.Values{}
	albumFilter.Set("limit", tidalTracksItems)
	err = t.GetJSON("artists/"+creatorID+"/albums", albumFilter, &tAlbums)
	if err != nil {
		return nil, err
	}
	if len(tAlbums.Items) > 0 {
		tCreator.Albums = append(tCreator.Albums, tAlbums.Items...)
	}
	albums := make([]*Object, len(tCreator.Albums))
	for i := 0; i < len(albums); i++ {
		tAlbum := tCreator.Albums[i]
		objAlbum := &ObjectAlbum{
			URI:  "tidal:album:" + tAlbum.ID.String(),
			Name: tAlbum.Title,
		}
		albums[i] = &Object{
			URI:      "tidal:album:" + tAlbum.ID.String(),
			Type:     "album",
			Provider: "tidal",
			Object:   &jsontwo.RawMessage{},
		}
		albums[i].Object.UnmarshalJSON(objAlbum.JSON())
	}
	epsandsingles := TidalArtistAlbums{}
	albumFilter.Set("filter", "EPSANDSINGLES")
	err = t.GetJSON("artists/"+creatorID+"/albums", albumFilter, &epsandsingles)
	if err != nil {
		return nil, err
	}
	if len(epsandsingles.Items) > 0 {
		tCreator.EPsAndSingles = append(tCreator.EPsAndSingles, epsandsingles.Items...)
	}
	singles := make([]*Object, len(tCreator.EPsAndSingles))
	for i := 0; i < len(singles); i++ {
		tSingle := tCreator.EPsAndSingles[i]
		objAlbum := &ObjectAlbum{
			URI:  "tidal:album:" + tSingle.ID.String(),
			Name: tSingle.Title,
		}
		singles[i] = &Object{
			URI:      "tidal:album:" + tSingle.ID.String(),
			Type:     "album",
			Provider: "tidal",
			Object:   &jsontwo.RawMessage{},
		}
		singles[i].Object.UnmarshalJSON(objAlbum.JSON())
	}
	creator = &ObjectCreator{
		URI:         "tidal:artist:" + creatorID,
		Name:        tCreator.Name,
		Description: tBio.Text,
		TopStreams:  topTracks,
		Albums:      albums,
		Singles:     singles,
		Artworks:    t.ArtworkImg(tCreator.Picture, tidalSizesCreator),
	}
	return
}

// TidalAlbum holds a Tidal album
type TidalAlbum struct {
	ID              jsontwo.Number `json:"id"`
	Title           string         `json:"title"`
	Duration        jsontwo.Number `json:"duration,omitempty"`
	NumberOfTracks  jsontwo.Number `json:"numberOfTracks,omitempty"`
	NumberOfVideos  jsontwo.Number `json:"numberOfVideos,omitempty"`
	NumberOfVolumes jsontwo.Number `json:"numberOfVolumes,omitempty"`
	ReleaseDate     string         `json:"releaseDate,omitempty"`
	Copyright       string         `json:"copyright,omitempty"`
	Explicit        bool           `json:"explicit,omitempty"`
	AudioQuality    string         `json:"audioQuality,omitempty"`
	AudioModes      []string       `json:"audioModes,omitempty"` //usually just STEREO
	Artists         []TidalArtist  `json:"artists,omitempty"`
	Tracks          []TidalTrack   `json:"tracks,omitempty"`
	Cover           string         `json:"cover,omitempty"`      //An image cover for the album
	VideoCover      string         `json:"videoCover,omitempty"` //A video cover for the album
}

// TidalAlbumTracks holds a Tidal album's track list
type TidalAlbumTracks struct {
	Limit              int `json:"limit"`
	Offset             int `json:"offset"`
	TotalNumberOfItems int `json:"totalNumberOfItems"`
	Items              []struct {
		Item TidalTrack `json:"item"`
	} `json:"items"`
}

// Album gets an album object from Tidal
func (t *TidalClient) Album(albumID string) (album *ObjectAlbum, err error) {
	tAlbum := &TidalAlbum{}
	err = t.GetJSON("albums/"+albumID, nil, &tAlbum)
	if err != nil {
		return nil, err
	}
	creators := make([]*Object, len(tAlbum.Artists))
	for i := 0; i < len(creators); i++ {
		objCreator := &ObjectCreator{
			URI:  "tidal:artist:" + tAlbum.Artists[i].ID.String(),
			Name: tAlbum.Artists[i].Name,
		}
		creators[i] = &Object{
			URI:      "tidal:artist:" + tAlbum.Artists[i].ID.String(),
			Type:     "creator",
			Provider: "tidal",
			Object:   &jsontwo.RawMessage{},
		}
		creators[i].Object.UnmarshalJSON(objCreator.JSON())
	}
	tracks := TidalAlbumTracks{}
	albumFilter := url.Values{}
	albumFilter.Set("limit", tidalTracksItems)
	err = t.GetJSON("albums/"+albumID+"/items", albumFilter, &tracks)
	if err != nil {
		return nil, err
	}
	if len(tracks.Items) > 0 {
		for _, item := range tracks.Items {
			tAlbum.Tracks = append(tAlbum.Tracks, item.Item)
		}
	}
	discs := []*ObjectDisc{
		{
			Streams: make([]*Object, len(tAlbum.Tracks)),
			Disc:    1,
			Name:    tAlbum.Title,
		},
	}
	for i := 0; i < len(tAlbum.Tracks); i++ {
		tTrack := tAlbum.Tracks[i]
		tCreators := make([]*Object, len(tTrack.Artists))
		for j := 0; j < len(tCreators); j++ {
			objCreator := &ObjectCreator{
				URI:  "tidal:artist:" + tTrack.Artists[j].ID.String(),
				Name: tTrack.Artists[j].Name,
			}
			tCreators[j] = &Object{
				URI:      "tidal:artist:" + tTrack.Artists[j].ID.String(),
				Type:     "creator",
				Provider: "tidal",
				Object:   &jsontwo.RawMessage{},
			}
			tCreators[j].Object.UnmarshalJSON(objCreator.JSON())
		}
		trackNum, err := tTrack.TrackNumber.Int64()
		if err != nil {
			return nil, err
		}
		duration, err := tTrack.Duration.Int64()
		if err != nil {
			return nil, err
		}
		objStream := &ObjectStream{
			URI:      "tidal:track:" + tTrack.ID.String(),
			ID:       tTrack.ID.String(),
			Name:     tTrack.Title,
			Track:    int(trackNum),
			Duration: duration,
			Creators: tCreators,
			Album: &Object{
				URI:      "tidal:album:" + albumID,
				Type:     "album",
				Provider: "tidal",
				Object:   &jsontwo.RawMessage{},
			},
		}
		objAlbum := &ObjectAlbum{
			URI:      "tidal:album:" + albumID,
			Name:     tAlbum.Title,
			Creators: creators,
		}
		objStream.Album.Object.UnmarshalJSON(objAlbum.JSON())
		discs[0].Streams[i] = &Object{
			URI:      "tidal:track:" + tTrack.ID.String(),
			Type:     "stream",
			Provider: "tidal",
			Object:   &jsontwo.RawMessage{},
		}
		discs[0].Streams[i].Object.UnmarshalJSON(objStream.JSON())
	}
	album = &ObjectAlbum{
		URI:        "tidal:album:" + albumID,
		Name:       tAlbum.Title,
		Creators:   creators,
		Discs:      discs,
		Copyrights: []string{tAlbum.Copyright},
		Label:      tAlbum.Copyright,
		Artworks:   make([]*ObjectArtwork, 0),
		DateTime:   tAlbum.ReleaseDate,
		Explicit:   tAlbum.Explicit,
	}
	album.Artworks = append(album.Artworks, t.ArtworkImg(tAlbum.Cover, tidalSizesAlbum)...)
	album.Artworks = append(album.Artworks, t.ArtworkVid(tAlbum.VideoCover, tidalSizesAlbum)...)
	return
}

// TidalTrack holds a Tidal track
type TidalTrack struct {
	ID                   jsontwo.Number `json:"id"`
	Title                string         `json:"title"`
	Duration             jsontwo.Number `json:"duration"`
	ReplayGain           jsontwo.Number `json:"replayGain"`
	Peak                 jsontwo.Number `json:"peak"`
	AllowStreaming       bool           `json:"allowStreaming"`
	StreamReady          bool           `json:"streamReady"`
	AdSupportedStreaming bool           `json:"adSupportedStreaming"`
	StreamStartDate      string         `json:"streamStartDate"`
	PremiumStreamingOnly bool           `json:"premiumStreamingOnly"`
	TrackNumber          jsontwo.Number `json:"trackNumber,omitempty"`
	VolumeNumber         jsontwo.Number `json:"volumeNumber"`
	Version              string         `json:"version,omitempty"`
	Popularity           int            `json:"popularity,omitempty"`
	Copyright            string         `json:"copyright,omitempty"`
	TidalURL             string         `json:"url"`
	ISRC                 string         `json:"isrc"`
	Editable             bool           `json:"editable"`
	Explicit             bool           `json:"explicit,omitempty"`
	AudioQuality         string         `json:"audioQuality"`
	AudioModes           []string       `json:"audioModes"`
	MediaMetadata        struct {
		Tags []string `json:"tags,omitempty"`
	} `json:"mediaMetadata,omitempty"`
	Artist  TidalArtist   `json:"artist"`
	Artists []TidalArtist `json:"artists,omitempty"`
	Album   TidalAlbum    `json:"album,omitempty"`
	Mixes   struct {
		TrackMix string `json:"TRACK_MIX,omitempty"`
	} `json:"mixes,omitempty"`
}

// Stream gets a track object from Tidal
func (t *TidalClient) Stream(trackID string) (stream *ObjectStream, err error) {
	tTrack := &TidalTrack{}
	err = t.GetJSON("tracks/"+trackID, nil, &tTrack)
	if err != nil {
		return nil, err
	}
	creators := make([]*Object, len(tTrack.Artists))
	for i := 0; i < len(creators); i++ {
		objCreator := &ObjectCreator{
			URI:  "tidal:artist:" + tTrack.Artists[i].ID.String(),
			Name: tTrack.Artists[i].Name,
		}
		creators[i] = &Object{
			URI:      "tidal:artist:" + tTrack.Artists[i].ID.String(),
			Type:     "creator",
			Provider: "tidal",
			Object:   &jsontwo.RawMessage{},
		}
		creators[i].Object.UnmarshalJSON(objCreator.JSON())
	}
	formats := make([]*ObjectFormat, 0)
	formatList := t.FormatList()
	for i := 0; i < len(formatList); i++ {
		formats = append(formats, formatList[i]) //Assume the quality exists, bail down the ladder before playback
	}
	duration, err := tTrack.Duration.Int64()
	if err != nil {
		return nil, err
	}
	stream = &ObjectStream{
		Provider: t.Provider(),
		URI:      "tidal:track:" + trackID,
		ID:       trackID,
		Name:     tTrack.Title,
		Creators: creators,
		Album: &Object{
			URI:      "tidal:album:" + tTrack.Album.ID.String(),
			Type:     "album",
			Provider: "tidal",
			Object:   &jsontwo.RawMessage{},
		},
		Explicit: tTrack.Explicit,
		Duration: duration,
		Formats:  formats,
	}
	objAlbum := &ObjectAlbum{
		URI:  "tidal:album:" + tTrack.Album.ID.String(),
		Name: tTrack.Album.Title,
	}
	stream.Album.Object.UnmarshalJSON(objAlbum.JSON())
	return
}

// TidalVideo holds a Tidal video
type TidalVideo struct {
	Title    string         `json:"title"`
	ID       jsontwo.Number `json:"id"`
	Artists  []TidalArtist  `json:"artists,omitempty"`
	Album    TidalAlbum     `json:"album,omitempty"`
	Duration jsontwo.Number `json:"duration"`
}

// TidalPlaylist holds a Tidal playlist
type TidalPlaylist struct {
	Title          string         `json:"title"`
	UUID           string         `json:"uuid"`
	NumberOfTracks jsontwo.Number `json:"numberOfTracks"`
	NumberOfVideos jsontwo.Number `json:"numberOfVideos"`
	Creator        struct {
		ID jsontwo.Number `json:"id"`
	} `json:"creator"`
	Description     string         `json:"description"`
	Duration        jsontwo.Number `json:"duration"`
	LastUpdated     string         `json:"lastUpdated"`
	Created         string         `json:"created"`
	Type            string         `json:"type"` //USER
	PublicPlaylist  bool           `json:"publicPlaylist"`
	URL             string         `json:"url"`
	Image           string         `json:"image"`
	Popularity      jsontwo.Number `json:"popularity"`
	SquareImage     string         `json:"squareImage"`
	PromotedArtists []TidalArtist  `json:"promotedArtists"`
	LastItemAddedAt string         `json:"lastItemAddedAt"`
	Tracks          []TidalTrack   `json:"tracks,omitempty"`
}

// TidalPlaylistTracks holds a Tidal Playlist's track list
type TidalPlaylistTracks struct {
	Limit              int `json:"limit"`
	Offset             int `json:"offset"`
	TotalNumberOfItems int `json:"totalNumberOfItems"`
	Items              []struct {
		Item TidalTrack `json:"item"`
	} `json:"items"`
}

// GetPlaylist gets a playlist object from Tidal
func (t *TidalClient) GetPlaylist(playlistID string) (playlist *TidalPlaylist, err error) {
	err = t.GetJSON("playlists/"+playlistID, nil, &playlist)
	if err != nil {
		return nil, err
	}

	tracks := TidalPlaylistTracks{}
	err = t.GetJSON("playlists/"+playlistID+"/items", nil, &tracks)
	if err == nil && len(tracks.Items) > 0 {
		for _, item := range tracks.Items {
			playlist.Tracks = append(playlist.Tracks, item.Item)
		}
	}

	return playlist, err
}

// GetVideo gets a video object from Tidal
func (t *TidalClient) GetVideo(videoID string) (video *TidalVideo, err error) {
	err = t.GetJSON("videos/"+videoID, nil, &video)
	return video, err
}

// TidalLyrics holds lyric data for a Tidal track
type TidalLyrics struct {
	TrackID          jsontwo.Number `json:"trackId"`
	Provider         string         `json:"lyricsProvider"`
	ProviderTrackID  jsontwo.Number `json:"providerCommontrackId"`
	ProviderLyricsID jsontwo.Number `json:"providerLyricsId"`
	RightToLeft      bool           `json:"isRightToLeft"`
	Text             string         `json:"text"`
	Subtitles        string         `json:"subtitles"`
}

// Transcribe fills in a lyrics object from Tidal
func (t *TidalClient) Transcribe(stream *ObjectStream) (err error) {
	lyrics := &TidalLyrics{}
	uri := fmt.Sprintf("tracks/%s/lyrics", stream.ID)
	reqForm := url.Values{}
	reqForm.Set("deviceType", "BROWSER")
	reqForm.Set("locale", "en_US")
	err = t.GetJSON(uri, reqForm, &lyrics)
	if err != nil {
		return
	}
	//Trace.Printf("Tidal Lyrics object: %v\n", lyrics)

	text := lyrics.Text
	if text == "" {
		text = lyrics.Subtitles
		if text == "" {
			Error.Println("Tidal: Failed to find text or subtitles for " + stream.URI)
			return
		}
	}

	lines := strings.Split(text, "\n")
	objTranscript := &ObjectTranscript{
		RightToLeft:      lyrics.RightToLeft,
		Provider:         lyrics.Provider,
		ProviderLyricsID: lyrics.ProviderLyricsID.String(),
		ProviderTrackID:  lyrics.ProviderTrackID.String(),
		TimeSynced:       false,
		Lines:            make([]*ObjectTranscriptLine, 0),
	}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		var min, sec, ms = 0, 0, 0
		n, _ := fmt.Sscanf(line, "[%d:%d.%d]", &min, &sec, &ms)
		if n == 3 {
			objTranscript.TimeSynced = true
			startTimeMs := (min * 60 * 1000) + (sec * 1000) + (ms * 10)
			txt := strings.Split(line, "] ")[1]
			objTranscript.Lines = append(objTranscript.Lines, &ObjectTranscriptLine{
				StartTimeMs: startTimeMs,
				Text:        txt,
			})
		} else {
			objTranscript.Lines = append(objTranscript.Lines, &ObjectTranscriptLine{Text: line})
		}
	}
	if objTranscript.TimeSynced {
		Trace.Println("Tidal: Successfully time synced " + stream.URI)
	} else {
		Trace.Println("Tidal: Failed to find time sync data for " + stream.URI)
	}

	stream.Transcript = objTranscript
	return
}

// StreamFormat roundtrips a stream attempt for the given HTTP session
func (t *TidalClient) StreamFormat(w http.ResponseWriter, r *http.Request, stream *ObjectStream, format int) (err error) {
	objFormat := stream.GetFormat(format)
	if objFormat == nil {
		return fmt.Errorf("tidal: unknown format %d for stream %s", format, stream.ID)
	}
	manifest, err := t.GetAudioStream(stream.ID, objFormat.Name)
	if err != nil {
		return fmt.Errorf("tidal: unable to retrieve audio stream for stream %s at %s quality", stream.ID, objFormat.ID)
	}
	w.Header().Set("Content-Type", manifest.MimeType)
	for i := 0; i < len(manifest.URLs); i++ {
		req, err := http.NewRequest("GET", manifest.URLs[i], nil)
		if err != nil {
			jsonWriteErrorf(w, 500, "tidal: failed to create stream endpoint: %v", err)
			return err
		}
		resp, err := t.HTTP.RoundTrip(req)
		if err != nil {
			jsonWriteErrorf(w, 500, "tidal: failed to start stream: %v", err)
			return err
		}
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "broken pipe") || strings.Contains(errMsg, "connection reset") {
				return nil //The stream was successful, but interrupted
			}
			jsonWriteErrorf(w, 500, "tidal: failed to copy stream: %v", err)
			return err
		}
		streamer := rifs.NewSeekableBufferWithBytes(buf)
		http.ServeContent(w, r, stream.ID, time.Time{}, streamer)
		//_, err = io.Copy(w, resp.Body)
	}
	return nil
}

// GetAudioStream gets the stream for a given audio on Tidal
func (t *TidalClient) GetAudioStream(trackID, quality string) (manifest *TidalAudioManifest, err error) {
	reqForm := url.Values{}
	reqForm.Set("audioquality", quality)
	reqForm.Set("urlusagemode", "STREAM")
	reqForm.Set("assetpresentation", "FULL")
	err = t.GetJSON("tracks/"+trackID+"/urlpostpaywall", reqForm, &manifest)
	if err != nil {
		return nil, err
	}
	manifest.AudioQuality = quality
	manifest.Codec = "flac"
	manifest.MimeType = "audio/flac"
	return manifest, nil

	/*reqForm := url.Values{}
	reqForm.Set("audioquality", quality)
	reqForm.Set("playbackmode", "STREAM")
	reqForm.Set("assetpresentation", "FULL")
	reqForm.Set("prefetch", "false")

	err = t.GetJSON("tracks/"+trackID+"/playbackinfopostpaywall", reqForm, &stream)
	if err != nil {
		return nil, err
	}

	if needURL {
		decodedManifest, err := base64.StdEncoding.DecodeString(stream.ManifestBase64)
		if err != nil {
			return nil, fmt.Errorf("error decoding manifest of type %s: %v", stream.ManifestMimeType, err)
		}
		Trace.Printf("Tidal's decoded manifest for %s/%s:\n\n%s\n\n", trackID, quality, decodedManifest)

		if strings.Contains(stream.ManifestMimeType, "vnd.t.bt") {
			err = json.Unmarshal(decodedManifest, &stream.Manifest)
			if err != nil {
				return nil, fmt.Errorf("error unmarshalling vnd.t.bt manifest: %v\n\n%s", err, decodedManifest)
			}
		} else if stream.ManifestMimeType == "application/dash+xml" {
			//return nil, fmt.Errorf("not yet unmarshalling dash+xml:\n\n%s", string(decodedManifest))
			dashXML := &TidalAudioDashXML{}
			err = xml.Unmarshal([]byte(decodedManifest), &dashXML)
			if err != nil {
				return nil, fmt.Errorf("error unmarshalling dash+xml manifest: %v\n\n%s", err, decodedManifest)
			}
			stream.Manifest = &TidalAudioManifest{
				MimeType: dashXML.Period.AdaptationSet.MimeType,
				Codecs:   dashXML.Period.AdaptationSet.Representation.Codecs,
				URLs: []string{
					dashXML.Period.AdaptationSet.Representation.SegmentTemplate.Initialization,
					dashXML.Period.AdaptationSet.Representation.SegmentTemplate.Media,
				},
			}
		} else {
			return nil, fmt.Errorf("unsupported manifest type: %s", stream.ManifestMimeType)
		}
	}

	return stream, nil*/
}

// GetVideoStream gets the stream for a given video on Tidal
func (t *TidalClient) GetVideoStream(videoID, quality string) (stream *TidalVideoStream, err error) {
	err = t.GetJSON("videos/"+videoID+"/streamurl", nil, &stream)
	return stream, err
}

// TidalAudioStream holds a Tidal audio stream
type TidalAudioStream struct {
	TrackID            jsontwo.Number      `json:"trackId"`
	AssetPresentation  string              `json:"assetPresentation"`
	AudioMode          string              `json:"audioMode"`
	AudioQuality       string              `json:"audioQuality"`
	ManifestMimeType   string              `json:"manifestMimeType"`
	ManifestHash       string              `json:"manifestHash"`
	ManifestBase64     string              `json:"manifest"` //base64-encoded audio metadata
	Manifest           *TidalAudioManifest `json:"-"`
	AlbumReplayGain    jsontwo.Number      `json:"albumReplayGain"`
	AlbumPeakAmplitude jsontwo.Number      `json:"albumPeakAmplitude"`
	TrackReplayGain    jsontwo.Number      `json:"trackReplayGain"`
	TrackPeakAmplitude jsontwo.Number      `json:"trackPeakAmplitude"`
}

// TidalAudioManifest holds a Tidal audio stream's metadata manifest
type TidalAudioManifest struct {
	URLs               []string       `json:"urls"`
	TrackID            jsontwo.Number `json:"trackId"`
	AssetPresentation  string         `json:"assetPresentation"`
	AudioQuality       string         `json:"audioQuality"`
	AudioMode          string         `json:"audioMode"`
	StreamingSessionID string         `json:"streamingSessionId,omitempty"`
	Codec              string         `json:"codec"`
	SecurityType       string         `json:"securityType,omitempty"`
	SecurityToken      string         `json:"securityToken,omitempty"`
	MimeType           string         `json:"mimeType,omitempty"`
	Codecs             []string       `json:"codecs,omitempty"`
}

// TidalAudioDashXML was generated 2023-04-19 23:21:39 by https://xml-to-go.github.io/ in Ukraine.
type TidalAudioDashXML struct {
	XMLName                   xml.Name `xml:"MPD" json:"mpd,omitempty"`
	Text                      string   `xml:",chardata" json:"text,omitempty"`
	Xmlns                     string   `xml:"xmlns,attr" json:"xmlns,omitempty"`
	Xsi                       string   `xml:"xsi,attr" json:"xsi,omitempty"`
	Xlink                     string   `xml:"xlink,attr" json:"xlink,omitempty"`
	Cenc                      string   `xml:"cenc,attr" json:"cenc,omitempty"`
	SchemaLocation            string   `xml:"schemaLocation,attr" json:"schemalocation,omitempty"`
	Profiles                  string   `xml:"profiles,attr" json:"profiles,omitempty"`
	Type                      string   `xml:"type,attr" json:"type,omitempty"`
	MinBufferTime             string   `xml:"minBufferTime,attr" json:"minbuffertime,omitempty"`
	MediaPresentationDuration string   `xml:"mediaPresentationDuration,attr" json:"mediapresentationduration,omitempty"`
	Period                    struct {
		Text          string `xml:",chardata" json:"text,omitempty"`
		ID            string `xml:"id,attr" json:"id,omitempty"`
		AdaptationSet struct {
			Text             string `xml:",chardata" json:"text,omitempty"`
			ID               string `xml:"id,attr" json:"id,omitempty"`
			ContentType      string `xml:"contentType,attr" json:"contenttype,omitempty"`
			MimeType         string `xml:"mimeType,attr" json:"mimetype,omitempty"` //MimeType
			SegmentAlignment string `xml:"segmentAlignment,attr" json:"segmentalignment,omitempty"`
			Representation   struct {
				Text              string `xml:",chardata" json:"text,omitempty"`
				ID                string `xml:"id,attr" json:"id,omitempty"`
				Codecs            string `xml:"codecs,attr" json:"codecs,omitempty"` //Codecs
				Bandwidth         string `xml:"bandwidth,attr" json:"bandwidth,omitempty"`
				AudioSamplingRate string `xml:"audioSamplingRate,attr" json:"audiosamplingrate,omitempty"`
				SegmentTemplate   struct {
					Text            string `xml:",chardata" json:"text,omitempty"`
					Timescale       string `xml:"timescale,attr" json:"timescale,omitempty"`
					Initialization  string `xml:"initialization,attr" json:"initialization,omitempty"`
					Media           string `xml:"media,attr" json:"media,omitempty"` //URLs[0]
					StartNumber     string `xml:"startNumber,attr" json:"startnumber,omitempty"`
					SegmentTimeline struct {
						Text string `xml:",chardata" json:"text,omitempty"`
						S    []struct {
							Text string `xml:",chardata" json:"text,omitempty"`
							D    string `xml:"d,attr" json:"d,omitempty"`
							R    string `xml:"r,attr" json:"r,omitempty"`
						} `xml:"S" json:"s,omitempty"`
					} `xml:"SegmentTimeline" json:"segmenttimeline,omitempty"`
				} `xml:"SegmentTemplate" json:"segmenttemplate,omitempty"`
			} `xml:"Representation" json:"representation,omitempty"`
		} `xml:"AdaptationSet" json:"adaptationset,omitempty"`
	} `xml:"Period" json:"period,omitempty"`
}

// FormatList returns all available formats for Tidal
func (t *TidalClient) FormatList() (formats []*ObjectFormat) {
	formats = []*ObjectFormat{
		{
			ID:         0,
			Name:       "HI_RES",
			Format:     "flac",
			Codec:      "flac",
			BitRate:    9216000,
			BitDepth:   24,
			SampleRate: 96000,
		},
		{
			ID:         1,
			Name:       "LOSSLESS",
			Format:     "flac",
			Codec:      "flac",
			BitRate:    1411000,
			BitDepth:   24,
			SampleRate: 44100,
		},
		{
			ID:         2,
			Name:       "HIGH",
			Format:     "flac",
			Codec:      "flac",
			BitRate:    320000,
			BitDepth:   16,
			SampleRate: 44100,
		},
		{
			ID:         3,
			Name:       "LOW",
			Format:     "flac",
			Codec:      "flac",
			BitRate:    96000,
			BitDepth:   16,
			SampleRate: 44100,
		},
	}
	return
}

// TidalVideoStream holds a Tidal video stream
type TidalVideoStream struct {
	VideoID           jsontwo.Number      `json:"videoId"`
	StreamType        string              `json:"streamType"`
	AssetPresentation string              `json:"assetPresentation"`
	VideoQuality      string              `json:"videoQuality"`
	ManifestMimeType  string              `json:"manifestMimeType"`
	ManifestHash      string              `json:"manifestHash"`
	ManifestBase64    string              `json:"manifest"`
	Manifest          *TidalVideoManifest `json:"-"`
}

// TidalVideoManifest holds a Tidal video stream's metadata manifest
type TidalVideoManifest struct {
	MimeType       string   `json:"mimeType"`
	Codecs         string   `json:"codecs"`
	EncryptionType string   `json:"encryptionType"`
	URLs           []string `json:"urls"`
}

/*func tidalGenerateVideoFormats(videoID, topQuality string) []*ObjectFormat {
	//streamURLPrefix := "DOMAIN/v1/stream/tidal:video:" + videoID + "?format="

	formats := make([]*ObjectFormat, 0)
	possibleVideoQualities := []string{"LOW", "HIGH"}

	for i, quality := range possibleVideoQualities {
		tidalStream, err := t.GetVideoStream(videoID, quality)
		if err != nil {
			Error.Println(err)
			continue
		}

		tidalFormat := &ObjectFormat{
			ID: i,
			Name:     tidalStream.VideoQuality,
			URL:    tidalStream.Manifest.URLs[0], //streamURLPrefix + tidalStream.VideoQuality,
			Format: strings.Split(tidalStream.Manifest.MimeType, "/")[1],
			Codec:  tidalStream.Manifest.Codecs,
		}

		formats = append(formats, tidalFormat)
		if quality == topQuality {
			break
		}
	}

	return formats
}*/

// ArtworkImg retrieves a JPG artwork given a cover ID and acceptable size list
func (t *TidalClient) ArtworkImg(coverID string, sizes []int) []*ObjectArtwork {
	return t.Artwork(coverID, tidalImgURL, "jpg", sizes)
}

// ArtworkVid retrieves an MP4 artwork given a cover ID and acceptable size list
func (t *TidalClient) ArtworkVid(coverID string, sizes []int) []*ObjectArtwork {
	return t.Artwork(coverID, tidalVidURL, "mp4", sizes)
}

// Artwork handles the underlying artwork fetching
func (t *TidalClient) Artwork(coverID, coverURL, fileType string, sizes []int) (artworks []*ObjectArtwork) {
	if coverID == "" || fileType == "" {
		return nil
	}
	coverID = strings.ReplaceAll(coverID, "-", "/")
	artworks = make([]*ObjectArtwork, len(sizes))
	for i := 0; i < len(sizes); i++ {
		width := sizes[i]
		height := width
		url := fmt.Sprintf(coverURL, coverID, width, height)
		artworks[i] = NewObjArtwork(t.Provider(), fileType, url, width, height)
	}
	return artworks
}

// TidalSearchResults holds the Tidal results for a given search query
type TidalSearchResults struct {
	Artists struct {
		Limit              int           `json:"limit"`
		Offset             int           `json:"offset"`
		TotalNumberOfItems int           `json:"totalNumberOfItems"`
		Items              []TidalArtist `json:"items"`
	} `json:"artists"`
	Albums struct {
		Limit              int          `json:"limit"`
		Offset             int          `json:"offset"`
		TotalNumberOfItems int          `json:"totalNumberOfItems"`
		Items              []TidalAlbum `json:"items"`
	} `json:"albums"`
	Playlists struct {
		Limit              int             `json:"limit"`
		Offset             int             `json:"offset"`
		TotalNumberOfItems int             `json:"totalNumberOfItems"`
		Items              []TidalPlaylist `json:"items"`
	} `json:"playlists"`
	Tracks struct {
		Limit              int          `json:"limit"`
		Offset             int          `json:"offset"`
		TotalNumberOfItems int          `json:"totalNumberOfItems"`
		Items              []TidalTrack `json:"items"`
	} `json:"tracks"`
	Videos struct {
		Limit              int          `json:"limit"`
		Offset             int          `json:"offset"`
		TotalNumberOfItems int          `json:"totalNumberOfItems"`
		Items              []TidalVideo `json:"items"`
	} `json:"videos"`
}

// Search returns the results for a given search query
func (t *TidalClient) Search(query string) (results *ObjectSearchResults, err error) {
	results = &ObjectSearchResults{}
	searchResults := TidalSearchResults{}

	if t == nil {
		return results, nil
	}

	reqForm := url.Values{}
	reqForm.Set("query", query)
	reqForm.Set("limit", tidalSearchItems)

	//types := []string{"TRACKS", "ARTISTS", "ALBUMS", "PLAYLISTS"}
	types := []string{"TRACKS", "ARTISTS", "ALBUMS"}
	for i := 0; i < len(types); i++ {
		reqForm.Set("type", types[i])
		err = t.GetJSON("search", reqForm, &searchResults)
		if err != nil {
			return results, err
		}

		if searchResults.Artists.TotalNumberOfItems > 0 {
			artists := searchResults.Artists.Items
			for i := 0; i < len(artists); i++ {
				creator := &ObjectCreator{
					Name: artists[i].Name,
					URI:  "tidal:artist:" + artists[i].ID.String(),
				}
				objCreator := &Object{URI: creator.URI, Type: "creator", Provider: "tidal", Object: &jsontwo.RawMessage{}}
				objCreator.Object.UnmarshalJSON(creator.JSON())
				results.Creators = append(results.Creators, objCreator)
			}
		}
		if searchResults.Albums.TotalNumberOfItems > 0 {
			albums := searchResults.Albums.Items
			for i := 0; i < len(albums); i++ {
				album := &ObjectAlbum{
					Name: albums[i].Title,
					URI:  "tidal:album:" + albums[i].ID.String(),
				}
				objAlbum := &Object{URI: album.URI, Type: "album", Provider: "tidal", Object: &jsontwo.RawMessage{}}
				objAlbum.Object.UnmarshalJSON(album.JSON())
				results.Albums = append(results.Creators, objAlbum)
			}
		}
		if searchResults.Tracks.TotalNumberOfItems > 0 {
			tracks := searchResults.Tracks.Items
			for i := 0; i < len(tracks); i++ {
				stream := &ObjectStream{Name: tracks[i].Title}
				stream.URI = "tidal:track:" + tracks[i].ID.String()
				for _, artist := range tracks[i].Artists {
					objCreator := &ObjectCreator{Name: artist.Name, URI: "tidal:artist:" + artist.ID.String()}
					obj := &Object{URI: "tidal:artist:" + artist.ID.String(), Type: "creator", Provider: "tidal", Object: &jsontwo.RawMessage{}}
					obj.Object.UnmarshalJSON(objCreator.JSON())
					stream.Creators = append(stream.Creators, obj)
				}
				stream.Album = &Object{URI: "tidal:album:" + tracks[i].Album.ID.String(), Type: "album", Provider: "tidal", Object: &jsontwo.RawMessage{}}
				objAlbum := &ObjectAlbum{Name: tracks[i].Album.Title, URI: "tidal:album:" + tracks[i].Album.ID.String()}
				stream.Album.Object.UnmarshalJSON(objAlbum.JSON())
				stream.Duration, err = tracks[i].Duration.Int64()
				if err != nil {
					return results, err
				}
				objStream := &Object{URI: stream.URI, Type: "stream", Provider: "tidal", Object: &jsontwo.RawMessage{}}
				objStream.Object.UnmarshalJSON(stream.JSON())
				results.Streams = append(results.Streams, objStream)
			}
		}
		/*if searchResults.Playlists.TotalNumberOfItems > 0 {
			playlists := searchResults.Playlists.Items
			for i := 0; i < len(playlists); i++ {
				playlist := &ObjectPlaylist{
					Name: playlists[i].Title,
					URI:  "tidal:playlist:" + playlists[i].UUID,
				}

				results.Playlists = append(results.Playlists, &Object{Type: "playlist", Provider: "tidal", Object: playlist})
			}
		}*/
		/*if searchResults.Videos.TotalNumberOfItems > 0 {
			videos := searchResults.Videos.Items
			for i := 0; i < len(videos); i++ {
				stream := &ObjectStream{Name: videos[i].Title}
				stream.URI = "tidal:video:" + videos[i].ID.String()
				for _, artist := range videos[i].Artists {
					stream.Creators = append(stream.Creators, &Object{Type: "creator", Provider: "tidal", Object: &ObjectCreator{Name: artist.Name, URI: "tidal:artist:" + artist.ID.String()}})
				}
				stream.Album = &Object{Type: "album", Provider: "tidal", Object: &ObjectAlbum{Name: videos[i].Album.Title, URI: "tidal:album:" + videos[i].Album.ID.String()}}
				stream.Duration, err = videos[i].Duration.Int64()
				if err != nil {
					return results, err
				}

				results.Streams = append(results.Streams, &Object{Type: "stream", Provider: "tidal", Object: stream})
			}
		}*/
	}

	return results, nil
}

// TidalDeviceCode holds the response to a device authorization request
type TidalDeviceCode struct {
	StartTime time.Time                 `json:"-"`      //Used to expire this code when time.Now() > AuthStart + ExpiresIn
	Token     *oauth2.Token             `json:"token"`  //Holds the token for communicating with authenticated Tidal content
	OAuth2Cfg *clientcredentials.Config `json:"oauth2"` //Holds the OAuth2 configuration for this device code

	DeviceCode string `json:"deviceCode"`
	//UserCode                string  `json:"userCode"`
	//VerificationURI         string  `json:"verificationUri"`
	VerificationURIComplete string  `json:"verificationUriComplete"`
	ExpiresIn               float64 `json:"expiresIn"`
	Interval                int64   `json:"interval"`

	//UserID      int64  `json:"userID"`
	CountryCode string `json:"countryCode"`
}

// NeedsAuth returns true if client needs authenticating
func (t *TidalClient) NeedsAuth() bool {
	if t.Auth == nil || t.Auth.Token == nil || !t.Auth.Token.Valid() || t.Auth.OAuth2Cfg == nil || t.Auth.DeviceCode == "" || t.Auth.CountryCode == "" {
		return true
	}
	return false
}

// NewDeviceCode tries to get a new device code for user account pairing
func (t *TidalClient) NewDeviceCode() error {
	t.ClearDeviceCode()

	reqForm := url.Values{}
	reqForm.Set("client_id", t.ClientID)
	reqForm.Set("scope", "r_usr+w_usr+w_sub")
	resp, err := http.PostForm(tidalAuth+"/device_authorization", reqForm)
	if err != nil {
		return err
	}

	if err := json.UnmarshalBody(resp, &t.Auth); err != nil {
		return err
	}

	t.Auth.StartTime = time.Now()

	reqForm = url.Values{}
	reqForm.Set("device_code", t.Auth.DeviceCode)
	reqForm.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	t.Auth.OAuth2Cfg = &clientcredentials.Config{
		ClientID:     t.ClientID,
		ClientSecret: t.ClientSecret,
		TokenURL:     tidalAuth + "/token",
		Scopes: []string{
			"r_usr", "w_usr", "w_sub",
		},
		EndpointParams: reqForm,
		AuthStyle:      oauth2.AuthStyleInParams,
	}

	return nil
}

// WaitForAuth checks for a token pairing at regular intervals, failing if the device code expires
func (t *TidalClient) WaitForAuth() error {
	if t.Auth == nil || t.Auth.OAuth2Cfg == nil {
		return fmt.Errorf("no device code")
	}

	if t.HTTP == nil {
		t.HTTP = &oauth2.Transport{
			Source: t.Auth.OAuth2Cfg.TokenSource(context.Background()),
		}
	}
	if t.Auth.Token != nil && t.Auth.Token.Valid() {
		t.HTTP.Source = oauth2.ReuseTokenSource(t.Auth.Token, t.HTTP.Source)
		return nil
	}

	for {
		if t.Auth == nil || t.Auth.OAuth2Cfg == nil {
			return fmt.Errorf("device code revoked")
		}

		if time.Since(t.Auth.StartTime).Seconds() > t.Auth.ExpiresIn {
			return fmt.Errorf("device code expired")
		}

		token, err := t.HTTP.Source.Token()
		if err != nil {
			time.Sleep(time.Second * time.Duration(t.Auth.Interval))
			continue
		}

		t.Auth.Token = token
		//t.Auth.UserID = t.Auth.Token.Extra("user").(map[string]interface{})["userId"].(int64)
		t.Auth.CountryCode = t.Auth.Token.Extra("user").(map[string]interface{})["countryCode"].(string)
		return nil
	}
}

// ClearDeviceCode clears the current device code, nulling everything about the authenticated user
func (t *TidalClient) ClearDeviceCode() {
	t.Auth = nil
	t.HTTP = nil
}

// NewTidal returns a new Tidal client for continuous use
func NewTidal(clientID, clientSecret string) (t *TidalClient, err error) {
	t = &TidalClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	err = t.NewDeviceCode()
	return t, err
}

// NewTidalBlob returns a new Tidal client from a blob file for continuous use
func NewTidalBlob(blobPath string) (t *TidalClient, err error) {
	blob, err := os.Open(blobPath)
	if err != nil {
		return nil, err
	}
	defer blob.Close()

	blobJSON, err := ioutil.ReadAll(blob)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(blobJSON, &t)
	if err != nil {
		return nil, err
	}

	t.HTTP = &oauth2.Transport{
		Source: t.Auth.OAuth2Cfg.TokenSource(context.Background()),
	}
	if t.Auth.Token != nil && t.Auth.Token.Valid() {
		t.HTTP.Source = oauth2.ReuseTokenSource(t.Auth.Token, t.HTTP.Source)
	}

	return t, err
}

// SaveBlob saves a Tidal client to a blob file for later use
func (t *TidalClient) SaveBlob(blobPath string) {
	blob, err := os.Create(blobPath)
	if err != nil {
		Error.Println(err)
		return
	}
	defer blob.Close()

	blobJSON, err := json.Marshal(t, true)
	if err != nil {
		Error.Println(err)
		return
	}

	_, _ = blob.Write(blobJSON)
}

// ReplaceURI replaces all instances of a URI with a libremedia-acceptable URI
func (t *TidalClient) ReplaceURI(text string) string {
	return tidalurire.ReplaceAllStringFunc(text, tidalReplaceURI)
}

func tidalReplaceURI(link string) string {
	fmt.Println("Testing string: " + link)
	match := tidalurire.FindAllStringSubmatch(link, 1)
	if len(match) > 0 {
		typed := match[0][1]
		id := match[0][2]
		name := match[0][3]
		switch typed {
		case "albumId":
			typed = "album"
		case "artistId":
			typed = "creator"
		}
		return "[" + name + "](/" + typed + "?uri=tidal:" + typed + ":" + id + ")"
	}
	return link
}
