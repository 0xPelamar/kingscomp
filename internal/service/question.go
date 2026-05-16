package service

import "github.com/0xpelamar/kingscomp/internal/repository"

type QuestionService struct {
	repository.Question
}

func NewQuestionService(rep repository.Question) *QuestionService {
	return &QuestionService{
		Question: rep,
	}
}
