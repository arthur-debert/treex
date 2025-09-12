package commands

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

//go:embed formats.help.txt
var formatsTopicContent string

//go:embed config.yaml
var configTopicContent string

// Topic represents a help topic
type Topic struct {
	Name        string
	Description string
	Content     string
}

// availableTopics lists all available help topics
var availableTopics = map[string]Topic{
	"formats": {
		Name:        "formats",
		Description: "Available output formats and their usage",
		Content:     formatsTopicContent,
	},
	"config": {
		Name:        "config",
		Description: "Default configuration file reference",
		Content:     configTopicContent,
	},
}

// topicsCmd represents the topics command
var topicsCmd = &cobra.Command{
	Use:     "topics [topic]",
	GroupID: "help",
	Short:   "List available help topics or show a specific topic",
	Long: `The topics command provides access to detailed help on various treex topics.

Without arguments, it lists all available topics.
With a topic name, it displays the topic content using your system pager.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTopicsCmd,
}

func init() {
	// Register the command with root
	rootCmd.AddCommand(topicsCmd)
}

// runTopicsCmd handles the topics command
func runTopicsCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// List all available topics
		return listTopics(cmd)
	}

	// Show specific topic
	topicName := args[0]
	return showTopic(cmd, topicName)
}

// listTopics displays all available topics
func listTopics(cmd *cobra.Command) error {
	cmd.Println("Available help topics:")
	cmd.Println()

	for _, topic := range availableTopics {
		cmd.Printf("  %-12s %s\n", topic.Name, topic.Description)
	}

	cmd.Println()
	cmd.Println("Use 'treex topics <topic>' to view a topic in detail.")
	return nil
}

// showTopic displays a specific topic using the system pager
func showTopic(cmd *cobra.Command, topicName string) error {
	topic, exists := availableTopics[topicName]
	if !exists {
		return fmt.Errorf("unknown topic: %s", topicName)
	}

	// Try to use system pager, fallback to direct output
	if err := displayWithPager(topic.Content); err != nil {
		// Fallback: just print to stdout
		cmd.Print(topic.Content)
	}

	return nil
}

// displayWithPager attempts to display content using the system pager
func displayWithPager(content string) error {
	// Determine pager command
	pager := os.Getenv("PAGER")
	if pager == "" {
		if runtime.GOOS == "windows" {
			pager = "more"
		} else {
			pager = "less"
		}
	}

	// Create pager command
	pagerCmd := exec.Command(pager)
	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr

	// Set up pipe to pager
	stdin, err := pagerCmd.StdinPipe()
	if err != nil {
		return err
	}

	// Start the pager
	if err := pagerCmd.Start(); err != nil {
		return err
	}

	// Write content to pager
	_, err = stdin.Write([]byte(content))
	if err != nil {
		_ = stdin.Close()
		return err
	}

	// Close stdin and wait for pager to finish
	_ = stdin.Close()
	return pagerCmd.Wait()
}