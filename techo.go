/*
Package techo is for running "real mocked" HTTP services in your test code, by
starting a stoppable HTTP (Echo) server in its own goroutine on a random port.
This is especially helpful with testing of REST clients.

Example:

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
*/
package techo

import (
	"fmt"
	"log"
	"net"

	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/tylerb/graceful"
)

type Techo struct {
	// Port is the port number the server is listening at.
	Port int
	// Base is the base URL (scheme + host + port), e.g. http://127.0.0.1:61241
	Base string
	// Addr provides access to the underlying TCP address object.
	Addr *net.TCPAddr
	*echo.Echo
	srv *graceful.Server
}

// New starts a server at any available port. This value is available in the Port field.
// In the unlikely event of an error, it is logged, and nil is returned.
func New() *Techo {
	te, err := listenAndStart("localhost:")
	if err != nil {
		log.Println(err)
	}
	return te
}

// NewAt starts a server at addr (e.g. "127.0.0.1:8080").
func NewAt(addr string) (*Techo, error) {
	return listenAndStart(addr)
}

func listenAndStart(addr string) (*Techo, error) {

	t := new(Techo)
	t.Echo = echo.New()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	t.Addr = l.Addr().(*net.TCPAddr)
	t.Port = t.Addr.Port
	t.Base = fmt.Sprintf("http://%v:%v", t.Addr.IP, t.Port)
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

// Stop instructs the server to shut down.
func (t *Techo) Stop() {
	t.srv.Stop(time.Millisecond * 1)
}

func (t *Techo) String() string {
	return t.Base
}

// URL constructs an absolute URL from the supplied (relative) path. For example,
// calling te.URL("/my/path") could return "http://127.0.0.1:53262/my/path".
func (t *Techo) URL(path string) string {

	if len(path) == 0 {
		return t.Base
	}

	if path[0] == '/' {
		return t.Base + path
	}

	return t.Base + "/" + path
}
