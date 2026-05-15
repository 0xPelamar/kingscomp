package webapp

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type validateInitDataRequest struct {
	InitData string `json:"initData"`
}

func (w *WebApp) validateInitdata(c echo.Context) error {
	var req validateInitDataRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	isValid, err := ValidateWebAppInputData(req.InitData, w.telegramBotToken)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, ResponseOK(200, J{
		"is_valid": isValid,
	}))
}
