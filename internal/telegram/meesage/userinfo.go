package meesage

import (
	"fmt"

	"github.com/0xpelamar/kingscomp/internal/entity"
)

func MainMenuText(account entity.Account) string {
	return fmt.Sprintf("🏰 Welcome «%s»\nWhat can I do for you your grace?", account.FirstName)
}
