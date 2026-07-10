package entity

type UserState struct {
	IsReady    bool `json:"is_ready"`
	IsResigned bool `json:"is_resigned"`

	CurrentQuestionIndex    int  `json:"current_question_index"`
	CurrentQuestionAnswered bool `json:"current_question_answered"`
}

type Lobby struct {
	ID            string              `json:"id"`
	Participants  []int64             `json:"participants"`
	CreatedAtUnix int64               `json:"created_at"`
	State         string              `json:"state"`
	UserState     map[int64]UserState `json:"user_state"`
	Questions     []Question          `json:"questions"`
}

func (l Lobby) EntityID() ID {
	return NewID("lobby", l.ID)
}

type Question struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	Answers       []string `json:"answers"`
	CorrectAnswer int      `json:"correct_answer"`
}

func (q Question) EntityID() ID {
	return NewID("question", q.ID)
}
