# techo

###*Real Mocking*™ for HTTP/REST services.

`techo` is a package for transparently mocking HTTP servers, inline in your
test code. As far as the code under test is concerned, the mocked server is
a "real" server because... well, it is a real server. An [Echo](https://github.com/labstack/echo) server to be precise, running on a separate goroutine on any available port.

Here's how you might use the thing:

```go
func TestHello(t *testing.T) {

	te := techo.New()
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
```
The techo struct returned by `techo.New()` embeds an `Echo` server, which is started
automatically. This server happens to be wrapped in a [Graceful](https://github.com/tylerb/graceful) server, and you can (and should) shut it using Stop. As in the example above, it's often idiomatic to call `defer te.Stop()` immediately after invoking `New()`.


### Acknowledgements

These guys do all the magic:

* [Echo](https://github.com/labstack/echo)
* [Graceful](https://github.com/tylerb/graceful)
* [Testify](https://github.com/stretchr/testify)