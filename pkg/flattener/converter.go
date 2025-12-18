package flattener

import (
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

// ContentConverter handles HTML to Markdown conversion.
type ContentConverter struct {
	mdConverter *md.Converter
}

// NewConverter creates a new ContentConverter.
func NewConverter() *ContentConverter {
	converter := md.NewConverter("", true, nil)
	converter.AddRules(
		md.Rule{
			Filter: []string{"details"},
			Replacement: func(content string, selec *goquery.Selection, options *md.Options) *string {
				// Unwrap details, just return content
				return &content
			},
		},
		md.Rule{
			Filter: []string{"summary"},
			Replacement: func(content string, selec *goquery.Selection, options *md.Options) *string {
				// Make summary a header
				text := strings.TrimSpace(content)
				newContent := fmt.Sprintf("\n#### %s\n\n", text)
				return &newContent
			},
		},
	)
	return &ContentConverter{mdConverter: converter}
}

// Convert converts a goquery selection to Markdown.
func (c *ContentConverter) Convert(selection *goquery.Selection) string {
	return c.mdConverter.Convert(selection)
}
