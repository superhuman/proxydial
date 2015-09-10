package proxydialer

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

type httpbinGetResponse struct {
	Args    map[string]interface{} `json:"args"`
	Headers map[string]string      `json:"headers"`
	Origin  string                 `json:"origin"`
	URL     string                 `json:"url"`
}

var client = http.Client{
	Transport: &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

func TestDialer(t *testing.T) {

	response, err := client.Get("https://httpbin.org/get")
	if err != nil {
		t.Fatal(err)
	}

	body := httpbinGetResponse{}
	json.NewDecoder(response.Body).Decode(&body)
	defer response.Body.Close()

	if body.URL != "https://httpbin.org/get" {
		t.Fatalf("body.Url %s ", body.Url)
	}
}

func TestBlockedPorts(t *testing.T) {
	_, err := client.Get("https://httpbin.org:25/get")
	if err == nil || err.Error() != "Get https://httpbin.org:25/get: dialer.Dial httpbin.org:25: blocked port" {
		t.Fatal(err)
	}
}

func TestLocalHost(t *testing.T) {
	_, err := client.Get("http://localhost/")
	if err == nil || err.Error() != "Get http://localhost/: dialer.Dial localhost:80: blocked range (::1)" {
		t.Error(err)
	}
}

func TestBlockedIPv4(t *testing.T) {

}

func TestRedirects(t *testing.T) {

}

func TestBlockedIPv6(t *testing.T) {

}

func TestTimeouts(t *testing.T) {

}

func TestResolutionFailure(t *testing.T) {
	_, err := client.Get("https://fails.to.resolve.lllllllllllll")
	if err == nil || err.Error() != "Get https://fails.to.resolve.lllllllllllll: lookup fails.to.resolve.lllllllllllll: no such host" {
		t.Error(err)
	}
}
