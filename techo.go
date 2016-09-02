/*
Package techo is equivalent to http.httptest, but uses labstack/echo for greater
ease of use, and provides several additional useful things.

For more, visit: https://github.com/neilotoole/techo

Example:

	func TestHello(t *testing.T) {

		te := techo.New()
		defer te.Stop()
		te.GET("/hello", func(c echo.Context) error {
			param := c.QueryParam("name")
			assert.Equal(t, param, "World")
			return c.String(http.StatusOK, fmt.Sprintf("Hello %v", param))
		})

		resp, err := http.Get(te.AbsURL("/hello?name=World"))
		defer resp.Body.Close()
		require.Nil(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)
		assert.Equal(t, "Hello World", string(body))
	}
*/
package techo

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"time"

	"io/ioutil"
	"os"

	"sync"

	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/tylerb/graceful"
)

// Techo is a techo server instance.
type Techo struct {
	// Port is the port number the server is listening at.
	Port int
	// Base is the base URL (scheme + host + port), e.g. http://127.0.0.1:61241
	URL string
	// Addr provides access to the underlying TCP address object.
	Addr *net.TCPAddr
	*echo.Echo
	srv *graceful.Server

	certFilePath string
	keyFilePath  string
	mutex        *sync.Mutex
}

// Config is the options available for staring a techo instance with techo.NewWith().
type Config struct {
	// Addr is the address to listen on, e.g. ":1234" or "localhost:8080".
	Addr string
	// TLS indicates to start a TLS/HTTPS server.
	TLS bool
	// TLSCert is the TLS certificate to use.
	TLSCert []byte
	// TLSKey is the TLS private key to use.
	TLSKey []byte
}

// New starts a server on any available port. This value is available in the Port field.
// In the unlikely event of an error, the error is logged, and nil is returned.
func New() *Techo {
	te, err := listenAndStart("localhost:")
	if err != nil {
		log.Println(err)
	}
	return te
}

// NewWith starts a server using the supplied config.
func NewWith(cfg *Config) (*Techo, error) {
	if cfg.TLS == false {
		if cfg.Addr == "" {
			return listenAndStart("localhost:")
		}
		return listenAndStart(cfg.Addr)
	}

	// cfg.TLS == true
	cert := defaultCert
	key := defaultKey

	if len(cfg.TLSCert) > 0 {
		cert = cfg.TLSCert
	}

	if len(cfg.TLSKey) > 0 {
		cert = cfg.TLSKey
	}

	if cfg.Addr == "" {
		return listenAndStartTLS("localhost:", cert, key)
	}

	return listenAndStartTLS(cfg.Addr, cert, key)
}

func listenAndStart(addr string) (*Techo, error) {

	t := new(Techo)
	t.Echo = echo.New()
	t.mutex = &sync.Mutex{}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	t.Addr = l.Addr().(*net.TCPAddr)
	t.Port = t.Addr.Port
	t.URL = fmt.Sprintf("http://%v:%v", t.Addr.IP, t.Port)
	std := standard.New(fmt.Sprintf(":%v", t.Addr.Port))
	std.SetHandler(t.Echo)
	t.srv = &graceful.Server{
		Timeout: time.Millisecond * 1,
		Server:  std.Server,
	}

	go func() {
		err := t.srv.Serve(l)
		if err != nil {
			log.Printf("techo error: %v\n", err)
		}
	}()

	return t, nil
}

// NewTLS starts a TLS/HTTPS server on a random port. In the unusual event of an error,
// the error is logged, and nil is returned.
func NewTLS() *Techo {

	te, err := listenAndStartTLS("localhost:", defaultCert, defaultKey)
	if err != nil {
		log.Println(err)
		return nil
	}
	return te
}

func listenAndStartTLS(addr string, tlsCert []byte, tlsKey []byte) (*Techo, error) {

	t := new(Techo)
	t.Echo = echo.New()
	t.mutex = &sync.Mutex{}

	err := t.writeTLSFiles(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}

	std := standard.WithTLS(addr, t.certFilePath, t.keyFilePath)
	std.SetHandler(t.Echo)

	t.srv = &graceful.Server{
		Timeout: time.Millisecond * 1,
		Server:  std.Server,
	}

	l, err := t.srv.ListenTLS(t.certFilePath, t.keyFilePath)

	if err != nil {
		return nil, err
	}

	t.Addr = l.Addr().(*net.TCPAddr)
	t.Port = t.Addr.Port
	t.URL = fmt.Sprintf("https://%v:%v", t.Addr.IP, t.Port)

	go func() {
		err := t.srv.Serve(l)
		if err != nil {
			log.Printf("techo error: %v\n", err)
		}
		t.cleanupTLSFiles()
	}()

	return t, nil
}

// writeTLSFiles writes out the cert and key files required when using TLS. It is
// necessary to write these to disk (as opposed to providing the bytes directly)
// as the echo API requires these files be loaded from disk.
func (t *Techo) writeTLSFiles(cert []byte, key []byte) error {

	t.mutex.Lock()
	defer t.mutex.Unlock()
	certFile, err := ioutil.TempFile("", "techo-tls-cert_")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(certFile.Name(), cert, os.ModePerm)
	if err != nil {
		return err
	}

	keyFile, err := ioutil.TempFile("", "techo-tls-key_")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(keyFile.Name(), key, os.ModePerm)
	if err != nil {
		return err
	}

	t.certFilePath = certFile.Name()
	t.keyFilePath = keyFile.Name()

	return nil

}

// cleanupTLSFiles attempts to delete the temporary TLS files created by tech.
// Errors are logged but not returned.
func (t *Techo) cleanupTLSFiles() {

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.certFilePath != "" {
		err := os.Remove(t.certFilePath)
		if err != nil {
			log.Println(err)
		}
		t.certFilePath = ""
	}
	if t.keyFilePath != "" {
		err := os.Remove(t.keyFilePath)
		if err != nil {
			log.Println(err)
		}
		t.keyFilePath = ""
	}

}

// Stop instructs the server to shut down.
func (t *Techo) Stop() {
	t.srv.Stop(time.Millisecond * 1)
	t.cleanupTLSFiles()
}

func (t *Techo) String() string {
	return t.URL
}

// AbsURL constructs an absolute URL from the supplied (relative) path. For example,
// calling te.AbsURL("/my/path") could return "http://127.0.0.1:53262/my/path".
func (t *Techo) AbsURL(path string) string {

	if len(path) == 0 {
		return t.URL
	}

	if path[0] == '/' {
		return t.URL + path
	}

	return t.URL + "/" + path
}

var defaultCert []byte
var defaultKey []byte

func init() {
	defaultCert = localhostCert
	defaultKey = localhostKey
}

// SetDefaultTLSCert is used to specify the TLS cert/key used by NewTLS().
// Set the params to nil to restore to the internal default (localhost) cert.
func SetDefaultTLSCert(cert []byte, key []byte) {

	if cert == nil {
		defaultCert = localhostCert
	} else {
		defaultCert = cert
	}
	if key == nil {
		defaultKey = localhostKey
	} else {
		defaultKey = key
	}

}

// SkipDefaultClientInsecureTLSVerify is a convenience method that sets
// InsecureSkipVerify to true on http.DefaultClient. This means that you can use
// insecure certs without receiving an error (assuming your client is using
// http.DefaultClient).
func SkipDefaultClientInsecureTLSVerify() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	http.DefaultClient.Transport = tr
}

// NOTE: copied from http.httptest.internal

// localhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`)

// localhostKey is the private key for localhostCert.
var localhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
-----END RSA PRIVATE KEY-----`)
