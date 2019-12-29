package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"runtime"
)

func uploadHandlerFunc(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("upload.html")
	t.Execute(w, len(TILESDB))
}

func fetchHandlerFunc(w http.ResponseWriter, r *http.Request) {
	files, _ := ioutil.ReadDir("tiles")
	t, _ := template.ParseFiles("fetch.html")
	t.Execute(w, len(files))
}

func RecoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Starting mosaic server ...")
	mux := http.NewServeMux()
	files := http.FileServer(http.Dir("public"))
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	mux.Handle("/", RecoverWrap(http.HandlerFunc(uploadHandlerFunc)))
	mux.Handle("/reload", RecoverWrap(http.HandlerFunc(reloadTilesDBHandlerFunc)))
	mux.Handle("/fetch", RecoverWrap(http.HandlerFunc(fetchHandlerFunc)))
	mux.Handle("/fetch_tiles", RecoverWrap(http.HandlerFunc(fetchTilesHandlerFunc)))
	mux.Handle("/mosaic_no_concurrency", RecoverWrap(http.HandlerFunc(noConcurrencyHandlerFunc)))
	mux.Handle("/mosaic_fanout_channel", RecoverWrap(http.HandlerFunc(fanOutWithChannelHandlerFunc)))
	mux.Handle("/mosaic_fanout_fanin", RecoverWrap(http.HandlerFunc(fanOutFanInHandlerFunc)))

	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}
	TILESDB = tilesDB()
	fmt.Println("Mosaic server started.")
	server.ListenAndServe()

}
