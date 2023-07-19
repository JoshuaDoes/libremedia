# libremedia

## How to configure for testing

* WARNING: Spotify playback is broken with the Go port of librespot right now. The catalogue functions still work, but don't expect playback until the plugin system is finished and we switch to using the Rust version of librespot.
* Check back here with any `git pull` attempts, just in case a configuration format change has been made.

Place the following JSON in a file named `config.json` directly next to your compiled `libremedia` binary (`libremedia.exe` on Windows):

```JSON
{
        "httpAddr": ":80",
        "baseURL": "http://example.com",
        "handlers": {
                "tidal": {
                        "active": true,
                        "username": "zU4XHVVkc2tDPo4t",
                        "password": "VJKhDFqJPqvsPVNBV6ukXTJmwlvbttP7wlMlrc72se4=",
                        "deviceName": "Android Auto (Normal, High, HiFi, Master)",
                        "blobPath": ".tidal.blob"
                },
                "spotify": {
                        "active": true,
                        "username": "changeme",
                        "password": "changeme",
                        "deviceName": "librespot",
                        "blobPath": ".spotify.blob"
                }
        },
}
```

- Change the port in `httpAddr` to something else if you have an HTTP website using port 80 on your host. HTTPS is not supported at this time, use an HTTPS reverse proxy with Apache2 or similar.
- Change `baseURL` to either your IP address or the domain you'll be using to test libremedia.
- If the example Tidal API key stops working, find your own or wait for this README to update with a new one.
- Change the `username` and `password` fields under the Spotify handler to match your account.
- If desired, change the `blobPath` in your handlers to point to where you want your authentication tokens to be saved. The defaults will normally hide them on Linux.
- If you don't have an account for a given handler, set the `active` field to false.

### Progress tracker before release

# User interface

- Fetch artwork for an object using `:artwork` URI extension instead of re-fetching the entire source object, use `:1280` to smartly limit size to 1280px or next size up for max quality
- Add visibility toggle buttons to both the search bar and the audio player
- Increase size of controls in tables to maximize room for touch screens, OR migrate all controls to drop-down list, whichever works out better (or condense controls to it if table entry is squished?)

## Pages

- Display top 100 streams and top 100 downloads on home page
- Add "download album" control on album page, generates and saves single-folder ZIP of all streams client-side with per-stream progress bar
- Add "download discography" control on creator page, generates and saves multi-folder ZIP of all albums, EPs and singles client-side with per-album and per-stream progress bars
- Add playlists section on search and creator pages, uses same handler for album objects
- Display entry numbers and total X of each table section

- Create queue management using /queue with no params
* Generic streams table, only shows either the regular queue or the shuffled queue
* Add reorder control, injects clickable dividers in-between streams that will move selected stream to that location
* Add unqueue control, removes stream from queue
* Allow saving current queue as a libremedia playlist

- Rewrite logic for transcript page
* Fetch transcript with timings using `:transcript` URI extension instead of re-fetching entire stream object
* Start audio position 0 with empty line (unless first line begins at pos 0) using timings index 0 instead of fake -1 like before (invalid index)
* Signal unplayed/restarted auto-scroller position tracker using -1 instead of -2 like before
* Always load now playing transcript into timings buffer when stream changes, automaticlly scroll back to top
* Add fixed resume auto-scroll button to bottom right when user scrolls to disable auto-scroll

## Audio player

- Add vertical volume slider
- Add horizontal audio position seeker directly above audio player, show regardless of audio player visibility
- Change transcript button to specify no params instead of pointing to now playing stream URI, no params will sync transcript with now playing
- Display smaller album art somewhere, use either same size or next size up for max quality

- Add shuffled queue support
* Place shuffle button somewhere
* Generate new shuffled queue using existing queue on every enable, and use it for next/prev handling until disable
* On disable, return to original queue regardless of what was next prior to enable

# Backend

- Start anonymously tracking listen and download counts
- At end of object expansion goroutine, spawn new goroutine to search other providers for matching object to fill in missing metadata (such as credited creators, biographies, artwork, albums, streams, etc)
- Migrate all client-side player controls to the server, simulating client actions based on client requests
- Require clients to request to start playback (automatically acting as "I'm ready" for shared sessions to minimize latency), so they always load from `/v1/stream` with no params afterward
- Add `/v1/providers` endpoint to return all upstream providers and a path to retrieve their icons, which can follow upstream to the source provider
- Add providers array param to `/v1/search` endpoint, allows client-side filtering of providers via request (useful to minimize processing and deduplicate responses when a provider is shared across upstream instances)
- Convert transcript handler to be separated transcript providers, also available as plugins
- Allow catalogue and database providers to be implemented as multimedia providers, without the streams
- Implement support for ffmpeg (for custom format, codec, and quality params, plus metadata injection with `/v1/download` endpoint)
- Implement support for ffprobe (pointing to internal `/v1/stream` API) to identify format details if not available on provider, but direct stream is available
- Find a reliable and free way to identify audio streams with no metadata

## Plugins

- Finish writing new plugin system, simple concept but a lot to explain, later
- Implement the Rust librespot client as a binary plugin to provide Spotify, and add podcast support
- Remove hardcoded Spotify provider
- Migrate hardcoded Tidal provider to a binary plugin, and add music video support
- Create plugins for many other things, like local content, disc drive / ISO support, torrent support, YouTube, Bandcamp, SoundCloud, Apple Music, Deezer, etc

## User accounts

- Provide a default guest user using server config
- Provide a default admin user with permission to manage additional users and control the quality settings, globally and of each user

- Cache now playing stream to disk as it buffers (can be random access too)
* Hold lock on file until all sessions holding lock either timeout or all choose new streams
* Avoid repeated reads with session sharing by using the same buffer, as all sessions are synced

- Add session sharing support
* Generate an invite link to share with any user
* Sync queue, now playing, and audio position in real time (with both a low-latency and data saver mode client-side)
* If sharer toggles it, users may vote to skip with >50% majority
* If sharer toggles it, users may append to queue (with a rotational selection mode in client join order, and a free for all mode)
