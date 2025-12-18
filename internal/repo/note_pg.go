package repo

import (
	"context"
	"database/sql"
	"time"

	"example.com/notes-api/internal/core"
)

// NoteRepoPG — PostgreSQL реализация репозитория заметок.
type NoteRepoPG struct {
	db *sql.DB
}

// NewNoteRepoPG создаёт новый экземпляр репозитория PostgreSQL.
func NewNoteRepoPG(db *sql.DB) *NoteRepoPG {
	return &NoteRepoPG{db: db}
}

// Create создаёт новую заметку и возвращает её ID.
func (r *NoteRepoPG) Create(ctx context.Context, n core.NoteCreate) (int64, error) {
	stmt, err := r.db.PrepareContext(ctx, `
		INSERT INTO notes (title, content)
		VALUES ($1, $2)
		RETURNING id
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int64
	if err := stmt.QueryRowContext(ctx, n.Title, n.Content).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// CreateWithLogTx демонстрирует транзакцию: создание заметки + лог в одной транзакции.
func (r *NoteRepoPG) CreateWithLogTx(ctx context.Context, n core.NoteCreate) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // откат если Commit не вызван

	// Вставка заметки
	var noteID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id`,
		n.Title, n.Content,
	).Scan(&noteID)
	if err != nil {
		return 0, err
	}

	// Вставка лог-действия
	_, err = tx.ExecContext(ctx,
		`INSERT INTO notes_log (note_id, action, created_at) VALUES ($1, $2, $3)`,
		noteID, "created", time.Now(),
	)
	if err != nil {
		return 0, err
	}

	// Коммит транзакции
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return noteID, nil
}

// GetByID возвращает заметку по ID.
func (r *NoteRepoPG) GetByID(ctx context.Context, id int64) (*core.Note, error) {
	stmt, err := r.db.PrepareContext(ctx, `
		SELECT id, title, content, created_at, updated_at
		FROM notes
		WHERE id = $1
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var note core.Note
	if err := stmt.QueryRowContext(ctx, id).Scan(
		&note.ID,
		&note.Title,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &note, nil
}

// Update обновляет заметку по ID.
func (r *NoteRepoPG) Update(ctx context.Context, id int64, u core.NoteUpdate) error {
	stmt, err := r.db.PrepareContext(ctx, `
		UPDATE notes
		SET title = COALESCE($1, title),
		    content = COALESCE($2, content),
		    updated_at = $3
		WHERE id = $4
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, u.Title, u.Content, time.Now(), id)
	return err
}

// Delete удаляет заметку по ID.
func (r *NoteRepoPG) Delete(ctx context.Context, id int64) error {
	stmt, err := r.db.PrepareContext(ctx, `
		DELETE FROM notes WHERE id = $1
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}

// ListFirstPage возвращает первые N заметок, отсортированных по дате создания.
func (r *NoteRepoPG) ListFirstPage(ctx context.Context, limit int) ([]core.Note, error) {
	stmt, err := r.db.PrepareContext(ctx, `
		SELECT id, title, content, created_at, updated_at
		FROM notes
		ORDER BY created_at DESC, id DESC
		LIMIT $1
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []core.Note
	for rows.Next() {
		var n core.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// ListAfterCursor возвращает заметки после указанного курсора (keyset-пагинация).
func (r *NoteRepoPG) ListAfterCursor(ctx context.Context, cursor core.NoteCursor, limit int) ([]core.Note, error) {
	stmt, err := r.db.PrepareContext(ctx, `
		SELECT id, title, content, created_at, updated_at
		FROM notes
		WHERE (created_at, id) < ($1, $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, cursor.CreatedAt, cursor.ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []core.Note
	for rows.Next() {
		var n core.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// GetByIDs возвращает короткую информацию по массиву ID заметок (батчинг).
func (r *NoteRepoPG) GetByIDs(ctx context.Context, ids []int64) ([]core.NoteShort, error) {
	if len(ids) == 0 {
		return []core.NoteShort{}, nil
	}

	stmt, err := r.db.PrepareContext(ctx, `
		SELECT id, title
		FROM notes
		WHERE id = ANY($1)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []core.NoteShort
	for rows.Next() {
		var n core.NoteShort
		if err := rows.Scan(&n.ID, &n.Title); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, nil
}

// GetAll возвращает все заметки, отсортированные по дате создания.
func (r *NoteRepoPG) GetAll(ctx context.Context) ([]core.Note, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, content, created_at, updated_at
		FROM notes
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []core.Note
	for rows.Next() {
		var n core.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}
