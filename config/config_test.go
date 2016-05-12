package config

import (
	"crypto/tls"
	"fmt"
	"github.com/cldmnky/f5er/f5"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	username string
	host     string
	password string
)

func testTools(code int, body string) (*httptest.Server, *f5.Device) {

	username = "testuser"
	password = "testpass"
	host = "bigip"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		if r.URL.Path == "/mgmt/shared/authn/login" {
			w.WriteHeader(code)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"token": {"token": "123456789", "expirationMicros": 123456789}}`)
		} else {
			w.WriteHeader(code)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, body)
		}
	}))

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	appliance := f5.NewInsecure(host, username, password, f5.TOKEN)
	appliance.Session.Client.Transport = tr
	return server, appliance
}

func TestLoadConfigFile(t *testing.T) {
	_, err := LoadConfigFile("test/good.yml")
	if err != nil {
		t.Fatalf("Error parsing %s: %s", "test/good.yml", err)
	}
}

func TestGetVirtualServer(t *testing.T) {
	server, client := testTools(200, `
		{"fullPath": "/partition/virtual-server",
		 "rules": ["/Common/irule", "/partition/irule"]}
		`)
	defer server.Close()
	vs, err := GetVirtualServer("virtual-server", client)
	assert.Nil(t, err)
	assert.Equal(t, "/partition/virtual-server", vs.Name)
}
