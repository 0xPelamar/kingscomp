/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/0xpelamar/kingscomp/internal/matchmaking"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/0xpelamar/kingscomp/internal/repository/redis"
	"github.com/0xpelamar/kingscomp/internal/service"
	"github.com/0xpelamar/kingscomp/internal/telegram"
	"github.com/0xpelamar/kingscomp/internal/webapp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	// setup app
	app := service.NewApp(
		service.NewAccountService(accountRepository),
		service.NewLobbyService(lobbyRepository),
	)

	mm := matchmaking.NewRedisMatchMaking(redisClient, lobbyRepository)
	tel, err := telegram.NewTelegram(app, mm, os.Getenv("BOT_TOKEN"))
	if err != nil {
		logrus.WithError(err).Fatalln("could not create telegram bot")
	}
	logrus.Infoln("Connected to telegram successfully.")

	go tel.Start()

	wa := webapp.NewWebApp(app, "0.0.0.0:8080")

	if os.Getenv("env") == "local" {
		// TODO
	}

	wa.Start()

}

func init() {
	rootCmd.AddCommand(serveCmd)

}
