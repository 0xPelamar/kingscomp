package webapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func HTML(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func ValidateWebAppInputData(inputData string, telegramBotToken string) (bool, error) {
	initData, err := url.ParseQuery(inputData)
	if err != nil {
		logrus.WithError(err).Error("failed to parse input data")
		return false, err
	}

	dataCheckString := make([]string, 0, len(initData))
	for k, v := range initData {
		if k == "hash" {
			continue
		}
		dataCheckString = append(dataCheckString, fmt.Sprintf("%s=%s", k, v[0]))
	}
	sort.Strings(dataCheckString)

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(telegramBotToken))

	hHash := hmac.New(sha256.New, secret.Sum(nil))
	hHash.Write([]byte(strings.Join(dataCheckString, "\n")))

	hash := hex.EncodeToString(hHash.Sum(nil))

	if initData.Get("hash") != hash {
		return false, nil
	}
	return true, nil
}

type J map[string]any

func ResponseOK(code int, data any) any {
	return J{
		"ok":   true,
		"code": code,
		"data": data,
	}
}

func ResponseError(code int, data any) any {
	return J{
		"ok":   false,
		"code": code,
		"data": data,
	}
}
