/*
Package techo is for transparently "mocking" HTTP services in your
test code, by starting a real (Echo) server in its own goroutine, than can
be stopped easily.

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

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/tylerb/graceful"
)

type Techo struct {
	Port int
	// Base is the base URL (scheme + host + port), e.g. http://127.0.0.1:61241
	Base string
	Addr *net.TCPAddr
	*echo.Echo
	srv *graceful.Server
}

// New starts a server at any available port. This value is available in the Port field.
func New() *Techo {
	return listenAndStart("localhost:")
}

// NewAt starts a server at addr (e.g. "127.0.0.1:8080").
func NewAt(addr string) *Techo {
	return listenAndStart(addr)
}

func listenAndStart(addr string) *Techo {

	t := new(Techo)
	t.Echo = echo.New()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Logger().Error(err)
		return nil
	}

	t.Addr = l.Addr().(*net.TCPAddr)
	t.Port = t.Addr.Port
	t.Base = fmt.Sprintf("http://%v:%v", t.Addr.IP, t.Port)
	std := standard.New(fmt.Sprintf(":%v", t.Addr.Port))
	std.SetHandler(t.Echo)
	t.srv = &graceful.Server{
		Server: std.Server,
	}

	go func() {
		err := t.srv.Serve(l)
		if err != nil {
			log.Printf("techo error: %v\n", err)
		}
		log.Printf("techo exiting [%v]\n", t)
	}()

	//log.Printf("techo listening at %v", t.URL)
	return t
}

// Stop instructs the server to shutdown.
func (t *Techo) Stop() {

	//log.Printf("t.Stop(%v)\n", t)
	t.srv.Stop(0)

}

func (t *Techo) String() string {
	return t.Base
}

// URL constructs an absolute URL from the supplied (relative) path.
func (t *Techo) URL(path string) string {

	if len(path) == 0 {
		return t.Base
	}

	if path[0] == '/' {
		return t.Base + path
	}

	return t.Base + "/" + path
}
