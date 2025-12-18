package core

import "time"

type Note struct {
	ID        int64
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type NoteCreate struct {
	Title   string `json:"title" example:"Новая заметка"`
	Content string `json:"content" example:"Текст заметки"`
}

type NoteUpdate struct {
	Title   *string `json:"title,omitempty" example:"Обновлено"`
	Content *string `json:"content,omitempty" example:"Новый текст"`
}

type NoteCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

type NoteShort struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}
