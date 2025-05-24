package charNorm

import (
	"fmt"
	"net/http"
	"strings"
	"golang.org/x/net/html"
	"errors"
)

// ParseNormalizationTable fetches an HTML table from a URL and parses it
// to create a map of ASCII characters to their Unicode normalizations.
func ParseNormalizationTable(url string) (map[rune][]rune, error) {
	if url == "" {
		return nil, errors.New("url cannot be empty")
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL %s: status code %d", url, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML from %s: %w", url, err)
	}

	normalizationMap := make(map[rune][]rune)
	var originalChars []rune
	var tableFound bool
	var rowCount int

	// findTableAndParse is a recursive function to find the table and parse it.
	var findTableAndParse func(*html.Node)
	findTableAndParse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tableFound = true
			// Traverse rows of the table
			for tr := n.FirstChild; tr != nil; tr = tr.NextSibling {
				if tr.Type == html.ElementNode && tr.Data == "tbody" { // HTML often implicitly adds tbody
					for r := tr.FirstChild; r != nil; r = r.NextSibling {
						if r.Type == html.ElementNode && r.Data == "tr" {
							rowCount++
							var currentCellIndex int
							if rowCount == 1 { // Skip first row (hex codes)
								continue
							} else if rowCount == 2 { // Second row: original ASCII characters
								for td := r.FirstChild; td != nil; td = td.NextSibling {
									if td.Type == html.ElementNode && td.Data == "td" {
										charStr := strings.TrimSpace(extractText(td))
										if len(charStr) > 0 {
											// Assuming single rune per cell for original chars
											originalChars = append(originalChars, []rune(charStr)[0])
											normalizationMap[[]rune(charStr)[0]] = []rune{} // Initialize slice
										} else {
											originalChars = append(originalChars, ' ') // Placeholder for empty cells to maintain index
										}
									}
								}
							} else if rowCount >= 3 && rowCount <= 62 { // Data rows (3rd to 62nd)
								for td := r.FirstChild; td != nil; td = td.NextSibling {
									if td.Type == html.ElementNode && td.Data == "td" {
										if currentCellIndex < len(originalChars) {
											normCharStr := strings.TrimSpace(extractText(td))
											if len(normCharStr) > 0 {
												normRune := []rune(normCharStr)[0]
												originalChar := originalChars[currentCellIndex]
												
												// Add if not already present
												found := false
												for _, existingRune := range normalizationMap[originalChar] {
													if existingRune == normRune {
														found = true
														break
													}
												}
												if !found {
													normalizationMap[originalChar] = append(normalizationMap[originalChar], normRune)
												}
											}
										}
										currentCellIndex++
									}
								}
							}
						}
					}
				} else if tr.Type == html.ElementNode && tr.Data == "tr" { // Handle tables without explicit tbody
					rowCount++
					var currentCellIndex int
					if rowCount == 1 { // Skip first row (hex codes)
						continue
					} else if rowCount == 2 { // Second row: original ASCII characters
						for td := tr.FirstChild; td != nil; td = td.NextSibling {
							if td.Type == html.ElementNode && td.Data == "td" {
								charStr := strings.TrimSpace(extractText(td))
								if len(charStr) > 0 {
									originalChars = append(originalChars, []rune(charStr)[0])
									normalizationMap[[]rune(charStr)[0]] = []rune{} 
								} else {
									originalChars = append(originalChars, ' ') 
								}
							}
						}
					} else if rowCount >= 3 && rowCount <= 62 { // Data rows
						for td := tr.FirstChild; td != nil; td = td.NextSibling {
							if td.Type == html.ElementNode && td.Data == "td" {
								if currentCellIndex < len(originalChars) {
									normCharStr := strings.TrimSpace(extractText(td))
									if len(normCharStr) > 0 {
										normRune := []rune(normCharStr)[0]
										originalChar := originalChars[currentCellIndex]
										
										found := false
										for _, existingRune := range normalizationMap[originalChar] {
											if existingRune == normRune {
												found = true
												break
											}
										}
										if !found {
											normalizationMap[originalChar] = append(normalizationMap[originalChar], normRune)
										}
									}
								}
								currentCellIndex++
							}
						}
					}
				}
			}
			return // Stop searching after finding and processing the first table
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if tableFound { // If table is processed, no need to traverse further
				return
			}
			findTableAndParse(c)
		}
	}

	// extractText recursively extracts all text from a node and its children.
	var extractText func(*html.Node) string
	extractText = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		if n.Type != html.ElementNode {
			return ""
		}
		var ret string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			ret += extractText(c)
		}
		return ret
	}

	findTableAndParse(doc)

	if !tableFound {
		return nil, errors.New("no table found in HTML")
	}
	if rowCount < 62 {
		// It's possible some tables might be shorter, but the spec implies 62 rows are key.
		// Depending on strictness, this could be an error or a warning.
		// For now, returning an error if not enough rows for the full specified normalization data.
		return nil, fmt.Errorf("table found, but did not contain enough rows (expected at least 62, got %d)", rowCount)
	}
	
	// Remove any original characters that ended up with no normalizations
	// (e.g. if they were placeholders for empty cells in the second row)
	for key, val := range normalizationMap {
		if key == ' ' && len(val) == 0 { // Check for placeholder space character
			delete(normalizationMap, key)
		}
	}


	return normalizationMap, nil
}
