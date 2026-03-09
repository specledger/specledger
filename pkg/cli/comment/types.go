package comment

type ReviewComment struct {
	ID              string `json:"id"`
	ChangeID        string `json:"change_id"`
	FilePath        string `json:"file_path"`
	Content         string `json:"content"`
	SelectedText    string `json:"selected_text"`
	Line            *int   `json:"line"`
	StartLine       *int   `json:"start_line"`
	IsResolved      bool   `json:"is_resolved"`
	AuthorID        string `json:"author_id"`
	AuthorName      string `json:"author_name"`
	AuthorEmail     string `json:"author_email"`
	ParentCommentID string `json:"parent_comment_id"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type ThreadReply struct {
	ID         string
	ParentID   string
	AuthorName string
	Content    string
	CreatedAt  string
}

type ReplyMap map[string][]ReviewComment
