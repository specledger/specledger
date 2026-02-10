package commands

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/specledger/specledger/pkg/cli/auth"
	"github.com/spf13/cobra"
)

const (
	// ProductionAuthURL is the production SpecLedger authentication URL
	ProductionAuthURL = "https://app.specledger.io/cli/auth"
	// DevAuthURL is the development authentication URL
	DevAuthURL = "http://localhost:3000/cli/auth"
	// AuthTimeout is the maximum time to wait for browser authentication
	AuthTimeout = 5 * time.Minute
)

// getAuthURL returns the authentication URL based on environment
// Priority: SPECLEDGER_AUTH_URL env > SPECLEDGER_ENV=dev > production default
func getAuthURL() string {
	// Allow explicit override
	if envURL := os.Getenv("SPECLEDGER_AUTH_URL"); envURL != "" {
		return envURL
	}
	// Check if running in dev mode
	if env := os.Getenv("SPECLEDGER_ENV"); env == "dev" || env == "development" {
		return DevAuthURL
	}
	return ProductionAuthURL
}

// VarAuthCmd represents the auth command group
var VarAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long: `Manage authentication for SpecLedger CLI.

Authentication is required for accessing protected features like
private specifications and remote synchronization.

Examples:
  sl auth login   # Sign in via browser
  sl auth logout  # Sign out and clear tokens
  sl auth status  # Check authentication status`,
}

// VarAuthLoginCmd represents the login command
var VarAuthLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in via browser",
	Long: `Open your browser to sign in to SpecLedger.

This command will:
1. Start a temporary local server to receive authentication
2. Open your default browser to the SpecLedger sign-in page
3. Wait for you to complete authentication
4. Store your credentials securely

If the browser doesn't open automatically, you can manually navigate
to the URL shown in the terminal.`,
	RunE: runLogin,
}

// VarAuthLogoutCmd represents the logout command
var VarAuthLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out and clear tokens",
	Long:  `Sign out of SpecLedger and remove stored credentials.`,
	RunE:  runLogout,
}

// VarAuthStatusCmd represents the status command
var VarAuthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Display the current authentication status and user information.`,
	RunE:  runStatus,
}

// VarAuthRefreshCmd represents the refresh command
var VarAuthRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token",
	Long:  `Manually refresh the access token using the stored refresh token.`,
	RunE:  runRefresh,
}

// VarAuthSupabaseCmd represents the supabase config command
var VarAuthSupabaseCmd = &cobra.Command{
	Use:   "supabase",
	Short: "Show Supabase configuration",
	Long:  `Display Supabase URL and anon key for API access.`,
	RunE:  runSupabase,
}

// VarAuthTokenCmd represents the token command (for scripts)
var VarAuthTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Print access token (for scripts)",
	Long: `Print the current access token to stdout.

This command is designed for use in scripts and automation.
It outputs only the token with no other text, making it easy
to capture in a variable:

  ACCESS_TOKEN=$(sl auth token)

