package segment

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TableSerializer converts an HTML <table> into a single text string for a RawBlock.
type TableSerializer interface {
	Serialize(table *goquery.Selection) string
}

// extractHeaders returns <th> texts from the first row (or <thead>).
func extractHeaders(table *goquery.Selection) []string {
	var headers []string
	// Try <thead> first, then first <tr>
	headerRow := table.Find("thead tr").First()
	if headerRow.Length() == 0 {
		headerRow = table.Find("tr").First()
	}
	headerRow.Find("th").Each(func(_ int, th *goquery.Selection) {
		headers = append(headers, strings.TrimSpace(th.Text()))
	})
	return headers
}

// dataRows returns all <tr> that contain <td> (skipping header-only rows).
func dataRows(table *goquery.Selection) []*goquery.Selection {
	var rows []*goquery.Selection
	table.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		if tr.Find("td").Length() > 0 {
			rows = append(rows, tr)
		}
	})
	return rows
}

// --- CompactTableSerializer ---

// CompactTableSerializer puts headers once on the first line,
// then one line per data row. Cells separated by ", ", rows by "\n".
//
// Предмет, Оценка, Преподаватель.
// Математика, 5, Иванов.
// Физика, 4, Петрова.
type CompactTableSerializer struct{}

func (s *CompactTableSerializer) Serialize(table *goquery.Selection) string {
	headers := extractHeaders(table)
	rows := dataRows(table)
	if len(rows) == 0 {
		return ""
	}

	var sb strings.Builder

	if len(headers) > 0 {
		sb.WriteString(strings.Join(headers, ", "))
		sb.WriteString(".\n")
	}

	for i, tr := range rows {
		var cells []string
		tr.Find("td").Each(func(_ int, td *goquery.Selection) {
			cells = append(cells, strings.TrimSpace(td.Text()))
		})
		if len(cells) == 0 {
			continue
		}
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(strings.Join(cells, ", "))
		sb.WriteString(".")
	}

	return sb.String()
}

// --- LinearizedTableSerializer ---

// LinearizedTableSerializer repeats column headers for each data row.
// Better for search (query "оценка по физике" matches the row directly)
// but inflates word count.
//
// Предмет: Математика. Оценка: 5. Преподаватель: Иванов
// Предмет: Физика. Оценка: 4. Преподаватель: Петрова
type LinearizedTableSerializer struct{}

func (s *LinearizedTableSerializer) Serialize(table *goquery.Selection) string {
	headers := extractHeaders(table)
	rows := dataRows(table)
	if len(rows) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, tr := range rows {
		var cells []string
		tr.Find("td").Each(func(j int, td *goquery.Selection) {
			cell := strings.TrimSpace(td.Text())
			if j < len(headers) && headers[j] != "" {
				cell = headers[j] + ": " + cell
			}
			cells = append(cells, cell)
		})
		if len(cells) == 0 {
			continue
		}
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(strings.Join(cells, ". "))
	}

	return sb.String()
}
