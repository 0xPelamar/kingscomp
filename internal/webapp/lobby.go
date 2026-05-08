package webapp

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	lobbyID := c.Param("lobby_id")
	lobby, players, err := w.App.LobbyPlayers(c.Request().Context(), lobbyID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"players": players,
		"lobby":   lobby,
	})
}
