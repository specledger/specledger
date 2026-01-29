package commands

import (
	"github.com/spf13/cobra"
)

// VarGraphCmd represents the graph command
var VarGraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Display dependency graph",
}

// VarShowCmd represents the show command
var VarShowCmd = &cobra.Command{
	Use:   "show [--format <format>] [--include-transitive]",
	Short: "Show the dependency graph",
	Long:  `Display the complete dependency graph with all nodes and edges.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runShowGraph,
}

// VarExportCmd represents the export command
var VarExportCmd = &cobra.Command{
	Use:   "export --format <format> --output <file>",
	Short: "Export graph to file",
	Run:   runExportGraph,
}

// VarTransitiveCmd represents the transitive command
var VarTransitiveCmd = &cobra.Command{
	Use:   "transitive [--depth <n>]",
	Short: "Show transitive dependencies",
	Run:   runTransitiveDependencies,
}

func init() {
	VarGraphCmd.AddCommand(VarShowCmd, VarExportCmd, VarTransitiveCmd)

	VarShowCmd.Flags().StringP("format", "f", "text", "Output format: text, json, svg")
	VarShowCmd.Flags().BoolP("include-transitive", "t", false, "Include transitive dependencies")
	VarExportCmd.Flags().StringP("format", "f", "json", "Export format: json, svg, text")
	VarExportCmd.Flags().StringP("output", "o", "deps.svg", "Output file path")
	VarTransitiveCmd.Flags().IntP("depth", "d", 0, "Maximum depth (0 = unlimited)")
}

func runShowGraph(cmd *cobra.Command, args []string) {
	format, _ := cmd.Flags().GetString("format")
	cmd.Printf("Showing dependency graph (format: %s)\n", format)
}

func runExportGraph(cmd *cobra.Command, args []string) {
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	cmd.Printf("Exporting graph (format: %s, output: %s)\n", format, output)
}

func runTransitiveDependencies(cmd *cobra.Command, args []string) {
	depth, _ := cmd.Flags().GetInt("depth")
	cmd.Printf("Showing transitive dependencies (depth: %d)\n", depth)
}
