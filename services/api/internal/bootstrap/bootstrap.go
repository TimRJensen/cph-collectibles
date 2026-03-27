package bootstrap

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/seed"
	"github.com/cph-collectibles/db/tables/files"
	"github.com/cph-collectibles/db/wrappers"
	"github.com/oklog/ulid/v2"
)

type SeedConfig struct {
	CSVPath   string
	ZIPPath   string
	AssetsDir string
	FSURL     string
}

func SeedDB(p *db.Pool, cfg SeedConfig) error {
	r, err := zip.OpenReader(cfg.ZIPPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	rexp := regexp.MustCompile(`\w+/(\w+)[.]png`)
	imgs := map[string]*zip.File{}
	for _, f := range r.File {
		matches := rexp.FindStringSubmatch(f.Name)
		if len(matches) < 2 {
			continue
		}
		imgs[matches[1]] = f
	}

	data, err := seed.SeedCSV(cfg.CSVPath, p)
	if err != nil {
		return fmt.Errorf("seed csv: %w", err)
	}

	filedir, err := filepath.Abs(cfg.AssetsDir)
	if err != nil {
		return fmt.Errorf("abs assets dir: %w", err)
	}

	ctx := context.Background()
	tx := p.Transaction(ctx)

	for _, d := range data {
		id, raw := d.Id.String(), d.Meta.RawId

		path := filepath.Join(filedir, id)
		if err := os.MkdirAll(path, 0o777); err != nil {
			return fmt.Errorf("mkdir %s: %w", path, err)
		}

		v, ok := imgs[raw]
		if !ok {
			continue
		}

		dstPath := filepath.Join(path, fmt.Sprintf("%s.png", raw))
		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("create dst %s: %w", dstPath, err)
		}

		src, err := v.Open()
		if err != nil {
			dst.Close()
			return fmt.Errorf("open zip entry %s: %w", v.Name, err)
		}

		_, copyErr := io.Copy(dst, src)
		closeDstErr := dst.Close()
		closeSrcErr := src.Close()

		if copyErr != nil {
			return fmt.Errorf("copy image %s: %w", raw, copyErr)
		}
		if closeDstErr != nil {
			return fmt.Errorf("close dst %s: %w", dstPath, closeDstErr)
		}
		if closeSrcErr != nil {
			return fmt.Errorf("close src %s: %w", v.Name, closeSrcErr)
		}

		rec := files.Data{}
		rec.Id = wrappers.ULID(ulid.Make())
		rec.PosterId = wrappers.ULID(d.Id)
		rec.URL = fmt.Sprintf("/fs/assets/%s/%s.png", id, raw)

		if err := files.Insert(tx, ctx, &rec); err != nil {
			return fmt.Errorf("insert file row for %s: %w", raw, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func DefaultSeedConfig() SeedConfig {
	cwd, _ := os.Getwd()
	return SeedConfig{
		CSVPath:   filepath.Join(cwd, "data", "posters.csv"),
		ZIPPath:   filepath.Join(cwd, "data", "posters.zip"),
		AssetsDir: filepath.Join(cwd, "assets"),
		FSURL:     os.Getenv("FS_URL"),
	}
}

func MustSeedDB(p *db.Pool, cfg SeedConfig) {
	if err := SeedDB(p, cfg); err != nil {
		log.Fatal(err)
	}
}
