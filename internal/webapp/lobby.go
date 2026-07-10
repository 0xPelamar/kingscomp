package webapp

import (
	"net/http"

	"github.com/0xpelamar/kingscomp/internal/webapp/views/pages"
	"github.com/labstack/echo/v4"
)

func (w *WebApp) lobbyIndex(c echo.Context) error {
	lobbyID := c.Param("lobby_id")
	lobby, players, err := w.App.LobbyPlayers(c.Request().Context(), lobbyID)
	if err != nil {
		return err
	}
	_, _ = lobby, players
	return HTML(c, pages.LobbyPage(c.Param("lobbyID")))

}

func (w *WebApp) lobbyReady(c echo.Context) error {
	account := getAccount(c)
	lobby := getLobby(c)

	if err := w.App.Lobby.UpdateUserState(c.Request().Context(), lobby.ID, account.ID, "is_ready", true); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, ResponseOK(http.StatusOK, "done"))
}
