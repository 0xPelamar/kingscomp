package webapp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/0xpelamar/kingscomp/pkg/jsonhelper"
	"github.com/labstack/echo/v4"
)

func (w *WebApp) urls() {
	lobby := w.e.Group("/lobby")
	lobby.GET("/:lobby_id", w.lobbyIndex)

	auth := w.e.Group("/auth")
	auth.POST("/validate", w.validateInitdata, w.authorize)
}

func (w *WebApp) authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		initData := c.Request().Header.Get("Authorization")
		isValid, err := ValidateWebAppInputData(initData, w.telegramBotToken)
		if err != nil {
			return err
		}
		if !isValid {
			return c.JSON(http.StatusUnauthorized, ResponseError(http.StatusUnauthorized, "invalid init data"))
		}
		parsed, _ := url.ParseQuery(initData)
		authTimestamp, _ := strconv.ParseInt(parsed.Get("auth_token"), 10, 64)
		authDate := time.Unix(authTimestamp, 0)
		account := jsonhelper.Decode[entity.Account]([]byte(parsed.Get("user")))

		account, err = w.App.Account.Account.Get(context.Background(), entity.NewID("account", account.ID))
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return c.JSON(http.StatusUnauthorized, ResponseError(http.StatusUnauthorized, "account not found"))
			}
			return err
		}
		c.Set("account", account)
		c.Set("auth_data", authDate)
		return next(c)
	}
}
