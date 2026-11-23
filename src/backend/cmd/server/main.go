package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"eigentsar/api"
	"eigentsar/model/updater"
	"eigentsar/storage/dataset"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dataDir := flag.String("data", "../../data", "dataset root containing train/ and validation/")
	modelDir := flag.String("models", "../../models", "directory for the saved model artifact")
	k := flag.Int("k", 50, "number of eigenfaces (principal components)")
	flag.Parse()

	mgr := updater.New(dataset.Store{Root: *dataDir}, filepath.Join(*modelDir, "model.gob"), *k)
	if err := mgr.Init(); err != nil {
		log.Printf("WARNING: model belum siap: %v", err)
	}

	srv := &http.Server{Addr: *addr, Handler: (&api.API{Mgr: mgr}).Routes(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()
	log.Printf("listening on %s", *addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
