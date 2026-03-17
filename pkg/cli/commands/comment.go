package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/specledger/specledger/pkg/cli/auth"
	"github.com/specledger/specledger/pkg/cli/comment"
	cligit "github.com/specledger/specledger/pkg/cli/git"
	"github.com/spf13/cobra"
)

var VarCommentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Manage review comments",
	Long: `Manage review comments from Supabase.

Subcommands:
  list     List review comments (compact or JSON format)
  show     Show full comment details with thread replies
  reply    Reply to a comment thread
  resolve  Mark comments as resolved (--reason required)`,
	Args:         cobra.NoArgs,
	RunE:         func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	SilenceUsage: true,
}

var commentListCmd = &cobra.Command{
	Use:   "list [branch-name]",
	Short: "List review comments",
	Long: `List unresolved review comments for a spec.

Output formats:
  Default: Compact format with truncated previews (80 chars)
  --json:  Full JSON array with all comment details

Status filters:
  --status open      Only unresolved comments (default)
  --status resolved  Only resolved comments
  --status all       All comments

Exit codes:
  0: Success (including no comments)
  1: Auth failure (silent exit for agent integration)

Examples:
  sl comment list
  sl comment list 601-cli-skills
  sl comment list --json --status open
  sl comment list --status resolved`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runCommentList,
	SilenceUsage: true,
}

var commentShowCmd = &cobra.Command{
	Use:   "show <comment-id> [comment-id...]",
	Short: "Show full comment details",
	Long: `Show full comment details with thread replies.

Output formats:
  Default: Human-readable format with all details
  --json:  Full JSON object with comment and replies

Arguments:
  One or more comment IDs to display

Examples:
  sl comment show abc123
  sl comment show abc123 def456
  sl comment show abc123 --json`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runCommentShow,
	SilenceUsage: true,
}

var commentReplyCmd = &cobra.Command{
	Use:   "reply <comment-id> <message>",
	Short: "Reply to a comment thread",
	Long: `Post a reply to an existing comment thread.

Arguments:
  comment-id: The ID of the parent comment
  message:    The reply message text

Output formats:
  Default: Success message with reply ID
  --json:  JSON object with reply_id and timestamp

Examples:
  sl comment reply abc123 "Fixed in commit def456"
  sl comment reply abc123 "Addressed the issue" --json`,
	Args:         cobra.ExactArgs(2),
	RunE:         runCommentReply,
	SilenceUsage: true,
}

var (
	commentListJSON      bool
	commentListStatus    string
	commentShowJSON      bool
	commentReplyJSON     bool
	commentResolveJSON   bool
	commentResolveReason string
)

func init() {
	commentListCmd.Flags().BoolVar(&commentListJSON, "json", false, "Output as JSON array")
	commentListCmd.Flags().StringVar(&commentListStatus, "status", "open", "Filter by status: open, resolved, all")

	commentShowCmd.Flags().BoolVar(&commentShowJSON, "json", false, "Output as JSON")

	commentReplyCmd.Flags().BoolVar(&commentReplyJSON, "json", false, "Output as JSON")

	commentResolveCmd.Flags().BoolVar(&commentResolveJSON, "json", false, "Output as JSON")
	commentResolveCmd.Flags().StringVar(&commentResolveReason, "reason", "", "Resolution reason (required — posted as reply before resolving)")
	_ = commentResolveCmd.MarkFlagRequired("reason")

	VarCommentCmd.AddCommand(commentListCmd)
	VarCommentCmd.AddCommand(commentShowCmd)
	VarCommentCmd.AddCommand(commentReplyCmd)
	VarCommentCmd.AddCommand(commentResolveCmd)
}

func runCommentList(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		os.Exit(1)
		return nil
	}

	client := comment.NewClient(accessToken)

	var specKey string
	if len(args) > 0 {
		specKey = args[0]
	} else {
		currentBranch, err := cligit.GetCurrentBranch(cwd)
		if err != nil {
			os.Exit(1)
			return nil
		}
		specKey = currentBranch
	}

	repoOwner, repoName, err := cligit.GetRepoOwnerName(cwd)
	if err != nil {
		return fmt.Errorf("failed to get repo info: %w", err)
	}

	project, err := client.GetProject(repoOwner, repoName)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	spec, err := client.GetSpec(project.ID, specKey)
	if err != nil {
		return fmt.Errorf("failed to get spec: %w", err)
	}

	change, err := client.GetChange(spec.ID)
	if err != nil {
		return fmt.Errorf("failed to get change: %w", err)
	}

	comments, err := fetchCommentsByStatus(client, change.ID, commentListStatus)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	if commentListJSON {
		return outputCommentsJSON(comments)
	}

	return outputCommentsCompact(comments, client, change.ID)
}

func fetchCommentsByStatus(client *comment.Client, changeID, status string) ([]comment.ReviewComment, error) {
	switch status {
	case "open":
		return client.FetchComments(changeID)
	case "resolved":
		return fetchResolvedComments(client, changeID)
	case "all":
		open, err := client.FetchComments(changeID)
		if err != nil {
			return nil, err
		}
		resolved, err := fetchResolvedComments(client, changeID)
		if err != nil {
			return nil, err
		}
		return append(open, resolved...), nil
	default:
		return nil, fmt.Errorf("invalid status filter: %s (use: open, resolved, all)", status)
	}
}

