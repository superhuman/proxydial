package proxydial

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

var client = http.Client{
	Transport: &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

func TestDialer(t *testing.T) {

	response, err := client.Get("http://proxydial.herokuapp.com/remote")
	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		t.Fatal(err)
	}

	if string(bytes) != "DONE" {
		t.Fatal(bytes)
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
	_, err := client.Get("http://10.1.1.2/")
	if err == nil || err.Error() != "Get http://10.1.1.2/: dialer.Dial 10.1.1.2:80: blocked range (10.1.1.2)" {
		t.Error(err)
	}
}

func TestBlockedIPv6(t *testing.T) {
	_, err := client.Get("http://[fe80::1]/")
	if err == nil || err.Error() != "Get http://[fe80::1]/: dialer.Dial [fe80::1]:80: blocked range (fe80::1)" {
		t.Error(err)
	}
}

func TestRedirects(t *testing.T) {

	tests := map[string]string{
		"https://proxydial.herokuapp.com/file":    "unsupported protocol scheme",
		"https://proxydial.herokuapp.com/local":   "blocked range",
		"https://proxydial.herokuapp.com/v6":      "blocked range",
		"https://proxydial.herokuapp.com/port":    "blocked port",
		"https://proxydial.herokuapp.com/recurse": "stopped after 10 redirects",
	}

	for url, errmsg := range tests {
		_, err := client.Get(url)

		if err == nil {
			t.Errorf("successfully fetched %s", url)

		}
		if !strings.Contains(err.Error(), errmsg) {
			t.Error(err)
		}
	}
}

func TestResolutionFailure(t *testing.T) {
	_, err := client.Get("https://fails.to.resolve.lllllllllllll")
	if err == nil || err.Error() != "Get https://fails.to.resolve.lllllllllllll: lookup fails.to.resolve.lllllllllllll: no such host" {
		t.Error(err)
	}
}
