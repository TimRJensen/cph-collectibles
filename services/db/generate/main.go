package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/cph-collectibles/db/seed"
)

//go:embed insert.sql.tmpl
var insert string

func main() {
	src, err := os.Open("./posters.csv")
	if err != nil {
		log.Fatal(fmt.Errorf("gen: %w", err))
	}
	defer src.Close()

	data, err := seed.ParseCSV(csv.NewReader(src))
	if err != nil {
		log.Fatal(fmt.Errorf("gen: %w", err))
	}

	tmpl, err := template.New("insert").
		Funcs(template.FuncMap{
			"add": func(a, b int) int { return a + b },
		}).Parse(insert)
	if err != nil {
		log.Fatal(fmt.Errorf("gen: %w", err))
	}

	buff := bytes.NewBuffer(nil)
	err = tmpl.Execute(buff, data)
	if err != nil {
		log.Fatal(fmt.Errorf("gen: %w", err))
	}

	dst, err := os.Create("./seed.sql")
	if err != nil {
		log.Fatal(fmt.Errorf("gen: %w", err))
	}
	defer dst.Close()

	dst.WriteString(buff.String())
	os.Exit(0)
}
