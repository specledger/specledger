package comment

const (
	ContentPreviewLen      = 120
	SelectedTextPreviewLen = 80
)

type ReviewComment struct {
	ID              string  `json:"id"`
	ChangeID        string  `json:"change_id"`
	FilePath        string  `json:"file_path"`
	Content         string  `json:"content"`
	SelectedText    *string `json:"selected_text"`
	Line            *int    `json:"line"`
	StartLine       *int    `json:"start_line"`
	IsResolved      bool    `json:"is_resolved"`
	AuthorID        string  `json:"author_id"`
	AuthorName      *string `json:"author_name"`
	AuthorEmail     *string `json:"author_email"`
	ParentCommentID *string `json:"parent_comment_id"`
	IsAIGenerated   *bool   `json:"is_ai_generated"`
	TriggeredByUser *string `json:"triggered_by_user_id"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type CommentThread struct {
	Parent  ReviewComment   `json:"parent"`
	Replies []ReviewComment `json:"replies"`
}

type CommentSummary struct {
	ID                  string `json:"id"`
	FilePath            string `json:"file_path"`
	ContentPreview      string `json:"content_preview"`
	SelectedTextPreview string `json:"selected_text_preview"`
	AuthorName          string `json:"author_name"`
	ReplyCount          int    `json:"reply_count"`
	IsResolved          bool   `json:"is_resolved"`
	CreatedAt           string `json:"created_at"`
}

type CommentListOutput struct {
	SpecKey    string           `json:"spec_key"`
	ChangeID   string           `json:"change_id"`
	Comments   []CommentSummary `json:"comments"`
	TotalCount int              `json:"total_count"`
	Hint       string           `json:"hint"`
}

type CommentDetail struct {
	ID            string        `json:"id"`
	FilePath      string        `json:"file_path"`
	Content       string        `json:"content"`
	SelectedText  *string       `json:"selected_text"`
	Line          *int          `json:"line"`
	StartLine     *int          `json:"start_line"`
	IsResolved    bool          `json:"is_resolved"`
	AuthorName    *string       `json:"author_name"`
	AuthorEmail   *string       `json:"author_email"`
	IsAIGenerated *bool         `json:"is_ai_generated"`
	CreatedAt     string        `json:"created_at"`
	Replies       []ReplyDetail `json:"replies"`
}

type ReplyDetail struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	AuthorName string `json:"author_name"`
	CreatedAt  string `json:"created_at"`
}

type CommentShowOutput struct {
	Comments []CommentDetail `json:"comments"`
}

type CommentReplyInput struct {
	ParentCommentID string `json:"parent_comment_id"`
	Content         string `json:"content"`
	AuthorName      string `json:"author_name"`
	AuthorEmail     string `json:"author_email"`
}

type CommentResolveResult struct {
	ResolvedID   string `json:"resolved"`
	CascadeCount int    `json:"cascade_count"`
}

type CommentReplyResult struct {
	ID              string `json:"id"`
	ParentCommentID string `json:"parent_comment_id"`
}

func TruncatePreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
