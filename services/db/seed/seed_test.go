package seed

import (
	"encoding/csv"
	"os"
	"testing"
)

const testdata = `,,,,,,
,ID Number ,Title,Poster size,Condition,Price,
,0001,"Porsche 928 S Dealer (Germany, c.1986)",100 cm x 75 cm,Good – Small edge tear,120£,
,0029,Mercedes-Benz – 50 Years of Classic Sport (1926–1976) ,54 × 39.5 cm ,"Fair – fold lines, foxing and staining along lower edge. ",60£,
,0032,Michelin Motorsport Poster – Tom Kristensen (c.2000) ,59 × 42 cm ,"Fair – fold lines, edge tear, creasing. ",110£,
,0041,Porsche 944 Cutaway – Der neue Porsche 944” ,100 × 75 cm ,"Good condition. Light creasing and a small edge tear.
",120£,

`

func TestParseCSV(t *testing.T) {
	dir := t.TempDir()

	f, err := os.CreateTemp(dir, "posters.csv")
	if err != nil {
		t.Fatalf("test error: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(testdata)
	if err != nil {
		t.Fatalf("test error: %v", err)
	}
	f.Seek(0, 0)

	data, err := ParseCSV(csv.NewReader(f))
	if err != nil {
		t.Fatalf("test error: %v", err)
	}

	detail := "Porsche 928 S Dealer"
	got := data[0]
	if got.Detail.Heading != detail {
		t.Fatalf("test mismatch: got: %s, want: %s", got.Detail.Heading, detail)
	}

	detail = "Germany"
	if got.Detail.Origin.Source != detail {
		t.Fatalf("test mismatch: got: %s, want: %s", got.Detail.Origin.Source, detail)
	}

	detail = "c.1986"
	if got.Detail.Origin.Year != detail {
		t.Fatalf("test mismatch: got: %s, want: %s", got.Detail.Origin.Year, detail)
	}

	size := 100.0
	if got.Detail.Width != size {
		t.Fatalf("test mismatch: got: %f, want: %f", got.Detail.Width, size)
	}

	size = 75.0
	if got.Detail.Height != size {
		t.Fatalf("test mismatch: got: %f, want: %f", got.Detail.Height, size)
	}

	cond := "Good"
	if got.Condition.Rating != cond {
		t.Fatalf("test mismatch: got: %s, want: %s", got.Condition.Rating, cond)
	}

	cond = "Small edge tear"
	if got.Condition.Notes != cond {
		t.Fatalf("test mismatch: got: %s, want: %s", got.Condition.Rating, cond)
	}

	detail = "1926-1976"
}
