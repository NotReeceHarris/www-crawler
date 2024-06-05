package main

import (
	"strings"
	//"fmt"
	"net/url"

	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
)

func parseURL(inputURL string) (string, string, string, error) {
    u, err := url.Parse(inputURL)
    if err != nil {
        return "", "", "", err
    }
    scheme := u.Scheme
    domain := u.Hostname()
    path := u.Path
    return scheme, domain, path, nil
}

func get(inputURL string, pathID int) ([]string, error) {
    req := fasthttp.AcquireRequest()
    resp := fasthttp.AcquireResponse()
    defer fasthttp.ReleaseRequest(req)
    defer fasthttp.ReleaseResponse(resp)

    req.SetRequestURI(inputURL)

    if err := fasthttp.Do(req, resp); err != nil {
        return nil, err
    }

    bodyBytes := resp.Body()
    httpCode := resp.Header.StatusCode()

    // Parse HTML and extract all links
    doc, err := html.Parse(strings.NewReader(string(bodyBytes)))
    if err != nil {
        return nil, err
    }

    links := make(map[string]bool) // create a set to store unique links

    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "a" {
            for _, a := range n.Attr {
                if a.Key == "href" {
                    // Check if the link is a valid URL and not a fragment, tel, or mailto link
                    u, err := url.Parse(a.Val)
                    if err == nil && u.Scheme != "" {
                        invalidSchemes := map[string]bool{
                            "mailto": true,
                            "tel":    true,
                            "#":      true,
                        }
                    
                        _, isInvalid := invalidSchemes[u.Scheme]
                        if !isInvalid && !strings.HasPrefix(a.Val, "#") {
                            links[a.Val] = true
                        }
                    }
                }
            }
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)

    markScanned(pathID, httpCode)

    // Convert the links map to a slice
    var linksSlice []string
    for link := range links {
        linksSlice = append(linksSlice, link)
    }

    return linksSlice, nil
}