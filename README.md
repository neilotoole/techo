# techo

###Why *Fake Mock* when you can *Real Mock™*?

`techo` is a package for transparently mocking HTTP/REST servers, inline in your
test code. As far as the code under test is concerned, the mocked server is
a "real" server because... well, it is a real server. An [Echo](https://github.com/labstack/echo) server to be precise, running on a separate goroutine on any available port.

Here's how you might use the thing:

```go
func TestHello(t *testing.T) {

	te := techo.New() // start the web server - it's running on some random port now
	defer te.Stop() // stop the server at the end of this function
	
	// just for fun, check what random port we're running on
	fmt.Println("My port: ", te.Port)
	
	// serve up some content (using echo, cuz it's so hot right now)
	te.GET("/hello", func(c echo.Context) error {
		param := c.QueryParam("name")
		assert.Equal(t, param, "World") // assert some stuff
		return c.String(http.StatusOK, fmt.Sprintf("Hello %v", param))
	})

    // Let's call the web server!
	resp, _ := http.Get(te.URL("/hello?name=World"))
	// Note that te.URL() turns "/hello" into "http://127.0.0.1:[PORT]/hello"
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "Hello World", string(body))
}
```

Referencing the example above, here's a few things that are going on:

- The `te` object (of type `techo.Techo`) returned by`techo.New()` has an embedded `Echo` server.
    - So you can treat `techo.Techo` just like `echo.Echo`, and do things like `te.POST( ...)` etc.
- When `techo.New()` is called, the server is started automatically...
    - on its own goroutine (transparently to the caller)
    - on a random available port
- The port number is here: `te.Port`.
- The base URL of the server (`scheme://host:port`) is here: `te.Base`
- There's a handy function for getting the URL of a path on the techo server. Just do `te.URL('/my/path')`, and you'll get back something like `http://127.0.0.1:52713/my/path`.
- Stop the server using `te.Stop()`. A common idiom is `defer te.Stop()` immediately after the call to `techo.New()`. FYI, the stoppability is due a (hidden) [Graceful](https://github.com/tylerb/graceful) server.


### Acknowledgements

These guys do all the magic:

* [Echo](https://github.com/labstack/echo)
* [Graceful](https://github.com/tylerb/graceful)
* [Testify](https://github.com/stretchr/testify)
