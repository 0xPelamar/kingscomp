package webapp

import (
	"net/http"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/labstack/echo/v4"
)

func (w *WebApp) validateInitData(c echo.Context) error {
	acc := c.Get("account").(entity.Account)

	return c.JSON(http.StatusOK, ResponseOK(200, J{
		"is_valid": true,
		"account":  acc,
	}))
}
