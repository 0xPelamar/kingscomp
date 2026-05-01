/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/0xpelamar/kingscomp/internal/repository/redis"
	"github.com/0xpelamar/kingscomp/internal/service"
	"github.com/0xpelamar/kingscomp/internal/telegram"
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
	accountService := service.NewAccountService(accountRepository)

	// setup app
	app := service.NewApp(accountService)

	tel, err := telegram.NewTelegram(app, os.Getenv("BOT_TOKEN"))
	if err != nil {
		logrus.WithError(err).Fatalln("could not create telegram bot")
	}
	logrus.Infoln("Connected to telegram successfully.")
	tel.Start()
}

func init() {
	rootCmd.AddCommand(serveCmd)

}
