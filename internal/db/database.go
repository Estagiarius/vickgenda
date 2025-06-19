package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vickgenda/internal/models" // Assuming 'vickgenda' as module name from go.mod
	"vickgenda/internal/store"  // Import store package for InitTable functions

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var db *sql.DB

// InitDB initializes the database connection and creates tables if they don't exist.
func InitDB(dbPath string) error {
	if dbPath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get user config directory: %w", err)
		}

		dbDir := filepath.Join(configDir, "vickgenda")
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
			}
		}
		dbPath = filepath.Join(dbDir, "vickgenda.db")
	}

	fmt.Printf("Using database at: %s\n", dbPath)

	var sqlErr error
	db, sqlErr = sql.Open("sqlite3", dbPath)
	if sqlErr != nil {
		return fmt.Errorf("failed to open database at %s: %w", dbPath, sqlErr)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database at %s: %w", dbPath, err)
	}

	// Call createTables which now initializes all tables including Squad 3's
	if err := createTables(); err != nil {
	    return fmt.Errorf("failed during table creation: %w", err) // Consolidated error message
	}
	return nil
}

// GetDB returns the initialized database instance.
func GetDB() *sql.DB {
	if db == nil {
		// This case should ideally not be hit if InitDB is always called at startup.
		// Consider logging a warning or panicking if db is nil, as it indicates an improper app lifecycle.
		fmt.Println("Warning: GetDB() called before InitDB or InitDB failed. Attempting fallback initialization.")
		// Attempt a fallback initialization or return an error.
		// For now, let's try a default InitDB. This might not be ideal for all scenarios.
		if err := InitDB(""); err != nil { // Use default path
			fmt.Printf("Error in fallback InitDB: %v\n", err)
			return nil // Or panic, depending on desired strictness
		}
	}
	return db
}

// createTables creates the necessary tables in the database if they don't already exist.
func createTables() error {
	// Squad 5 - Questions Table
	questionsTableSQL := `
	CREATE TABLE IF NOT EXISTS questions (
		id TEXT PRIMARY KEY,
		subject TEXT NOT NULL,
		topic TEXT NOT NULL,
		difficulty TEXT NOT NULL,
		question_text TEXT NOT NULL,
		answer_options TEXT,
		correct_answers TEXT NOT NULL,
		question_type TEXT NOT NULL,
		source TEXT,
		tags TEXT,
		created_at TIMESTAMP NOT NULL,
		last_used_at TIMESTAMP,
		author TEXT
	);`

	_, err := db.Exec(questionsTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create questions table: %w", err)
	}

	// Squad 3 - Academic Tables
	if err := store.InitTermsTable(); err != nil {
		return fmt.Errorf("failed to initialize terms table: %w", err)
	}
	if err := store.InitStudentsTable(); err != nil {
		return fmt.Errorf("failed to initialize students table: %w", err)
	}

	if err := store.InitLessonsTable(); err != nil {
		return fmt.Errorf("failed to initialize lessons table: %w", err)
	}
	if err := store.InitGradesTable(); err != nil {
		return fmt.Errorf("failed to initialize grades table: %w", err)
	}

	// Add other squads' table initializations here as needed

	return nil
}

// --- CRUD Functions for Question Model (remain unchanged) ---

