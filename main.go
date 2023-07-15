package main

/*
	Title:		libremedia
	Version:	1.0
	Author:		Joshua "JoshuaDoes" Wickings
	License:	GPL v3

	Failure to comply with this license will result in legal penalties as permitted by copyright law.
*/

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eolso/librespot-golang/librespot/utils"
)

type exporterr struct {
	Error string `json:"error"`
}

type valid struct {
	Valid bool `json:"valid"`
}

var (
	service *Service
)

var (
	//Trace logs trace info
	Trace *log.Logger
	//Info logs information
	Info *log.Logger
	//Warning logs warnings
	Warning *log.Logger
	//Error logs errors
	Error *log.Logger
)

func initLogging(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	initLogging(os.Stderr, os.Stdout, os.Stderr, os.Stderr)

	var err error

	//Open the configuration
	configFile, err := os.Open("config.json")
	if err != nil {
		Error.Println("need configuration")
		return
	}
	defer configFile.Close()

	//Load the configuration into memory
	configParser := json.NewDecoder(configFile)
	if err = configParser.Decode(&service); err != nil {
		Error.Println("error loading configuration: " + fmt.Sprintf("%v", err))
		return
	}

	err = service.Login()
	if err != nil {
		Error.Println("error logging in: " + fmt.Sprintf("%v", err))
	}

	//libremedia API v1
	http.HandleFunc("/v1/", v1Handler)
	http.HandleFunc("/v1/stream/", v1StreamHandler)
	http.HandleFunc("/v1/download/", v1DownloadHandler)

	//Built-in utilities that may not be recreatable in some circumstances
	http.HandleFunc("/util/gid2id/", gid2id)

	//Web interfaces
	http.HandleFunc("/", webHandler)

	Warning.Fatal(http.ListenAndServe(service.HostAddr, nil))
}

func v1Handler(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path[4:]
	obj := GetObject(uri)
	if obj == nil {
		jsonWriteErrorf(w, 404, "no matching object")
		return
	}
	if !obj.Expanded && !obj.Expanding {
		switch obj.Type {
		case "album":
			obj.Expand()
		default:
			go obj.Expand()
		}
	}
	jsonWrite(w, obj)
}

func v1DownloadHandler(w http.ResponseWriter, r *http.Request) {
	settings := r.URL.Query()

	path := strings.Split(r.URL.Path[13:], "?")
	objectStream := GetObjectLive(path[0])
	if objectStream == nil {
		jsonWriteErrorf(w, 404, "no matching stream object")
		return
	}
	if objectStream.Type != "stream" {
		jsonWriteErrorf(w, 404, "no matching stream object")
		return
	}
	stream := objectStream.Stream()
	if stream == nil {
		jsonWriteErrorf(w, 500, "unable to process stream object")
		return
	}
	formatCfg := settings.Get("format")
	if formatCfg == "" {
		formatCfg = "0"
	}
	formatNum, err := strconv.Atoi(formatCfg)
	if err != nil {
		jsonWriteErrorf(w, 500, "libremedia: format selection unavailable")
		return
	}
	err = service.Download(w, r, stream, formatNum)
	if err != nil {
		if settings.Get("format") != "" {
			jsonWriteErrorf(w, 500, "libremedia: format selection unavailable for matched stream object")
			return
		}
		if len(stream.Formats) == 0 {
			jsonWriteErrorf(w, 404, "libremedia: no formats available to select from for matched stream object")
			return
		}
		streamed := false
		for i := formatNum; i < len(stream.Formats); i++ {
			if stream.Formats[i] != nil {
				Trace.Println("Selecting format " + stream.Formats[i].Name + " automatically")
				err = service.Download(w, r, stream, i)
				if err != nil {
					continue
				}
				streamed = true
				break
			}
		}
		if !streamed {
			jsonWriteErrorf(w, 500, "libremedia: all formats available to select from matched stream object were null")
			return
		}
	}
	return
}

