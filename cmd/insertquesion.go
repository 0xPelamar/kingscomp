package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/repository"
	"github.com/0xpelamar/kingscomp/internal/repository/redis"
	"github.com/0xpelamar/kingscomp/pkg/jsonhelper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// insertQuestionCmd represents the serve command
var insertQuestionCmd = &cobra.Command{
	Use:   "insertquestion",
	Short: "insert question command",

	Run: insertQuestion,
}

func insertQuestion(cmd *cobra.Command, args []string) {
	filePath, err := cmd.Flags().GetString("file-path")

	if err != nil || filePath == "" {
		logrus.WithError(err).Fatalln("could not parse file path. use --file-path")
	}
	fmt.Println("the file path is :", filePath)
	b, err := os.ReadFile(filePath)
	if err != nil {
		logrus.WithError(err).Fatalln("could not read the question file.")
	}
	questions := jsonhelper.Decode[[]entity.Question](b)

	// setup repositories
	redisClient, err := redis.NewRedisClient(os.Getenv("REDIS_URL"))
	if err != nil {
		logrus.WithError(err).Fatalln("could not connect to redis")
	}
	logrus.Infoln("Connected to redis successfully.")
	questionRepository := repository.NewQuestionRedisRepository(redisClient)

	logrus.WithField("num", len(questions)).Infoln("inserting new questions...")
	err = questionRepository.PushActiveQuestion(context.Background(), questions...)
	if err != nil {
		logrus.WithError(err).Fatalln("could not insert question.")
	}

	logrus.Infoln("Successfully inserted new questions.")
}

func init() {
	rootCmd.AddCommand(insertQuestionCmd)

	insertQuestionCmd.PersistentFlags().String("file-path", "", "path to the json questions file")
}