func fetchResolvedComments(client *comment.Client, changeID string) ([]comment.ReviewComment, error) {
	return client.FetchResolvedComments(changeID)
}

func outputCommentsJSON(comments []comment.ReviewComment) error {
	type CommentOutput struct {
		ID           string `json:"id"`
		FilePath     string `json:"file_path"`
		Line         *int   `json:"line"`
		StartLine    *int   `json:"start_line"`
		Content      string `json:"content"`
		SelectedText string `json:"selected_text"`
		AuthorName   string `json:"author_name"`
		AuthorEmail  string `json:"author_email"`
		IsResolved   bool   `json:"is_resolved"`
		CreatedAt    string `json:"created_at"`
	}

	output := make([]CommentOutput, 0, len(comments))
	for _, c := range comments {
		output = append(output, CommentOutput{
			ID:           c.ID,
			FilePath:     c.FilePath,
			Line:         c.Line,
			StartLine:    c.StartLine,
			Content:      c.Content,
			SelectedText: c.SelectedText,
			AuthorName:   c.AuthorName,
			AuthorEmail:  c.AuthorEmail,
			IsResolved:   c.IsResolved,
			CreatedAt:    c.CreatedAt,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputCommentsCompact(comments []comment.ReviewComment, client *comment.Client, changeID string) error {
	if len(comments) == 0 {
		fmt.Printf("0 comments\n")
		return nil
	}

	replies, _ := client.FetchReplies(changeID)
	replyMap := comment.BuildReplyMap(replies)

	artifacts := make(map[string]struct{})
	for _, c := range comments {
		artifacts[c.FilePath] = struct{}{}

		lineStr := ""
		if c.StartLine != nil && c.Line != nil {
			lineStr = fmt.Sprintf("%d-%d", *c.StartLine, *c.Line)
		} else if c.Line != nil {
			lineStr = fmt.Sprintf("%d", *c.Line)
		} else {
			lineStr = "-"
		}

		content := c.Content
		if len(content) > 80 {
			content = content[:77] + "..."
		}
		content = strings.ReplaceAll(content, "\n", " ")

		author := c.AuthorName
		if author == "" {
			author = c.AuthorEmail
		}

		replyInfo := ""
		if reps := replyMap[c.ID]; len(reps) > 0 {
			noun := "reply"
			if len(reps) != 1 {
				noun = "replies"
			}
			replyInfo = fmt.Sprintf(" | %d %s", len(reps), noun)
		}

		fmt.Printf("%s | %s:%s | %s | %s%s\n",
			c.ID, c.FilePath, lineStr, content, author, replyInfo)
	}

	fmt.Printf("\n%d comment(s) across %d artifact(s)\n", len(comments), len(artifacts))
	return nil
}

func runCommentShow(cmd *cobra.Command, args []string) error {
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w\n\nRun 'sl auth login' to authenticate.", err)
	}

	client := comment.NewClient(accessToken)

	for i, commentID := range args {
		if i > 0 {
			fmt.Println("\n---")
		}

		c, err := client.FetchCommentByID(commentID)
		if err != nil {
			return fmt.Errorf("failed to fetch comment %s: %w", commentID, err)
		}

		replies, err := client.FetchRepliesByParentID(commentID)
		if err != nil {
			return fmt.Errorf("failed to fetch replies for comment %s: %w", commentID, err)
		}

		if commentShowJSON {
			if err := outputCommentJSON(c, replies); err != nil {
				return err
			}
		} else {
			if err := outputCommentHuman(c, replies); err != nil {
				return err
			}
		}
	}

	return nil
}

func outputCommentJSON(c *comment.ReviewComment, replies []comment.ReviewComment) error {
	type ReplyOutput struct {
		ID         string `json:"id"`
		Content    string `json:"content"`
		AuthorName string `json:"author_name"`
		CreatedAt  string `json:"created_at"`
	}

	type CommentOutput struct {
		ID           string        `json:"id"`
		FilePath     string        `json:"file_path"`
		Line         *int          `json:"line"`
		StartLine    *int          `json:"start_line"`
		Content      string        `json:"content"`
		SelectedText string        `json:"selected_text"`
		AuthorName   string        `json:"author_name"`
		AuthorEmail  string        `json:"author_email"`
		IsResolved   bool          `json:"is_resolved"`
		CreatedAt    string        `json:"created_at"`
		Replies      []ReplyOutput `json:"replies"`
	}

	output := CommentOutput{
		ID:           c.ID,
		FilePath:     c.FilePath,
		Line:         c.Line,
		StartLine:    c.StartLine,
		Content:      c.Content,
		SelectedText: c.SelectedText,
		AuthorName:   c.AuthorName,
		AuthorEmail:  c.AuthorEmail,
		IsResolved:   c.IsResolved,
		CreatedAt:    c.CreatedAt,
		Replies:      make([]ReplyOutput, 0, len(replies)),
	}

	for _, r := range replies {
		output.Replies = append(output.Replies, ReplyOutput{
			ID:         r.ID,
			Content:    r.Content,
			AuthorName: r.AuthorName,
			CreatedAt:  r.CreatedAt,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputCommentHuman(c *comment.ReviewComment, replies []comment.ReviewComment) error {
	fmt.Printf("Comment ID: %s\n", c.ID)
	fmt.Printf("File: %s", c.FilePath)

	if c.StartLine != nil && c.Line != nil {
		fmt.Printf(":%d-%d\n", *c.StartLine, *c.Line)
	} else if c.Line != nil {
		fmt.Printf(":%d\n", *c.Line)
	} else {
		fmt.Println()
	}

	fmt.Printf("Author: %s <%s>\n", c.AuthorName, c.AuthorEmail)
	fmt.Printf("Status: %s\n", map[bool]string{true: "Resolved", false: "Open"}[c.IsResolved])
	fmt.Printf("Created: %s\n", c.CreatedAt)

	if c.SelectedText != "" {
		fmt.Printf("\nSelected Text:\n%s\n", c.SelectedText)
	}

	fmt.Printf("\nComment:\n%s\n", c.Content)

	if len(replies) > 0 {
		fmt.Printf("\nThread Replies (%d):\n", len(replies))
		for i, r := range replies {
			fmt.Printf("\n%d. %s (%s)\n", i+1, r.AuthorName, r.CreatedAt)
			fmt.Printf("   %s\n", r.Content)
		}
	}

	return nil
}

func runCommentReply(cmd *cobra.Command, args []string) error {
	commentID := args[0]
	message := args[1]

	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w\n\nRun 'sl auth login' to authenticate.", err)
	}

	client := comment.NewClient(accessToken)

	reply, err := client.CreateReply(commentID, message)
	if err != nil {
		return fmt.Errorf("failed to post reply: %w", err)
	}

	if commentReplyJSON {
		type ReplyOutput struct {
			ID        string `json:"reply_id"`
			Timestamp string `json:"timestamp"`
		}

		output := ReplyOutput{
			ID:        reply.ID,
			Timestamp: reply.CreatedAt,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	fmt.Printf("Reply posted successfully\n")
	fmt.Printf("Reply ID: %s\n", reply.ID)
	fmt.Printf("Timestamp: %s\n", reply.CreatedAt)

	return nil
}

var commentResolveCmd = &cobra.Command{
	Use:   "resolve <comment-id> [comment-id...] --reason \"text\"",
	Short: "Mark comments as resolved with a reason",
	Long: `Mark one or more review comments as resolved.

A reason is required — it is posted as a reply before resolving.
When resolving a parent comment, all thread replies are also resolved (cascade).

Arguments:
  One or more comment IDs to resolve

Output formats:
  Default: Success message with resolved IDs
  --json:  JSON array with resolved comment IDs

Examples:
  sl comment resolve abc123 --reason "Fixed in PR #42"
  sl comment resolve abc123 def456 --reason "Batch resolved: all addressed in latest revision"
  sl comment resolve abc123 --reason "No action needed" --json`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runCommentResolve,
	SilenceUsage: true,
}

func runCommentResolve(cmd *cobra.Command, args []string) error {
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w\n\nRun 'sl auth login' to authenticate.", err)
	}

	client := comment.NewClient(accessToken)

	resolvedIDs := make([]string, 0, len(args))

	for _, commentID := range args {
		// Post reason as a reply before resolving (audit trail)
		if _, err := client.CreateReply(commentID, commentResolveReason); err != nil {
			return fmt.Errorf("failed to post resolution reason for %s: %w\n→ The comment was NOT resolved. Fix the reply issue first.", commentID, err)
		}

		replies, err := client.FetchRepliesByParentID(commentID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch replies for %s: %v\n", commentID, err)
		}

		if len(replies) > 0 {
			replyIDs := make([]string, 0, len(replies))
			for _, r := range replies {
				replyIDs = append(replyIDs, r.ID)
			}

			if err := client.ResolveCommentWithReplies(commentID, replyIDs); err != nil {
				return fmt.Errorf("failed to resolve comment %s with replies: %w", commentID, err)
			}

			allIDs := make([]string, 0, 1+len(replyIDs))
			allIDs = append(allIDs, commentID)
			allIDs = append(allIDs, replyIDs...)
			resolvedIDs = append(resolvedIDs, allIDs...)
		} else {
			if err := client.ResolveComment(commentID); err != nil {
				return fmt.Errorf("failed to resolve comment %s: %w", commentID, err)
			}
			resolvedIDs = append(resolvedIDs, commentID)
		}
	}

	if commentResolveJSON {
		type ResolveOutput struct {
			ResolvedIDs []string `json:"resolved_ids"`
		}

		output := ResolveOutput{
			ResolvedIDs: resolvedIDs,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	fmt.Printf("Resolved %d comment(s)\n", len(resolvedIDs))
	for _, id := range resolvedIDs {
		fmt.Printf("  - %s\n", id)
	}

	return nil
}
