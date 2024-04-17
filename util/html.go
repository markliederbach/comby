package util

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// HTMLToPlainText converts HTML content to plain text, ignoring <a href> links.
func HTMLToPlainText(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = traverseAndExtractText(doc, &buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// traverseAndExtractText traverses the HTML nodes and writes text nodes to the buffer.
// It skips <a> tags with href attributes.
func traverseAndExtractText(n *html.Node, buf *bytes.Buffer) error {
	if n.Type == html.TextNode {
		_, err := buf.WriteString(n.Data)
		if err != nil {
			return err
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		// Skip <a> tags with href attributes
		if c.Type == html.ElementNode && c.Data == "a" && hasHref(c) {
			continue
		}
		err := traverseAndExtractText(c, buf)
		if err != nil {
			return err
		}
	}

	return nil
}

// hasHref checks if an <a> element node has an href attribute.
func hasHref(n *html.Node) bool {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return true
		}
	}
	return false
}
