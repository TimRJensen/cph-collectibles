package seed

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"

	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/tables/posters"
	"github.com/cph-collectibles/db/wrappers"
	"github.com/oklog/ulid/v2"
)

func atof(v string) float64 {
	ipart, fpart, k := 0.0, 0.0, 0.0
	for _, c := range v {
		if c == 0x2E || c == 0x2C {
			k = 1.0
			continue
		}
		if k == 0.0 {
			ipart = ipart*10 + float64(c-0x30)
		} else {
			k *= 0.1
			fpart = fpart + float64(c-0x30)*k
		}
	}
	return ipart + fpart
}

func ctof(v string) float64 {
	ipart, fpart, k := 0.0, 0.0, 0.0
	for _, c := range v {
		if c == 0xA3 {
			break
		}
		if c == 0x2E || c == 0x2C {
			k = 1.0
			continue
		}
		if k == 0.0 {
			ipart = ipart*10 + float64(c-0x30)
		} else {
			k *= 0.1
			fpart = fpart + float64(c-0x30)*k
		}
	}
	return ipart + fpart
}

type data []posters.Data

func parseOrigin(v string) []string {
	res := []string{"", ""}
	i, j := 0, 0
	for i < len(v) {
		switch v[i] {
		case 0x2C:
			if j < i {
				res[0] = strings.TrimSpace(v[j:i])
			}
			i += 1
			j = i
		case 0x28, 0x29:
			if j < i {
				res[1] = strings.TrimSpace(v[j:i])
			}
			i += 1
			j = i
		case 0x20:
			if j == i {
				i += 1
				j = i
			} else {
				i++
			}
		default:
			i++
		}
	}
	return res
}

func parseDetail(v string) []string {
	res := []string{}
	i, j := 0, 0
	for i < len(v) {
		switch v[i] {
		case 0x28:
			if j < i {
				res = append(res, strings.TrimSpace(v[j:i]))
				res = append(res, parseOrigin(v[i:])...)
			}
			i += len(v) - i
			j = i
		case 0x20:
			if j == i {
				i += 1
				j = i
			} else {
				i++
			}
		default:
			i++
		}
	}
	if j < i {
		res = append(res, strings.TrimSpace(v[j:i]))
	}
	return res
}

func parseSize(v string) []string {
	res := []string{}
	i, j := 0, 0
	for i < len(v) {
		switch v[i] {
		case 0x20:
			if j < i {
				res = append(res, strings.TrimSpace(v[j:i]))
			}
			i += 1
			j = i
		case 0x78:
			i += 1
			j = i
		case 0x63:
			if i+1 < len(v) && v[i+1] == 0x6D {
				i += 2
				j = i
			} else {
				i++
			}
		default:
			i++
		}
	}
	return res
}

func parseCondition(v string) []string {
	res := []string{}
	i, j := 0, 0
	for i < len(v) {
		switch v[i] {
		case 0x2D, 0x2E:
			if j < i {
				res = append(res, strings.TrimSpace(strings.ToLower(v[j:i])))
			}
			i += 1
			j = i
		default:
			i++
		}
	}
	if j < i {
		res = append(res, strings.TrimSpace(strings.ToLower(v[j:i])))
	}
	return res
}

func normalize(s string) string {
	replacer := strings.NewReplacer(
		"–", "-",
		"—", "-",
		"×", "x",
	)
	s = replacer.Replace(s)
	s = strings.TrimSpace(s)
	return s
}

func ParseCSV(r *csv.Reader) (data, error) {
	res := data{}

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		if i < 2 {
			continue
		}

		record = record[1:]
		if record[1] == "" {
			continue
		}

		data := posters.Data{Id: wrappers.ULID(ulid.Make())}
		cost := ctof(record[4])
		details := parseDetail(normalize(record[1]))
		size := parseSize(normalize(record[2]))
		cond := parseCondition(normalize(record[3]))

		// convert to float for display & let psql handle transform to minor units
		data.Cost.RawAmount = cost
		data.Cost.RawVAT = 0.0

		data.Meta.RawId = record[0]
		// details may not have origin+year
		if len(details) < 3 {
			data.Detail.Heading = details[0]
			data.Detail.Origin.Source = "unknown"
			data.Detail.Origin.Year = "unknown"
		} else {
			data.Detail.Heading = details[0]
			data.Detail.Origin.Source = details[1]
			data.Detail.Origin.Year = details[2]
		}
		data.Detail.Width = atof(size[0])
		data.Detail.Height = atof(size[1])
		data.Condition.Rating = cond[0]
		data.Condition.Notes = cond[1]
		res = append(res, data)
	}
	return res, nil
}

func SeedCSV(path string, p *db.Pool) (data, error) {
	cwd, _ := os.Getwd()
	path, err := filepath.Abs(filepath.Join(cwd, path))
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := ParseCSV(csv.NewReader((f)))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	tx := p.Transaction(ctx)
	for _, d := range data {
		if err := posters.Insert(tx, ctx, &d); err != nil {
			tx.Rollback(ctx)
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return data, nil
}
