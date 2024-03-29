# techo

### Echo-based alternative to `http.httptest`

`techo` is a package for transparently mocking HTTP/REST servers, inline in your
test code, just like [http.httptest](https://godoc.org/net/http/httptest). The key difference is that `techo` is based
on [Echo](https://github.com/labstack/echo), which makes writing tests cleaner and more expressive.

> **Note**
> The `techo` package was created way back in the early days of Echo. Since then there have been multiple
> major version releases of Echo (`v4` at the time of writing). This package hasn't been updated in that time,
> so it may not work with recent Echo releases. I may update it if I find myself working directly with Echo again,
> but in the meantime, please consider this package effectively archived.


Also:
- test multiple endpoints using the same `techo` instance
- specify a particular interface/port
- provide your own TLS certs

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

	// Let's make a request to the web server
	resp, _ := http.Get(te.AbsURL("/hello?name=World"))
	// Note that te.AbsURL() turns "/hello" into "http://127.0.0.1:[PORT]/hello"
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "Hello World", string(body))
}
```


Referencing the example above, here's a few things that are going on.

- The `te` object (of type `techo.Techo`) returned by`techo.New()` has an embedded `Echo` server.
    - So you can treat `techo.Techo` just like `echo.Echo`, and do things like `te.POST( ... )` etc.
- When `techo.New()` is called, the server is started automatically...
    - on its own goroutine (transparently to the caller)
    - on a random available port
- The base URL of the server (e.g. `http://127.0.0.1:53012`) is accessible via `te.URL`, and the port at `te.Port`.
- There's a handy function for getting the URL of a path on the techo server. Just do `te.AbsURL("/my/path")`, and you'll get back something like `http://127.0.0.1:52713/my/path`.
- Stop the server using `te.Stop()`. A common idiom is `defer te.Stop()` immediately after the call to `techo.New()`. FYI, the stoppability is due a (hidden) [Graceful](https://github.com/tylerb/graceful) server.


## Examples

Start a server at a specific address:

```go
	cfg := &techo.Config{Addr: "localhost:6666")}
	te, err := techo.NewWith(cfg)
	if err != nil {
		return fmt.Errorf("Probably that port is in use: %v", err)
	}
	te.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})
```

To add multiple endpoints:

```go
	te.GET("/callme", func(c echo.Context) error {
		return c.String(http.StatusOK, "maybe")
	})

	te.GET("/goodbye", func(c echo.Context) error {
		return c.String(http.StatusOK, "goodbye cruel world")
	})
	
```

Start a TLS (HTTPS) server:

```go
	te := techo.NewTLS()
	defer te.Stop()
	te.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})
```

To set the default certs for new TLS servers:

```go
	cert, _ := ioutil.ReadFile("path/to/server.crt")
	key, _ := ioutil.ReadFile("path/to/server.key")
	techo.SetDefaultTLSCert(cert, key)
	te := techo.NewTLS()
```

Or for one-time use of your own certs:

```go
	cert, _ := ioutil.ReadFile("path/to/server.crt")
	key, _ := ioutil.ReadFile("path/to/server.key")
	cfg := &techo.Config{TLS: true, TLSCert: cert, TLSKey: key}
	te, err := techo.NewWith(cfg)
```


**Tip:** if your client uses  `http.DefaultClient` as its underlying client, and you are
using TLS, you will likely want to skip verification of the cert before any
requests to the `techo` endpoint. `techo` provides a convenience method to do so.

```go
	techo.SkipDefaultClientInsecureTLSVerify()
	// then just use your http client as per usual, no more "invalid cert" errors
	resp, err = http.Get(te.AbsURL("/hello"))
```

### Acknowledgements

These guys do all the magic behind the scenes.

* [Echo](https://github.com/labstack/echo)
* [Graceful](https://github.com/tylerb/graceful)
* [Testify](https://github.com/stretchr/testify)