The token is automatically refreshed if expired.`,
	RunE: runToken,
}

func init() {
	VarAuthCmd.AddCommand(VarAuthLoginCmd, VarAuthLogoutCmd, VarAuthStatusCmd, VarAuthRefreshCmd, VarAuthSupabaseCmd, VarAuthTokenCmd)

	VarAuthSupabaseCmd.Flags().Bool("url", false, "Print only the Supabase URL")
	VarAuthSupabaseCmd.Flags().Bool("key", false, "Print only the Supabase anon key")

	VarAuthLoginCmd.Flags().Bool("dev", false, "Use development server (localhost:3000)")
	VarAuthLoginCmd.Flags().String("token", "", "Authenticate with an access token (for CI/headless environments)")
	VarAuthLoginCmd.Flags().String("refresh", "", "Authenticate with a refresh token (exchanges for access token)")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Check for token-based authentication (CI/headless)
	accessToken, _ := cmd.Flags().GetString("token")
	if accessToken != "" {
		return runAccessTokenLogin(accessToken)
	}

	refreshToken, _ := cmd.Flags().GetString("refresh")
	if refreshToken != "" {
		return runRefreshTokenLogin(refreshToken)
	}

	// Check if already authenticated - inform user but proceed with re-auth
	creds, err := auth.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to check existing credentials: %w", err)
	}

	if creds != nil && creds.IsValid() && !creds.IsExpired() {
		fmt.Printf("Currently signed in as %s. Re-authenticating...\n", creds.UserEmail)
		fmt.Println()
	}

	// Build auth URL with callback
	useDev, _ := cmd.Flags().GetBool("dev")
	var authURL string
	if useDev {
		authURL = DevAuthURL
	} else {
		authURL = getAuthURL()
	}
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		return fmt.Errorf("invalid auth URL: %w", err)
	}

	// Get frontend base URL for redirect after callback
	frontendURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Start callback server
	server, err := auth.NewCallbackServer(frontendURL)
	if err != nil {
		return fmt.Errorf("failed to start authentication server: %w", err)
	}
	defer func() { _ = server.Shutdown() }()

	server.Start()

	q := parsedURL.Query()
	q.Set("callback", server.CallbackURL())
	parsedURL.RawQuery = q.Encode()

	loginURL := parsedURL.String()

	// Open browser
	fmt.Println("Opening browser for authentication...")
	fmt.Println()

	if err := auth.OpenBrowser(loginURL); err != nil {
		fmt.Println("Could not open browser automatically.")
		fmt.Println()
		fmt.Println("Please open this URL in your browser:")
		fmt.Printf("  %s\n", loginURL)
		fmt.Println()
	}

	fmt.Println("Waiting for authentication...")
	fmt.Printf("(timeout: %s)\n", AuthTimeout)
	fmt.Println()
	fmt.Println("If callback fails, you can manually authenticate with:")
	fmt.Println("  sl auth login --token <access_token>")
	fmt.Println("  sl auth login --refresh <refresh_token>")
	fmt.Println()

	// Wait for callback
	result, err := server.WaitForCallback(AuthTimeout)
	if err != nil {
		fmt.Println()
		fmt.Println("Callback failed. You can try manual authentication:")
		fmt.Println("  1. Copy the token from the browser")
		fmt.Println("  2. Run: sl auth login --token <your_token>")
		fmt.Println()
		return fmt.Errorf("authentication failed: %w", err)
	}

	expiresIn := result.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600 // Default 1 hour
	}

	// Save credentials
	credentials := &auth.Credentials{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    expiresIn,
		CreatedAt:    time.Now().Unix(),
		UserEmail:    result.UserEmail,
		UserID:       result.UserID,
	}

	if err := auth.SaveCredentials(credentials); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("Authentication successful!")
	fmt.Printf("Signed in as: %s\n", result.UserEmail)
	fmt.Printf("Credentials stored at: %s\n", auth.GetCredentialsPath())

	return nil
}

// runAccessTokenLogin handles authentication via access token (for CI/headless environments)
func runAccessTokenLogin(accessToken string) error {
	fmt.Println("Authenticating with access token...")

	// Save credentials with the provided access token
	credentials := &auth.Credentials{
		AccessToken: accessToken,
		ExpiresIn:   3600, // Default 1 hour, will be refreshed on next use if expired
		CreatedAt:   time.Now().Unix(),
	}

	if err := auth.SaveCredentials(credentials); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("Authentication successful!")
	fmt.Printf("Credentials stored at: %s\n", auth.GetCredentialsPath())

	return nil
}

// runRefreshTokenLogin handles authentication via refresh token (exchanges for access token)
func runRefreshTokenLogin(refreshToken string) error {
	fmt.Println("Authenticating with refresh token...")

	// Use refresh token to get valid credentials
	creds, err := auth.RefreshAccessToken(refreshToken)
	if err != nil {
		return fmt.Errorf("token authentication failed: %w", err)
	}

	// Save credentials
	if err := auth.SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("Authentication successful!")
	if creds.UserEmail != "" {
		fmt.Printf("Signed in as: %s\n", creds.UserEmail)
	}
	fmt.Printf("Credentials stored at: %s\n", auth.GetCredentialsPath())

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Check if authenticated
	creds, err := auth.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to check credentials: %w", err)
	}

	if creds == nil {
		fmt.Println("Not currently signed in.")
		return nil
	}

	email := creds.UserEmail

	// Delete credentials
	if err := auth.DeleteCredentials(); err != nil {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	fmt.Printf("Signed out successfully. (was: %s)\n", email)

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	creds, err := auth.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to check credentials: %w", err)
	}

	if creds == nil || !creds.IsValid() {
		fmt.Println("Status: Not signed in")
		fmt.Println()
		fmt.Println("Use 'sl auth login' to sign in.")
		return nil
	}

	fmt.Println("Status: Signed in")
	fmt.Printf("Email:  %s\n", creds.UserEmail)

	if creds.IsExpired() {
		fmt.Println("Token:  Expired (will refresh on next request)")
	} else {
		remaining := time.Until(creds.ExpiresAt()).Round(time.Minute)
		fmt.Printf("Token:  Valid (expires in %s)\n", remaining)
		fmt.Printf("Credentials: %s\n", auth.GetCredentialsPath())
	}

	return nil
}

func runRefresh(cmd *cobra.Command, args []string) error {
	fmt.Println("Refreshing access token...")

	token, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Reload credentials to show updated info
	creds, _ := auth.LoadCredentials()
	if creds != nil {
		remaining := time.Until(creds.ExpiresAt()).Round(time.Minute)
		fmt.Println("Token refreshed successfully!")
		fmt.Printf("Email:   %s\n", creds.UserEmail)
		fmt.Printf("Expires: in %s\n", remaining)
	} else {
		fmt.Printf("Token refreshed (length: %d)\n", len(token))
	}

	return nil
}

func runSupabase(cmd *cobra.Command, args []string) error {
	urlOnly, _ := cmd.Flags().GetBool("url")
	keyOnly, _ := cmd.Flags().GetBool("key")

	supabaseURL := auth.GetSupabaseURL()
	supabaseKey := auth.GetSupabaseAnonKey()

	if urlOnly {
		fmt.Println(supabaseURL)
		return nil
	}

	if keyOnly {
		fmt.Println(supabaseKey)
		return nil
	}

	// Print both
	fmt.Printf("SUPABASE_URL=%s\n", supabaseURL)
	fmt.Printf("SUPABASE_ANON_KEY=%s\n", supabaseKey)
	return nil
}

func runToken(cmd *cobra.Command, args []string) error {
	token, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Output only the token, no other text (for use in scripts)
	fmt.Print(token)
	return nil
}
