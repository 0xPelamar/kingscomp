package webapp

import (
	"github.com/0xpelamar/kingscomp/internal/webapp/views"
	"github.com/labstack/echo/v4"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	lobbyID := c.Param("lobby_id")
	lobby, players, err := w.App.LobbyPlayers(c.Request().Context(), lobbyID)
	if err != nil {
		return err
	}
	_, _ = lobby, players
	return HTML(c, views.LobbyIndex())

}
