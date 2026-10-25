package model

import "time"

const (
	TitleMaxLength      = 100
	ContentMaxBytes     = 65535
	LatestSnippetsLimit = 10
)

var PermittedExpires = []int{1, 7, 365}

type Snippet struct {
	ID      uint      `gorm:"primaryKey"`
	Title   string    `gorm:"size:100;not null"`
	Content string    `gorm:"type:text;not null"`
	Created time.Time `gorm:"not null;index:idx_snippets_created"`
	Expires time.Time `gorm:"not null;index:idx_snippets_expires"`
}

func (Snippet) TableName() string {
	return "snippets"
}

func NewSnippet(title, content string, expiresInDays int, now time.Time) Snippet {
	created := now.UTC()
	return Snippet{
		Title:   title,
		Content: content,
		Created: created,
		Expires: created.Add(time.Duration(expiresInDays) * 24 * time.Hour),
	}
}

func (s Snippet) IsExpired(now time.Time) bool {
	return !s.Expires.After(now)
}
