package techo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {

	te := New()
	defer te.Stop()
	te.GET("/hello", func(c echo.Context) error {
		param := c.QueryParam("name")
		assert.Equal(t, param, "world")
		return c.String(http.StatusOK, fmt.Sprintf("hello %v", param))
	})

	resp, err := http.Get(te.AbsURL("/hello?name=world"))
	defer resp.Body.Close()
	require.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "hello world", string(body))
}

func TestNewWith(t *testing.T) {

	// So, we want to test that we can start a server (using NewAt) at a specific address,
	// e.g. localhost:1234. But how do we know what port will be free for our unit test?

	// Here's our strategy:

	// 1. Start a server
	te := New()

	// 2. Save that random port
	port := te.Port

	// 3. Stop the server. The port will (hopefully!) be available for a little while.
	//    This is definitely not guaranteed, don't do this for your nuclear control system.
	te.Stop()

	// 4. Here's the function we want to test... start a server on that port
	te2, err := NewWith(&Config{Addr: fmt.Sprintf("localhost:%v", port)})

	require.Nil(t, err) // 5. w00t, it worked!
	require.NotNil(t, te2)
	require.Equal(t, port, te2.Port)

	// 6. Let's just be paranoid and make sure we actually can get back content
	te2.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	require.Equal(t, fmt.Sprintf("http://127.0.0.1:%v/hello", te2.Port), te2.AbsURL("/hello"))

	resp, err := http.Get(te2.AbsURL("/hello"))
	defer resp.Body.Close()
	require.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "hello world", string(body))

	te2.Stop()
	// Let's try with TLS

	te3, err := NewWith(&Config{Addr: fmt.Sprintf("localhost:%v", port), TLS: true})

	require.Nil(t, err)
	require.NotNil(t, te3)
	require.Equal(t, port, te3.Port)

	// 6. Let's just be paranoid and make sure we actually can get back content
	te3.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	// Disable client cert checking
	SkipDefaultClientInsecureTLSVerify()
	resp, err = http.Get(te3.AbsURL("/hello"))
	require.Nil(t, err)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "hello world", string(body))

	te3.Stop()

}

func TestNewTLS(t *testing.T) {

	te := NewTLS()
	defer te.Stop()
	te.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	SkipDefaultClientInsecureTLSVerify()
	resp, err := http.Get(te.AbsURL("/hello"))
	defer resp.Body.Close()
	require.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "hello world", string(body))
}

func TestTLSWithUserCerts(t *testing.T) {

	SetDefaultTLSCert(testCert, testKey)

	te := NewTLS()
	defer te.Stop()
	te.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	SkipDefaultClientInsecureTLSVerify()
	resp, err := http.Get(te.AbsURL("/hello"))
	defer resp.Body.Close()
	require.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "hello world", string(body))
}

var testCert = []byte(`-----BEGIN CERTIFICATE-----
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

var testKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
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
