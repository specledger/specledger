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
  resolve  Mark comments as resolved`,
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

var (
	commentListJSON   bool
	commentListStatus string
)

func init() {
	commentListCmd.Flags().BoolVar(&commentListJSON, "json", false, "Output as JSON array")
	commentListCmd.Flags().StringVar(&commentListStatus, "status", "open", "Filter by status: open, resolved, all")

	VarCommentCmd.AddCommand(commentListCmd)
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
