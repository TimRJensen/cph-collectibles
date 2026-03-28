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

func seedDB(p *db.DB, pathToCSV string) (seed.Data, error) {
	return seed.SeedCSV(pathToCSV, p)
}

func seedFiles(p *db.DB, data seed.Data, pathToZIP, dst string) error {
	r, err := zip.OpenReader(pathToZIP)
	if err != nil {
		return fmt.Errorf("seedFiles: %w", err)
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

	filedir, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("seedFiles: %w", err)
	}

	ctx := context.Background()
	tx := p.Transaction(ctx)
	for _, d := range data {
		id, raw := d.Id.String(), d.Meta.RawId

		path := filepath.Join(filedir, id)
		if err := os.MkdirAll(path, 0o777); err != nil {
			return fmt.Errorf("seedFiles: %w", err)
		}

		v, ok := imgs[raw]
		if !ok {
			continue
		}

		dstPath := filepath.Join(path, fmt.Sprintf("%s.png", raw))
		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("seedFiles: %w", err)
		}

		src, err := v.Open()
		if err != nil {
			dst.Close()
			return fmt.Errorf("seedFiles: %w", err)
		}

		if _, err = io.Copy(dst, src); err != nil {
			return fmt.Errorf("seedFiles: %w", err)
		}
		if err = dst.Close(); err != nil {
			return fmt.Errorf("seedFiles: %w", err)
		}
		if err = src.Close(); err != nil {
			return fmt.Errorf("seedFiles: %w", err)
		}

		rec := files.Data{}
		rec.Id = wrappers.ULID(ulid.Make())
		rec.PosterId = wrappers.ULID(d.Id)
		rec.URL = fmt.Sprintf("/fs/assets/%s/%s.png", id, raw)

		if err := files.Insert(tx, ctx, &rec); err != nil {
			return fmt.Errorf("seedFiles: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("seedFiles: %w", err)
	}

	return nil
}

func MustBootstrap(p *db.DB, pathToCSV, pathToZIP, dst string) {
	data, err := seedDB(p, pathToCSV)
	if err != nil {
		log.Fatal(err)
	}
	err = seedFiles(p, data, pathToZIP, dst)
	if err != nil {
		log.Fatal(err)
	}
}
