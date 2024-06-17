package twitter

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"
)

// Default headers used in all requests to the Twitter API.
var headers = map[string]string{
	"Authorization":             "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
	"x-twitter-client-language": "en",
	"x-twitter-active-user":     "yes",
	"Accept-language":           "en",
	"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"content-type":              "application/json",
}

// Constants for Twitter API endpoints and regular expression for extracting tweet IDs from URLs.
const (
	tweetDetail      = "https://twitter.com/i/api/graphql/5GOHgZe-8U2j5sVHQzEm9A/TweetResultByRestId"
	userByScreenName = "https://api.twitter.com/graphql/qW5u-DAuXpMEG0zA1F7UGQ/UserByScreenName"
	tweetIDRegex     = `.*(?:twitter|x).com/.+status/([A-Za-z0-9]+)`
)

// requestParams contains parameters for making HTTP requests.
type requestParams struct {
	Method     string            // "GET", "OPTIONS" or "POST"
	Headers    map[string]string // Common headers for both GET and POST
	Query      map[string]string // Query parameters for GET
	BodyString []string          // Body of the request for POST
}

// Request sends a GET, OPTIONS or POST request to the specified link with the provided parameters and returns the response.
// The Link specifies the URL to send the request to.
// The params contain additional parameters for the request, such as headers, query parameters, and body.
// The Method field in params should be "GET" or "POST" to indicate the type of request.
//
// Example usage:
//
//	response := Request("https://api.example.com/users", RequestParams{
//		Method: "GET",
//		Headers: map[string]string{
//			"Authorization": "Bearer your-token",
//		},
//		Query: map[string]string{
//			"page":  "1",
//			"limit": "10",
//		},
//	})
//
//	response := Request("https://example.com/api", RequestParams{
//		Method: "POST",
//		Headers: map[string]string{
//			"Content-Type": "application/json",
//		},
//		BodyString: []string{
//			"param1=value1",
//			"param2=value2",
//		},
//	})
func request(Link string, params requestParams) *fasthttp.Response {
	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	client := &fasthttp.Client{
		ReadBufferSize:  16 * 1024,
		MaxConnsPerHost: 1024,
	}

	request.Header.SetMethod(params.Method)
	for key, value := range params.Headers {
		request.Header.Set(key, value)
	}

	if params.Method == fasthttp.MethodGet {
		request.SetRequestURI(Link)
		for key, value := range params.Query {
			request.URI().QueryArgs().Add(key, value)
		}
	} else if params.Method == fasthttp.MethodOptions {
		request.SetRequestURI(Link)
		for key, value := range params.Query {
			request.URI().QueryArgs().Add(key, value)
		}
	} else if params.Method == fasthttp.MethodPost {
		request.SetBodyString(strings.Join(params.BodyString, "&"))
		request.SetRequestURI(Link)
	} else {
		log.Print("[request/Request] Error: Unsupported method ", params.Method)
		return response
	}

	err := client.Do(request, response)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			return response
		}
		log.Print("[request/Request] Error: ", err)
	}

	return response
}

// getGuestToken retrieves a guest token from the Twitter API and returns it as a string.
// It sends a POST request to the "https://api.twitter.com/1.1/guest/activate.json" endpoint and parses the response.
func getGuestToken() string {
	type guestToken struct {
		GuestToken string `json:"guest_token"`
	}

	body := request("https://api.twitter.com/1.1/guest/activate.json", requestParams{
		Method:  "POST",
		Headers: headers,
	}).Body()

	var res guestToken
	err := json.Unmarshal(body, &res)
	if err != nil {
		log.Printf("Error unmarshalling guest token: %v", err)
	}

	return res.GuestToken
}

// Helper function to marshal data into JSON string.
func marshalJSON(data interface{}) string {
	result, _ := json.Marshal(data)
	return string(result)
}

// Helper function to extract Tweet ID from URL.
func extractTweetID(url string) (string, error) {
	matches := regexp.MustCompile(tweetIDRegex).FindStringSubmatch(url)
	if len(matches) != 2 {
		return "", fmt.Errorf("invalid tweet URL: %v", url)
	}
	return matches[1], nil
}
