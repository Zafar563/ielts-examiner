package models

import "time"

// User represents a Telegram bot user
type User struct {
	TelegramID   int64     `json:"telegram_id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	FeedbackLang string    `json:"feedback_lang"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Submission represents an essay submitted for evaluation
type Submission struct {
	ID           int       `json:"id"`
	UserID       int64     `json:"user_id"`
	Topic        string    `json:"topic"`
	EssayText    string    `json:"essay_text"`
	OverallScore int       `json:"overall_score"`
	TRScore      int       `json:"tr_score"`
	CCScore      int       `json:"cc_score"`
	GAScore      int       `json:"ga_score"`
	LRScore      int       `json:"lr_score"`
	CEFRLevel    string    `json:"cefr_level"`
	Feedback     string    `json:"feedback"`
	Lang         string    `json:"lang"`
	CreatedAt    time.Time `json:"created_at"`
}

// AssessmentResult holds the parsed score and raw feedback from the LLM
type AssessmentResult struct {
	OverallScore int
	CEFRLevel    string
	TRScore      int
	CCScore      int
	GAScore      int
	LRScore      int
	Feedback     string
}
