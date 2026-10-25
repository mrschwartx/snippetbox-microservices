package dto

import (
	"time"

	"snippetbox/backend-snippet/internal/model"
	"snippetbox/backend-snippet/internal/validator"
)

type CreateSnippetRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Expires int    `json:"expires"`
}

func (r CreateSnippetRequest) Validate() error {
	var v validator.Validator

	v.CheckField(validator.NotBlank(r.Title), "title", messageBlank)
	v.CheckField(validator.MaxChars(r.Title, model.TitleMaxLength), "title", messageTitleTooLong)
	v.CheckField(validator.NotBlank(r.Content), "content", messageBlank)
	v.CheckField(validator.MaxBytes(r.Content, model.ContentMaxBytes), "content", messageContentLong)
	v.CheckField(validator.PermittedValue(r.Expires, model.PermittedExpires...), "expires", messageExpires)

	return v.Err()
}

type SnippetResponse struct {
	ID      uint      `json:"id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	Created time.Time `json:"created"`
	Expires time.Time `json:"expires"`
}

func NewSnippetResponse(s *model.Snippet) SnippetResponse {
	return SnippetResponse{
		ID:      s.ID,
		Title:   s.Title,
		Content: s.Content,
		Created: s.Created.UTC(),
		Expires: s.Expires.UTC(),
	}
}

func NewSnippetListResponse(snippets []model.Snippet) []SnippetResponse {
	out := make([]SnippetResponse, 0, len(snippets))
	for i := range snippets {
		out = append(out, NewSnippetResponse(&snippets[i]))
	}
	return out
}
