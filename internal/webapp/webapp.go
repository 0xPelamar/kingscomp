package webapp

import (
	"net"
	"net/http"

	"github.com/0xpelamar/kingscomp/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type WebApp struct {
	App  *service.App
	e    *echo.Echo
	addr string
}

func NewWebApp(app *service.App, addr string) *WebApp {
	e := echo.New()
	wa := &WebApp{
		App:  app,
		e:    e,
		addr: addr,
	}
	wa.urls()
	return wa
}

func (w *WebApp) Start() error {
	w.e.Use(middleware.Recover())
	return w.e.Start(w.addr)
}

func (w *WebApp) StartDev(listener net.Listener) error {
	w.e.Use(middleware.Recover())
	w.e.Use(middleware.RequestLogger())
	return http.Serve(listener, w.e)
}
