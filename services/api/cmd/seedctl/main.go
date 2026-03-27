package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cph-collectibles/api/internal/bootstrap"
	"github.com/cph-collectibles/db"
)

func main() {
	var (
		csvPath   = flag.String("csv", "./data/posters.csv", "Path to posters.csv")
		zipPath   = flag.String("zip", "./data/posters.zip", "Path to posters.zip")
		assetsDir = flag.String("assets", "./assets", "Directory to write extracted poster images into")
		fsURL     = flag.String("fs-url", os.Getenv("FS_URL"), "Base FS_URL for generated asset URLs")
	)
	flag.Parse()

	if *fsURL == "" {
		log.Fatal("missing -fs-url or FS_URL")
	}

	p, err := db.NewPool(db.LoadConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	cfg := bootstrap.SeedConfig{
		CSVPath:   *csvPath,
		ZIPPath:   *zipPath,
		AssetsDir: *assetsDir,
		FSURL:     *fsURL,
	}

	if err := bootstrap.SeedDB(p, cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("seed complete")
}
