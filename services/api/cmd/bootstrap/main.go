package main

import (
	"flag"
	"log"

	"github.com/cph-collectibles/api/internal/bootstrap"
	"github.com/cph-collectibles/db"
)

func main() {
	var (
		csvPath = flag.String("csv", "./data/posters.csv", "Path to posters.csv")
		zipPath = flag.String("zip", "./data/posters.zip", "Path to posters.zip")
		dst     = flag.String("dst", "./assets", "Directory to write extracted poster images into")
	)
	flag.Parse()

	p, err := db.NewPool(db.LoadConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	bootstrap.MustBootstrap(p, *csvPath, *zipPath, *dst)
}
