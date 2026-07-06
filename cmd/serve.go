package cmd

import (
	"context"
	"net/url"
	"os"

	"github.com/0xpelamar/kingscomp/internal/config"
	"github.com/0xpelamar/kingscomp/internal/matchmaking"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/0xpelamar/kingscomp/internal/repository/redis"
	"github.com/0xpelamar/kingscomp/internal/service"
	"github.com/0xpelamar/kingscomp/internal/telegram"
	"github.com/0xpelamar/kingscomp/internal/webapp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.ngrok.com/ngrok"
	nconfig "golang.ngrok.com/ngrok/config"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve command",

	Run: serve,
}

func serve(cmd *cobra.Command, args []string) {
	// setup repositories
	redisClient, err := redis.NewRedisClient(os.Getenv("REDIS_URL"))
	if err != nil {
		logrus.WithError(err).Fatalln("could not connect to redis")
	}
	logrus.Infoln("Connected to redis successfully.")
	accountRepository := repository.NewAccountRedisRepository(redisClient)
	lobbyRepository := repository.NewLobbyRedisRepository(redisClient)
	questionRepository := repository.NewQuestionRedisRepository(redisClient)

	// setup app
	app := service.NewApp(
		service.NewAccountService(accountRepository),
		service.NewLobbyService(lobbyRepository),
	)

	mm := matchmaking.NewRedisMatchMaking(redisClient, lobbyRepository, questionRepository)
	tel, err := telegram.NewTelegram(app, mm, os.Getenv("BOT_TOKEN"))
	if err != nil {
		logrus.WithError(err).Fatalln("could not create telegram bot")
	}
	logrus.Infoln("Connected to telegram successfully.")

	go tel.Start()

	wa := webapp.NewWebApp(app, "0.0.0.0:8080", os.Getenv("BOT_TOKEN"))

	// Use ngrok if it's local
	if os.Getenv("ENV") == "local" {
		proxyURL, err := url.Parse("socks5://127.0.0.1:10808")
		if err != nil {
			logrus.WithError(err).Fatalln("could not parse proxy url")
		}
		listener, err := ngrok.Listen(
			context.Background(),
			nconfig.HTTPEndpoint(),
			ngrok.WithAuthtokenFromEnv(),
			ngrok.WithProxyURL(proxyURL),
		)
		logrus.Infoln("Running the web app in local mode")

		if err != nil {
			logrus.WithError(err).Fatalln("could not forward ngrok")
		}
		config.Default.WebAppAddr = "https://" + listener.Addr().String()
		logrus.Infof("Web App is available at: %s", config.Default.WebAppAddr)
		if err := wa.StartDev(listener); err != nil {
			logrus.WithError(err).Fatalln("could not start webapp with ngrok")
		}

	} else {
		if err := wa.Start(); err != nil {
			logrus.WithError(err).Fatalln("could not start webapp")
		}
	}

}

func init() {
	rootCmd.AddCommand(serveCmd)

}
