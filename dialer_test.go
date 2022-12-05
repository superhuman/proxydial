package proxydial

import (
	"fmt"
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

	// TODO: remove network request
	response, err := client.Get("http://example.com/")
	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(bytes), "<html>") {
		t.Fatal(string(bytes))
	}
}

func TestBlockedPorts(t *testing.T) {
	_, err := client.Get("https://httpbin.org:25/get")
	if err == nil || err.Error() != `Get "https://httpbin.org:25/get": dialer.Dial httpbin.org:25: blocked port` {
		t.Fatal(err)
	}
}

func TestLocalHost(t *testing.T) {
	_, err := client.Get("http://localhost/")
	if err == nil || err.Error() != `Get "http://localhost/": dialer.Dial localhost:80: blocked range (::1)` {
		t.Error(err)
	}
}

func TestBlockedIPv4(t *testing.T) {
	_, err := client.Get("http://10.1.1.2/")
	if err == nil || err.Error() != `Get "http://10.1.1.2/": dialer.Dial 10.1.1.2:80: blocked range (10.1.1.2)` {
		t.Error(err)
	}
}

func TestBlockedIPv6(t *testing.T) {
	_, err := client.Get("http://[fe80::1]/")
	if err == nil || err.Error() != `Get "http://[fe80::1]/": dialer.Dial [fe80::1]:80: blocked range (fe80::1)` {
		t.Error(err)
	}
}

func TestBlocks(t *testing.T) {
	for _, ip := range []string{
		"169.254.0.1",
		"[::]",
		"0x7f000001",           // hex
		"2130706433",           // decimal
		"000127.0.00000.00001", // leading zeros
	} {
		requiredPrefix := fmt.Sprintf(`Get "http://%s/": dialer.Dial %s:80: blocked`, ip, ip)
		_, err := client.Get(fmt.Sprintf("http://%s/", ip))

		if err == nil || !strings.HasPrefix(err.Error(), requiredPrefix) {
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
