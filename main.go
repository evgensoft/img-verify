package main

import (
	"html"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"img-verify/handlers"
	"img-verify/logger"

	_ "embed"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

//go:embed facefinder
var cascadeFile []byte

var log = logger.GetLogger()

func main() {
	log.Info("Запуск программы")

	if err := handlers.CascadeInit(cascadeFile); err != nil {
		log.Fatalf("Error reading the cascade file: %s", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		log.Infof("%v %q", r.Method, html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/get_image_info", handlers.GetImageInfo)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	s := <-sig

	log.Infof("SIGTERM received - %v - shutting down", s.String())
	_ = log.Sync()
}
