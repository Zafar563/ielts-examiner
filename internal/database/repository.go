package database

import (
	"context"
	"database/sql"
	"fmt"
	"ielts_bot/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// UpsertUser saves or updates a user in the database
func (r *Repository) UpsertUser(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (telegram_id, username, first_name, last_name, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	ON CONFLICT (telegram_id) DO UPDATE 
	SET username = EXCLUDED.username,
	    first_name = EXCLUDED.first_name,
	    last_name = EXCLUDED.last_name,
	    updated_at = CURRENT_TIMESTAMP;
	`
	_, err := r.db.ExecContext(ctx, query, user.TelegramID, user.Username, user.FirstName, user.LastName)
	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}
	return nil
}

// SaveSubmission records an essay assessment result
func (r *Repository) SaveSubmission(ctx context.Context, sub *models.Submission) error {
	query := `
	INSERT INTO submissions (
		user_id, topic, essay_text, overall_score, 
		tr_score, cc_score, ga_score, lr_score, cefr_level, feedback
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id, created_at;
	`
	err := r.db.QueryRowContext(ctx, query,
		sub.UserID, sub.Topic, sub.EssayText, sub.OverallScore,
		sub.TRScore, sub.CCScore, sub.GAScore, sub.LRScore, sub.CEFRLevel, sub.Feedback,
	).Scan(&sub.ID, &sub.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save submission: %w", err)
	}
	return nil
}

// GetUserSubmissions returns recent submissions for a user
func (r *Repository) GetUserSubmissions(ctx context.Context, userID int64, limit int) ([]models.Submission, error) {
	query := `
	SELECT id, user_id, topic, essay_text, overall_score, 
	       tr_score, cc_score, ga_score, lr_score, cefr_level, feedback, created_at
	FROM submissions
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2;
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query user submissions: %w", err)
	}
	defer rows.Close()

	var list []models.Submission
	for rows.Next() {
		var s models.Submission
		var topic sql.NullString
		if err := rows.Scan(
			&s.ID, &s.UserID, &topic, &s.EssayText, &s.OverallScore,
			&s.TRScore, &s.CCScore, &s.GAScore, &s.LRScore, &s.CEFRLevel, &s.Feedback, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}
		if topic.Valid {
			s.Topic = topic.String
		}
		list = append(list, s)
	}

	return list, nil
}

// GetUserStats returns total count and average score for a user
func (r *Repository) GetUserStats(ctx context.Context, userID int64) (int, float64, error) {
	query := `
	SELECT COUNT(*), COALESCE(AVG(overall_score), 0)
	FROM submissions
	WHERE user_id = $1;
	`
	var count int
	var avgScore float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count, &avgScore)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user stats: %w", err)
	}
	return count, avgScore, nil
}
