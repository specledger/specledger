package revise

// ReviewComment maps to the review_comments table in Supabase.
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

// ProcessedComment is an in-memory struct for a comment the user chose to process.
type ProcessedComment struct {
	Comment  ReviewComment
	Guidance string
	Index    int // 1-based display index
}

// RevisionContext is the template rendering context for the combined prompt.
type RevisionContext struct {
	SpecKey  string
	Comments []PromptComment
}

// ThreadReply is a single reply in a comment thread.
type ThreadReply struct {
	ID         string // Reply UUID (for cascade resolution)
	AuthorName string
	Content    string
	CreatedAt  string
}

// PromptComment is a single comment entry in the revision prompt template.
type PromptComment struct {
	Index    int    // 1-based display index
	ID       string // Comment UUID (internal, for resolution)
	FilePath string
	Target   string // selected_text, "Line N", or "General"
	Feedback string // Comment content
	Guidance string // Optional user guidance
	Replies  []ThreadReply
}

// AutoFixture is the fixture file structure for non-interactive automation mode.
type AutoFixture struct {
	Branch   string           `json:"branch"`
	Comments []FixtureComment `json:"comments"`
}

// FixtureComment is a single comment entry in the automation fixture.
type FixtureComment struct {
	FilePath     string `json:"file_path"`
	SelectedText string `json:"selected_text"`
	Guidance     string `json:"guidance"`
}
