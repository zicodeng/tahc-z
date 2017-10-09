package handlers

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// PreviewImage represents a preview image for a page.
type PreviewImage struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

// PreviewVideo represents a preview video for a page.
type PreviewVideo struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

// PageSummary represents summary properties for a web page.
type PageSummary struct {
	Type        string          `json:"type,omitempty"`
	URL         string          `json:"url,omitempty"`
	Title       string          `json:"title,omitempty"`
	SiteName    string          `json:"siteName,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	Keywords    []string        `json:"keywords,omitempty"`
	Icon        *PreviewImage   `json:"icon,omitempty"`
	Images      []*PreviewImage `json:"images,omitempty"`
	Videos      []*PreviewVideo `json:"videos,omitempty"`
}

// SummaryHandler handles requests for the page summary API.
// This API expects one query string parameter named `url`,
// which should contain a URL to a web page. It responds with
// a JSON-encoded PageSummary struct containing the page summary
// meta-data.
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	// Get the `url` query string parameter value from the request.
	// If not supplied, respond with an http.StatusBadRequest error.
	pageURL := r.URL.Query().Get("q")
	if len(pageURL) == 0 {
		http.Error(w, "no query found in the requested URL", http.StatusBadRequest)
		return
	}

	// Call fetchHTML() to fetch the requested URL.
	htmlStream, err := fetchHTML(pageURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching HTML: %v\n", err), http.StatusBadRequest)
		return
	}
	// Close the response HTML stream so that you don't leak resources.
	defer htmlStream.Close()

	// Call extractSummary() to extract the page summary meta-data.
	pageSummary, err := extractSummary(pageURL, htmlStream)
	if err != nil {
		http.Error(w, fmt.Sprintf("error extracting summary: %v", err), http.StatusBadRequest)
		return
	}

	// Add an HTTP header to the response with the name
	// "Access-Control-Allow-Origin" and a value of "*".
	// This will allow cross-origin AJAX requests to your server.
	w.Header().Add(headerAccessControlAllowOrigin, "*")

	w.Header().Add(headerContentType, contentTypeJSON)

	// Finally, respond with a JSON-encoded version of the PageSummary
	// struct. That way the client can easily parse the JSON back into
	// an object.
	json.NewEncoder(w).Encode(pageSummary)
}

// fetchHTML fetches `pageURL` and returns the body stream or an error.
// Errors are returned if the response status code is an error (>=400),
// or if the content type indicates the URL is not an HTML page.
func fetchHTML(pageURL string) (io.ReadCloser, error) {

	// Do an HTTP GET for the page URL.
	res, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("fetching URL failed: %v", err)
	}

	// If the response status code is >= 400, return a nil stream and an error.
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("response status code was %d", res.StatusCode)
	}

	// If the response content type does not indicate that the content is a web page,
	// return a nil stream and an error.
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		return res.Body, fmt.Errorf("response content type was %s, not text/html", contentType)
	}

	return res.Body, nil
}

// extractSummary tokenizes the `htmlStream` and populates a PageSummary
// struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {

	pageSummary := &PageSummary{}
	previewImages := []*PreviewImage{}
	previewImage := &PreviewImage{}
	previewVideos := []*PreviewVideo{}
	previewVideo := &PreviewVideo{}

	// If description or title is set by basic name property,
	// Twitter Card will have priority.
	// If they are set by Open Graph,
	// Twitter Card will lose priority.
	twitterTitlePriority := false
	twitterDescPriority := false

	// Create a new tokenizer over the response body.
	tokenizer := html.NewTokenizer(htmlStream)

	// Loop through tokens to find page summary data
	// until we encounter error token, EOF, or </head>.
	for {
		// Get the next token type.
		tokenType := tokenizer.Next()

		tagName, hasAttr := tokenizer.TagName()
		tagNameStr := fmt.Sprintf("%s", tagName)

		switch tokenType {

		// If it is an error token, we either reached
		// the end of the file, or the HTML was malformed.
		case html.ErrorToken:
			return pageSummary, tokenizer.Err()

		// Because all page summary related tags only live within <head> section,
		// we can stop tokenizing if we encounter </head>.
		case html.EndTagToken:
			if tagNameStr == "head" {
				return pageSummary, nil
			}

		default:
			if tagNameStr == "meta" || tagNameStr == "link" || tagNameStr == "title" {
				if hasAttr {

					var attrKey []byte
					var attrVal []byte
					moreAttr := true

					m := make(map[string]string)

					// Loop through all attributes of the tag,
					// and store them in a map.
					for moreAttr {
						attrKey, attrVal, moreAttr = tokenizer.TagAttr()

						// Convert []byte to string.
						attrKeyStr := fmt.Sprintf("%s", attrKey)
						attrValStr := fmt.Sprintf("%s", attrVal)

						m[attrKeyStr] = attrValStr
					}

					// Look through Open Graph properties.
					// If not found, fall back to Twitter Card.
					switch property := m["property"]; property {
					case "og:type":
						pageSummary.Type = m["content"]

					case "og:url":
						pageSummary.URL = m["content"]

					case "og:title":
						twitterTitlePriority = false
						pageSummary.Title = m["content"]

					case "og:site_name":
						pageSummary.SiteName = m["content"]

					case "og:description":
						twitterDescPriority = false
						pageSummary.Description = m["content"]

					// Preview images.
					// og:image or og:iamge:url indicates this is a new preview image.
					case "og:image", "og:image:url":
						// Create a new instance of PreviewImage.
						previewImage = &PreviewImage{}

						previewImage.URL = fixURL(pageURL, m["content"])
						previewImages = append(previewImages, previewImage)
						pageSummary.Images = previewImages

					case "og:image:secure_url":
						previewImage.SecureURL = fixURL(pageURL, m["content"])

					case "og:image:type":
						previewImage.Type = m["content"]

					case "og:image:width", "og:image:height":
						size, err := strconv.Atoi(m["content"])
						if err != nil {
							return nil, fmt.Errorf("error converting string to int: %v", err)
						}
						if property == "og:image:width" {
							previewImage.Width = size
						} else {
							previewImage.Height = size
						}

					case "og:image:alt":
						previewImage.Alt = m["content"]

					// Preview videos.
					case "og:video":
						// Create a new instance of PreviewVideo.
						previewVideo = &PreviewVideo{}

						previewVideo.URL = fixURL(pageURL, m["content"])
						previewVideos = append(previewVideos, previewVideo)
						pageSummary.Videos = previewVideos

					case "og:video:secure_url":
						previewVideo.SecureURL = fixURL(pageURL, m["content"])

					case "og:video:type":
						previewVideo.Type = m["content"]

					case "og:video:width", "og:video:height":
						size, err := strconv.Atoi(m["content"])
						if err != nil {
							return nil, fmt.Errorf("error converting string to int: %v", err)
						}
						if property == "og:video:width" {
							previewVideo.Width = size
						} else {
							previewVideo.Height = size
						}

					case "og:video:alt":
						previewVideo.Alt = m["content"]

					// Add Twitter Card fallback support.
					case "twitter:card":
						if pageSummary.Type == "" {
							pageSummary.Type = m["content"]
						}

					case "twitter:description":
						if pageSummary.Description == "" || twitterDescPriority {
							pageSummary.Description = m["content"]
						}

					case "twitter:title":
						if pageSummary.Title == "" || twitterTitlePriority {
							pageSummary.Title = m["content"]
						}

					case "twitter:image":
						// Add a new instance of PreviewImage only if
						// the previewImages slice doesn't contain an element for that same image URL.
						contains := false
						for _, image := range previewImages {
							if image.URL == m["content"] {
								contains = true
							}
						}
						if !contains {
							previewImage = &PreviewImage{}

							previewImage.URL = fixURL(pageURL, m["content"])
							previewImages = append(previewImages, previewImage)
							pageSummary.Images = previewImages
						}
					}

					// Look through basic <meta> tag name property.
					switch m["name"] {
					case "description":
						if pageSummary.Description == "" {
							twitterDescPriority = true
							pageSummary.Description = m["content"]
						}

					case "author":
						pageSummary.Author = m["content"]

					case "keywords":
						pageSummary.Keywords = regexp.MustCompile(",\\s*").Split(m["content"], -1)
					}

					// Handle <link> tag.
					if m["rel"] == "icon" {
						icon := &PreviewImage{
							URL:  fixURL(pageURL, m["href"]),
							Type: m["type"],
						}

						i := strings.Index(m["sizes"], "x")
						if i != -1 {
							width, err := strconv.Atoi(m["sizes"][i+1:])
							if err != nil {
								return nil, fmt.Errorf("error converting string to int: %v", err)
							}

							height, err := strconv.Atoi(m["sizes"][:i])
							if err != nil {
								return nil, fmt.Errorf("error converting string to int: %v", err)
							}

							icon.Width = width
							icon.Height = height
						}

						pageSummary.Icon = icon
					}
				} else {
					// If no attribute found, it implies this is a <title> tag.

					// Get the next token type which should be text with the page's title.
					tokenType = tokenizer.Next()

					// Just make sure it is actually a text token.
					if tokenType == html.TextToken {
						title := tokenizer.Token().Data
						if pageSummary.Title == "" {
							twitterTitlePriority = true
							pageSummary.Title = title
						}
					}
				}
			}
		}
	}
}

// If the URL is a relative URL e.g. /resource/path,
// fixURL converts it to an absolute URL e.g. http://www.example.com/resource/path.
func fixURL(pageURL, URL string) string {

	site := regexp.MustCompile("https?://[\\w.-]+").FindString(pageURL)

	if !strings.HasPrefix(URL, "http") {
		return site + URL
	}

	return URL
}
