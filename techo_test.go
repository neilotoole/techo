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

func TestHello(t *testing.T) {

	te := New()
	defer te.Stop()
	te.GET("/hello", func(c echo.Context) error {
		param := c.QueryParam("name")
		assert.Equal(t, param, "world")
		return c.String(http.StatusOK, fmt.Sprintf("hello %v", param))
	})

	resp, err := http.Get(te.URL("/hello?name=world"))
	defer resp.Body.Close()
	require.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "hello world", string(body))
}

func TestNewAt(t *testing.T) {

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
	te2, err := NewAt(fmt.Sprintf("localhost:%v", port))

	require.Nil(t, err) // 5. w00t, it worked!
	require.NotNil(t, te2)
	require.Equal(t, port, te2.Port)

	// 6. Let's just be paranoid and make sure we actually can get back content
	te2.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	resp, err := http.Get(te.URL("/hello"))
	defer resp.Body.Close()
	require.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "hello world", string(body))
}
