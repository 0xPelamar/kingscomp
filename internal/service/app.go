package service

type App struct {
	Account *AccountService
}

func NewApp(account *AccountService) *App {
	return &App{
		Account: account,
	}
}