func v1StreamHandler(w http.ResponseWriter, r *http.Request) {
	settings := r.URL.Query()

	path := strings.Split(r.URL.Path[11:], "?")
	objectStream := GetObjectLive(path[0])
	if objectStream == nil {
		jsonWriteErrorf(w, 404, "no matching stream object")
		return
	}
	if objectStream.Type != "stream" {
		jsonWriteErrorf(w, 404, "no matching stream object")
		return
	}
	stream := objectStream.Stream()
	if stream == nil {
		jsonWriteErrorf(w, 500, "unable to process stream object")
		return
	}
	formatCfg := settings.Get("format")
	if formatCfg == "" {
		formatCfg = "0"
	}
	formatNum, err := strconv.Atoi(formatCfg)
	if err != nil {
		jsonWriteErrorf(w, 500, "libremedia: format selection unavailable")
		return
	}

	err = service.Stream(w, r, stream, formatNum)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "broken pipe") || strings.Contains(errMsg, "connection reset") {
			return //The stream was successful, but interrupted
		}
		if settings.Get("format") != "" {
			jsonWriteErrorf(w, 500, "libremedia: format selection unavailable for matched stream object")
			return
		}
		if len(stream.Formats) == 0 {
			jsonWriteErrorf(w, 404, "libremedia: no formats available to select from for matched stream object")
			return
		}
		streamed := false
		for i := formatNum; i < len(stream.Formats); i++ {
			if stream.Formats[i] != nil {
				Trace.Println("Selecting format " + stream.Formats[i].Name + " automatically")
				err = service.Stream(w, r, stream, i)
				if err != nil {
					errMsg = err.Error()
					if strings.Contains(errMsg, "broken pipe") || strings.Contains(errMsg, "connection reset") {
						return //The stream was successful, but interrupted
					}
					continue
				}
				streamed = true
				break
			}
		}
		if !streamed {
			jsonWriteErrorf(w, 500, "libremedia: all formats available to select from matched stream object were null")
			return
		}
	}
	return
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	var file *os.File
	var err error

	low := len(r.URL.Path) - 4 //favicon.ico n
	high := len(r.URL.Path)    //favicon.ico o
	if r.URL.Path == "/" {
		Warning.Println("Serving HTML: /")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		file, err = os.Open("index.html")
		if err != nil {
			panic("need index.html!")
		}
	} else {
		file, err = os.Open(string(r.URL.Path[1:]))
		if err != nil {
			Warning.Println("Serving 404! " + r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if string(r.URL.Path[low-1:high]) == ".html" {
			Warning.Println("Serving HTML:", r.URL.Path)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else if string(r.URL.Path[low:high]) == ".css" {
			Warning.Println("Serving CSS:", r.URL.Path)
			w.Header().Set("Content-Type", "text/css")
		} else if string(r.URL.Path[low+1:high]) == ".js" {
			Warning.Println("Serving JS:", r.URL.Path)
			w.Header().Set("Content-Type", "text/javascript")
		} else {
			Warning.Println("Serving content:", r.URL.Path)
		}
	}
	defer file.Close()

	http.ServeContent(w, r, "", time.Time{}, file)
}

type expid struct {
	ID string `json:"id"`
}

func gid2id(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	remote := getRemote(r)

	Info.Println(remote, "gid2id:", "Getting gid")

	gid := ""
	for key, values := range r.URL.Query() {
		for _, value := range values {
			if key == "gid" {
				gid = value
				break
			}
		}
	}
	if gid == "" {
		Error.Println(remote, "gid2id:", "No gid specified")
		jsonWriteErrorf(w, 405, "endpoint requires gid")
		return
	}
	gid = strings.ReplaceAll(gid, " ", "+")

	str, err := base64.StdEncoding.DecodeString(gid)
	if err != nil {
		Error.Println(remote, "gid2id:", "Invalid gid", gid, ": ", err)
		jsonWriteErrorf(w, 405, "invalid gid %s: %v", gid, err)
	}

	id := utils.ConvertTo62(str)

	Info.Println(remote, "gid2id:", "Converted", gid, "to", id)
	jsonWrite(w, &expid{ID: id})
}

/*func suggestHandler(w http.ResponseWriter, r *http.Request) {
	remote := getRemote(r)

	Info.Println(remote, "suggest:", "Checking passkey")

	allowed := false
	suggestQuery := ""
	for key, values := range r.URL.Query() {
		for _, value := range values {
			switch key {
			case "pass":
				if value == passkey {
					allowed = true
				}
			case "query":
				suggestQuery = value
			}
		}
	}
	if allowed == false {
		Error.Println(remote, "suggest:", "Invalid passkey")
		jsonWriteErrorf(w, 401, "invalid pass key")
		return
	}

	Info.Println(remote, "suggest:", "Sending suggest for query \""+suggestQuery+"\"")
	displaySuggest(w, suggestQuery)
}*/

func jsonWrite(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//Allow marshalling special cases
	switch typedData := data.(type) {
	case error:
		json, err := json.Marshal(&exporterr{Error: typedData.Error()})
		if err != nil {
			Error.Println("Could not marshal data [ ", err, " ]:", data)
			jsonWriteErrorf(w, 500, "could not prep data")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		Error.Printf("Sent error: %v\n", typedData.Error())
	default:
		json, err := json.Marshal(data)
		if err != nil {
			Error.Println("Could not marshal data [ ", err, " ]:", data)
			jsonWriteErrorf(w, 500, "could not prep data")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}
}
func jsonWriteErrorf(w http.ResponseWriter, statusCode int, error string, data ...interface{}) {
	errMsg := fmt.Errorf(error, data...)

	json, err := json.Marshal(&exporterr{Error: errMsg.Error()})
	if err != nil {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		Error.Println("Could not marshal data [ ", err, " ]:", data)
		w.Write([]byte("500 Internal Server Error"))
		return
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(json)
	Error.Printf("Sent error %d: %v\n", statusCode, errMsg)
}

func getRemote(r *http.Request) string {
	userAgent := r.Header.Get("User-Agent")

	remote := "[" + r.RemoteAddr
	if userAgent != "" {
		remote += " UA(" + userAgent + ")"
	}
	remote += "]"

	return remote
}