// CreateQuestion adds a new question to the database.
func CreateQuestion(q models.Question) (string, error) {
	if q.ID == "" { q.ID = uuid.NewString() }
	if q.CreatedAt.IsZero() { q.CreatedAt = time.Now() }
	answerOptionsJSON, err := json.Marshal(q.AnswerOptions)
	if err != nil { return "", fmt.Errorf("failed to marshal AnswerOptions: %w", err) }
	correctAnswersJSON, err := json.Marshal(q.CorrectAnswers)
	if err != nil { return "", fmt.Errorf("failed to marshal CorrectAnswers: %w", err) }
	tagsJSON, err := json.Marshal(q.Tags)
	if err != nil { return "", fmt.Errorf("failed to marshal Tags: %w", err) }
	stmt, err := db.Prepare(`
		INSERT INTO questions (
			id, subject, topic, difficulty, question_text,
			answer_options, correct_answers, question_type,
			source, tags, created_at, last_used_at, author
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil { return "", fmt.Errorf("failed to prepare insert statement for question: %w", err) }
	defer stmt.Close()
	var lastUsedAt sql.NullTime
	if !q.LastUsedAt.IsZero() { lastUsedAt = sql.NullTime{Time: q.LastUsedAt, Valid: true} }
	_, err = stmt.Exec(
		q.ID, q.Subject, q.Topic, q.Difficulty, q.QuestionText,
		string(answerOptionsJSON), string(correctAnswersJSON), q.QuestionType,
		q.Source, string(tagsJSON), q.CreatedAt, lastUsedAt, q.Author,
	)
	if err != nil { return "", fmt.Errorf("failed to execute insert statement for question: %w", err) }
	return q.ID, nil
}

// GetQuestion retrieves a question by its ID.
func GetQuestion(id string) (models.Question, error) {
	var q models.Question
	var answerOptionsJSON, correctAnswersJSON, tagsJSON sql.NullString
	var lastUsedAt sql.NullTime
	row := db.QueryRow(`
		SELECT id, subject, topic, difficulty, question_text,
		       answer_options, correct_answers, question_type,
		       source, tags, created_at, last_used_at, author
		FROM questions WHERE id = ?
	`, id)
	err := row.Scan(
		&q.ID, &q.Subject, &q.Topic, &q.Difficulty, &q.QuestionText,
		&answerOptionsJSON, &correctAnswersJSON, &q.QuestionType,
		&q.Source, &tagsJSON, &q.CreatedAt, &lastUsedAt, &q.Author,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return models.Question{}, fmt.Errorf("question with ID %s not found: %w", id, err) }
		return models.Question{}, fmt.Errorf("failed to scan question row: %w", err)
	}
	if answerOptionsJSON.Valid { if err := json.Unmarshal([]byte(answerOptionsJSON.String), &q.AnswerOptions); err != nil { return models.Question{}, fmt.Errorf("failed to unmarshal AnswerOptions: %w", err) } }
	if correctAnswersJSON.Valid { if err := json.Unmarshal([]byte(correctAnswersJSON.String), &q.CorrectAnswers); err != nil { return models.Question{}, fmt.Errorf("failed to unmarshal CorrectAnswers: %w", err) } }
	if tagsJSON.Valid { if err := json.Unmarshal([]byte(tagsJSON.String), &q.Tags); err != nil { return models.Question{}, fmt.Errorf("failed to unmarshal Tags: %w", err) } }
	if lastUsedAt.Valid { q.LastUsedAt = lastUsedAt.Time }
	return q, nil
}

// UpdateQuestion updates an existing question in the database.
func UpdateQuestion(q models.Question) error {
	if q.ID == "" { return errors.New("cannot update question without ID") }
	answerOptionsJSON, err := json.Marshal(q.AnswerOptions)
	if err != nil { return fmt.Errorf("failed to marshal AnswerOptions for update: %w", err) }
	correctAnswersJSON, err := json.Marshal(q.CorrectAnswers)
	if err != nil { return fmt.Errorf("failed to marshal CorrectAnswers for update: %w", err) }
	tagsJSON, err := json.Marshal(q.Tags)
	if err != nil { return fmt.Errorf("failed to marshal Tags for update: %w", err) }
	var lastUsedAt sql.NullTime
	if !q.LastUsedAt.IsZero() { lastUsedAt = sql.NullTime{Time: q.LastUsedAt, Valid: true} }
	stmt, err := db.Prepare(`
		UPDATE questions SET
			subject = ?, topic = ?, difficulty = ?, question_text = ?,
			answer_options = ?, correct_answers = ?, question_type = ?,
			source = ?, tags = ?, created_at = ?, last_used_at = ?, author = ?
		WHERE id = ?
	`)
	if err != nil { return fmt.Errorf("failed to prepare update statement for question: %w", err) }
	defer stmt.Close()
	res, err := stmt.Exec(
		q.Subject, q.Topic, q.Difficulty, q.QuestionText,
		string(answerOptionsJSON), string(correctAnswersJSON), q.QuestionType,
		q.Source, string(tagsJSON), q.CreatedAt, lastUsedAt, q.Author,
		q.ID,
	)
	if err != nil { return fmt.Errorf("failed to execute update statement for question ID %s: %w", q.ID, err) }
	rowsAffected, err := res.RowsAffected()
	if err != nil { return fmt.Errorf("failed to get rows affected for question ID %s: %w", q.ID, err) }
	if rowsAffected == 0 { return fmt.Errorf("no question found with ID %s to update", q.ID) }
	return nil
}

// DeleteQuestion removes a question from the database by its ID.
func DeleteQuestion(id string) error {
	if id == "" { return errors.New("cannot delete question without ID: ID cannot be empty") }
	if db == nil { return errors.New("database is not initialized") }
	result, err := db.Exec("DELETE FROM questions WHERE id = ?", id)
	if err != nil { return fmt.Errorf("erro ao executar a remoção da questão %s: %w", id, err) }
	rowsAffected, err := result.RowsAffected()
	if err != nil { return fmt.Errorf("erro ao verificar linhas afetadas para questão %s: %w", id, err) }
	if rowsAffected == 0 { return sql.ErrNoRows }
	return nil
}

// ListQuestions (remains unchanged, already long and complex)
func ListQuestions(filters map[string]interface{}, sortBy string, order string, limit int, page int) ([]models.Question, int, error) {
	// ... existing ListQuestions implementation ...
	var questions []models.Question
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("SELECT id, subject, topic, difficulty, question_text, answer_options, correct_answers, question_type, source, tags, created_at, last_used_at, author FROM questions")
	countQueryBuilder := strings.Builder{}
	countQueryBuilder.WriteString("SELECT COUNT(*) FROM questions")
	whereClauses := []string{}
	var searchConditions []string
	standardQueryArgs := []interface{}
	searchQueryArgs := []interface{}
	searchQuery, hasSearchQuery := filters["search_query"].(string)
	searchFields, hasSearchFields := filters["search_fields"].([]string)
	for key, value := range filters {
		if key == "search_query" || key == "search_fields" { continue }
		if valStr, ok := value.(string); ok && valStr != "" {
			switch key {
			case "subject", "topic", "difficulty", "question_type", "author":
				whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", key))
				standardQueryArgs = append(standardQueryArgs, valStr)
			case "tags":
				whereClauses = append(whereClauses, "tags LIKE ?")
				standardQueryArgs = append(standardQueryArgs, "%"+valStr+"%")
			}
		}
	}
	if hasSearchQuery && searchQuery != "" && hasSearchFields && len(searchFields) > 0 {
		likeQuery := "%" + searchQuery + "%"
		for _, field := range searchFields {
			validSearchFields := map[string]bool{"id": true, "subject": true, "topic": true, "question_text": true, "source": true, "tags": true, "author": true, "difficulty": true, "question_type": true}
			if !validSearchFields[strings.ToLower(field)] { return nil, 0, fmt.Errorf("invalid search_field provided: %s", field) }
			searchConditions = append(searchConditions, fmt.Sprintf("%s LIKE ?", field))
			searchQueryArgs = append(searchQueryArgs, likeQuery)
		}
	}
	if len(searchConditions) > 0 { whereClauses = append(whereClauses, "("+strings.Join(searchConditions, " OR ")+")") }
	args := append(standardQueryArgs, searchQueryArgs...)
	if len(whereClauses) > 0 {
		whereString := " WHERE " + strings.Join(whereClauses, " AND ")
		queryBuilder.WriteString(whereString)
		countQueryBuilder.WriteString(whereString)
	}
	if sortBy != "" {
		validSortBy := map[string]bool{"id": true, "subject": true, "topic": true, "difficulty": true, "question_type": true, "created_at": true, "last_used_at": true, "author": true}
		if !validSortBy[strings.ToLower(sortBy)] { return nil, 0, fmt.Errorf("invalid sort_by column: %s", sortBy) }
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s", sortBy))
		if strings.ToUpper(order) == "DESC" { queryBuilder.WriteString(" DESC") } else { queryBuilder.WriteString(" ASC") }
	} else { queryBuilder.WriteString(" ORDER BY created_at DESC") }
	if limit <= 0 { limit = 20 }
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset))
	var totalCount int
	err := db.QueryRow(countQueryBuilder.String(), args...).Scan(&totalCount)
	if err != nil { return nil, 0, fmt.Errorf("failed to count questions: %w", err) }
	if totalCount == 0 { return []models.Question{}, 0, nil }
	rows, err := db.Query(queryBuilder.String(), args...)
	if err != nil { return nil, 0, fmt.Errorf("failed to list questions: %w", err) }
	defer rows.Close()
	for rows.Next() {
		var q models.Question
		var answerOptionsJSON, correctAnswersJSON, tagsJSON sql.NullString
		var lastUsedAt sql.NullTime
		if err := rows.Scan(&q.ID, &q.Subject, &q.Topic, &q.Difficulty, &q.QuestionText, &answerOptionsJSON, &correctAnswersJSON, &q.QuestionType, &q.Source, &tagsJSON, &q.CreatedAt, &lastUsedAt, &q.Author); err != nil {
			return nil, 0, fmt.Errorf("failed to scan question during list: %w", err)
		}
		if answerOptionsJSON.Valid { if err := json.Unmarshal([]byte(answerOptionsJSON.String), &q.AnswerOptions); err != nil { /* Log */ } }
		if correctAnswersJSON.Valid { if err := json.Unmarshal([]byte(correctAnswersJSON.String), &q.CorrectAnswers); err != nil { /* Log */ } }
		if tagsJSON.Valid { if err := json.Unmarshal([]byte(tagsJSON.String), &q.Tags); err != nil { /* Log */ } }
		if lastUsedAt.Valid { q.LastUsedAt = lastUsedAt.Time }
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil { return nil, 0, fmt.Errorf("error iterating question rows: %w", err) }
	return questions, totalCount, nil
}
package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vickgenda/internal/models" // Assuming 'vickgenda' as module name from go.mod
	"vickgenda/internal/store"  // Import store package for InitTable functions

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var db *sql.DB

// InitDB initializes the database connection and creates tables if they don't exist.
func InitDB(dbPath string) error {
	if dbPath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get user config directory: %w", err)
		}

		dbDir := filepath.Join(configDir, "vickgenda")
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
			}
		}
		dbPath = filepath.Join(dbDir, "vickgenda.db")
	}

	fmt.Printf("Using database at: %s\n", dbPath)

	var sqlErr error
	db, sqlErr = sql.Open("sqlite3", dbPath)
	if sqlErr != nil {
		return fmt.Errorf("failed to open database at %s: %w", dbPath, sqlErr)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database at %s: %w", dbPath, err)
	}

	// Call createTables which now initializes all tables including Squad 3's
	if err := createTables(); err != nil {
	    return fmt.Errorf("failed during table creation: %w", err) // Consolidated error message
	}
	return nil
}

// GetDB returns the initialized database instance.
func GetDB() *sql.DB {
	if db == nil {
		// This case should ideally not be hit if InitDB is always called at startup.
		// Consider logging a warning or panicking if db is nil, as it indicates an improper app lifecycle.
		fmt.Println("Warning: GetDB() called before InitDB or InitDB failed. Attempting fallback initialization.")
		// Attempt a fallback initialization or return an error.
		// For now, let's try a default InitDB. This might not be ideal for all scenarios.
		if err := InitDB(""); err != nil { // Use default path
			fmt.Printf("Error in fallback InitDB: %v\n", err)
			return nil // Or panic, depending on desired strictness
		}
	}
	return db
}

// createTables creates the necessary tables in the database if they don't already exist.
func createTables() error {
	// Squad 5 - Questions Table
	questionsTableSQL := `
	CREATE TABLE IF NOT EXISTS questions (
		id TEXT PRIMARY KEY,
		subject TEXT NOT NULL,
		topic TEXT NOT NULL,
		difficulty TEXT NOT NULL,
		question_text TEXT NOT NULL,
		answer_options TEXT,
		correct_answers TEXT NOT NULL,
		question_type TEXT NOT NULL,
		source TEXT,
		tags TEXT,
		created_at TIMESTAMP NOT NULL,
		last_used_at TIMESTAMP,
		author TEXT
	);`

	_, err := db.Exec(questionsTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create questions table: %w", err)
	}

	// Squad 3 - Academic Tables
	if err := store.InitTermsTable(); err != nil {
		return fmt.Errorf("failed to initialize terms table: %w", err)
	}
	if err := store.InitStudentsTable(); err != nil {
		return fmt.Errorf("failed to initialize students table: %w", err)
	}

	if err := store.InitLessonsTable(); err != nil {
		return fmt.Errorf("failed to initialize lessons table: %w", err)
	}
	if err := store.InitGradesTable(); err != nil {
		return fmt.Errorf("failed to initialize grades table: %w", err)
	}

	// Add other squads' table initializations here as needed

	return nil
}

// --- CRUD Functions for Question Model (remain unchanged) ---

// CreateQuestion adds a new question to the database.
func CreateQuestion(q models.Question) (string, error) {
	if q.ID == "" { q.ID = uuid.NewString() }
	if q.CreatedAt.IsZero() { q.CreatedAt = time.Now() }
	answerOptionsJSON, err := json.Marshal(q.AnswerOptions)
	if err != nil { return "", fmt.Errorf("failed to marshal AnswerOptions: %w", err) }
	correctAnswersJSON, err := json.Marshal(q.CorrectAnswers)
	if err != nil { return "", fmt.Errorf("failed to marshal CorrectAnswers: %w", err) }
	tagsJSON, err := json.Marshal(q.Tags)
	if err != nil { return "", fmt.Errorf("failed to marshal Tags: %w", err) }
	stmt, err := db.Prepare(`
		INSERT INTO questions (
			id, subject, topic, difficulty, question_text,
			answer_options, correct_answers, question_type,
			source, tags, created_at, last_used_at, author
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil { return "", fmt.Errorf("failed to prepare insert statement for question: %w", err) }
	defer stmt.Close()
	var lastUsedAt sql.NullTime
	if !q.LastUsedAt.IsZero() { lastUsedAt = sql.NullTime{Time: q.LastUsedAt, Valid: true} }
	_, err = stmt.Exec(
		q.ID, q.Subject, q.Topic, q.Difficulty, q.QuestionText,
		string(answerOptionsJSON), string(correctAnswersJSON), q.QuestionType,
		q.Source, string(tagsJSON), q.CreatedAt, lastUsedAt, q.Author,
	)
	if err != nil { return "", fmt.Errorf("failed to execute insert statement for question: %w", err) }
	return q.ID, nil
}

// GetQuestion retrieves a question by its ID.
func GetQuestion(id string) (models.Question, error) {
	var q models.Question
	var answerOptionsJSON, correctAnswersJSON, tagsJSON sql.NullString
	var lastUsedAt sql.NullTime
	row := db.QueryRow(`
		SELECT id, subject, topic, difficulty, question_text,
		       answer_options, correct_answers, question_type,
		       source, tags, created_at, last_used_at, author
		FROM questions WHERE id = ?
	`, id)
	err := row.Scan(
		&q.ID, &q.Subject, &q.Topic, &q.Difficulty, &q.QuestionText,
		&answerOptionsJSON, &correctAnswersJSON, &q.QuestionType,
		&q.Source, &tagsJSON, &q.CreatedAt, &lastUsedAt, &q.Author,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return models.Question{}, fmt.Errorf("question with ID %s not found: %w", id, err) }
		return models.Question{}, fmt.Errorf("failed to scan question row: %w", err)
	}
	if answerOptionsJSON.Valid { if err := json.Unmarshal([]byte(answerOptionsJSON.String), &q.AnswerOptions); err != nil { return models.Question{}, fmt.Errorf("failed to unmarshal AnswerOptions: %w", err) } }
	if correctAnswersJSON.Valid { if err := json.Unmarshal([]byte(correctAnswersJSON.String), &q.CorrectAnswers); err != nil { return models.Question{}, fmt.Errorf("failed to unmarshal CorrectAnswers: %w", err) } }
	if tagsJSON.Valid { if err := json.Unmarshal([]byte(tagsJSON.String), &q.Tags); err != nil { return models.Question{}, fmt.Errorf("failed to unmarshal Tags: %w", err) } }
	if lastUsedAt.Valid { q.LastUsedAt = lastUsedAt.Time }
	return q, nil
}

// UpdateQuestion updates an existing question in the database.
func UpdateQuestion(q models.Question) error {
	if q.ID == "" { return errors.New("cannot update question without ID") }
	answerOptionsJSON, err := json.Marshal(q.AnswerOptions)
	if err != nil { return fmt.Errorf("failed to marshal AnswerOptions for update: %w", err) }
	correctAnswersJSON, err := json.Marshal(q.CorrectAnswers)
	if err != nil { return fmt.Errorf("failed to marshal CorrectAnswers for update: %w", err) }
	tagsJSON, err := json.Marshal(q.Tags)
	if err != nil { return fmt.Errorf("failed to marshal Tags for update: %w", err) }
	var lastUsedAt sql.NullTime
	if !q.LastUsedAt.IsZero() { lastUsedAt = sql.NullTime{Time: q.LastUsedAt, Valid: true} }
	stmt, err := db.Prepare(`
		UPDATE questions SET
			subject = ?, topic = ?, difficulty = ?, question_text = ?,
			answer_options = ?, correct_answers = ?, question_type = ?,
			source = ?, tags = ?, created_at = ?, last_used_at = ?, author = ?
		WHERE id = ?
	`)
	if err != nil { return fmt.Errorf("failed to prepare update statement for question: %w", err) }
	defer stmt.Close()
	res, err := stmt.Exec(
		q.Subject, q.Topic, q.Difficulty, q.QuestionText,
		string(answerOptionsJSON), string(correctAnswersJSON), q.QuestionType,
		q.Source, string(tagsJSON), q.CreatedAt, lastUsedAt, q.Author,
		q.ID,
	)
	if err != nil { return fmt.Errorf("failed to execute update statement for question ID %s: %w", q.ID, err) }
	rowsAffected, err := res.RowsAffected()
	if err != nil { return fmt.Errorf("failed to get rows affected for question ID %s: %w", q.ID, err) }
	if rowsAffected == 0 { return fmt.Errorf("no question found with ID %s to update", q.ID) }
	return nil
}

// DeleteQuestion removes a question from the database by its ID.
func DeleteQuestion(id string) error {
	if id == "" { return errors.New("cannot delete question without ID: ID cannot be empty") }
	if db == nil { return errors.New("database is not initialized") }
	result, err := db.Exec("DELETE FROM questions WHERE id = ?", id)
	if err != nil { return fmt.Errorf("erro ao executar a remoção da questão %s: %w", id, err) }
	rowsAffected, err := result.RowsAffected()
	if err != nil { return fmt.Errorf("erro ao verificar linhas afetadas para questão %s: %w", id, err) }
	if rowsAffected == 0 { return sql.ErrNoRows }
	return nil
}

// ListQuestions (remains unchanged, already long and complex)
func ListQuestions(filters map[string]interface{}, sortBy string, order string, limit int, page int) ([]models.Question, int, error) {
	// ... existing ListQuestions implementation ...
	var questions []models.Question
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("SELECT id, subject, topic, difficulty, question_text, answer_options, correct_answers, question_type, source, tags, created_at, last_used_at, author FROM questions")
	countQueryBuilder := strings.Builder{}
	countQueryBuilder.WriteString("SELECT COUNT(*) FROM questions")
	whereClauses := []string{}
	var searchConditions []string
	standardQueryArgs := []interface{}
	searchQueryArgs := []interface{}
	searchQuery, hasSearchQuery := filters["search_query"].(string)
	searchFields, hasSearchFields := filters["search_fields"].([]string)
	for key, value := range filters {
		if key == "search_query" || key == "search_fields" { continue }
		if valStr, ok := value.(string); ok && valStr != "" {
			switch key {
			case "subject", "topic", "difficulty", "question_type", "author":
				whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", key))
				standardQueryArgs = append(standardQueryArgs, valStr)
			case "tags":
				whereClauses = append(whereClauses, "tags LIKE ?")
				standardQueryArgs = append(standardQueryArgs, "%"+valStr+"%")
			}
		}
	}
	if hasSearchQuery && searchQuery != "" && hasSearchFields && len(searchFields) > 0 {
		likeQuery := "%" + searchQuery + "%"
		for _, field := range searchFields {
			validSearchFields := map[string]bool{"id": true, "subject": true, "topic": true, "question_text": true, "source": true, "tags": true, "author": true, "difficulty": true, "question_type": true}
			if !validSearchFields[strings.ToLower(field)] { return nil, 0, fmt.Errorf("invalid search_field provided: %s", field) }
			searchConditions = append(searchConditions, fmt.Sprintf("%s LIKE ?", field))
			searchQueryArgs = append(searchQueryArgs, likeQuery)
		}
	}
	if len(searchConditions) > 0 { whereClauses = append(whereClauses, "("+strings.Join(searchConditions, " OR ")+")") }
	args := append(standardQueryArgs, searchQueryArgs...)
	if len(whereClauses) > 0 {
		whereString := " WHERE " + strings.Join(whereClauses, " AND ")
		queryBuilder.WriteString(whereString)
		countQueryBuilder.WriteString(whereString)
	}
	if sortBy != "" {
		validSortBy := map[string]bool{"id": true, "subject": true, "topic": true, "difficulty": true, "question_type": true, "created_at": true, "last_used_at": true, "author": true}
		if !validSortBy[strings.ToLower(sortBy)] { return nil, 0, fmt.Errorf("invalid sort_by column: %s", sortBy) }
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s", sortBy))
		if strings.ToUpper(order) == "DESC" { queryBuilder.WriteString(" DESC") } else { queryBuilder.WriteString(" ASC") }
	} else { queryBuilder.WriteString(" ORDER BY created_at DESC") }
	if limit <= 0 { limit = 20 }
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset))
	var totalCount int
	err := db.QueryRow(countQueryBuilder.String(), args...).Scan(&totalCount)
	if err != nil { return nil, 0, fmt.Errorf("failed to count questions: %w", err) }
	if totalCount == 0 { return []models.Question{}, 0, nil }
	rows, err := db.Query(queryBuilder.String(), args...)
	if err != nil { return nil, 0, fmt.Errorf("failed to list questions: %w", err) }
	defer rows.Close()
	for rows.Next() {
		var q models.Question
		var answerOptionsJSON, correctAnswersJSON, tagsJSON sql.NullString
		var lastUsedAt sql.NullTime
		if err := rows.Scan(&q.ID, &q.Subject, &q.Topic, &q.Difficulty, &q.QuestionText, &answerOptionsJSON, &correctAnswersJSON, &q.QuestionType, &q.Source, &tagsJSON, &q.CreatedAt, &lastUsedAt, &q.Author); err != nil {
			return nil, 0, fmt.Errorf("failed to scan question during list: %w", err)
		}
		if answerOptionsJSON.Valid { if err := json.Unmarshal([]byte(answerOptionsJSON.String), &q.AnswerOptions); err != nil { /* Log */ } }
		if correctAnswersJSON.Valid { if err := json.Unmarshal([]byte(correctAnswersJSON.String), &q.CorrectAnswers); err != nil { /* Log */ } }
		if tagsJSON.Valid { if err := json.Unmarshal([]byte(tagsJSON.String), &q.Tags); err != nil { /* Log */ } }
		if lastUsedAt.Valid { q.LastUsedAt = lastUsedAt.Time }
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil { return nil, 0, fmt.Errorf("error iterating question rows: %w", err) }
	return questions, totalCount, nil
}

```
