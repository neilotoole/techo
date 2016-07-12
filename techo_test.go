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
		assert.Equal(t, param, "World")
		return c.String(http.StatusOK, fmt.Sprintf("Hello %v", param))
	})

	resp, err := http.Get(te.URL("/hello?name=World"))
	defer resp.Body.Close()
	require.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "Hello World", string(body))
}
