package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// groupsCmd represents the groups command
var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Group management",
	Long:  `Commands for managing WhatsApp groups`,
}

var groupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	Long:  `List all WhatsApp groups the user is a member of`,
	RunE:  runGroupsList,
}

var groupsRenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename a group",
	Long:  `Rename a WhatsApp group`,
	RunE:  runGroupsRename,
}

// groupsFlags holds the flags for groups commands
var groupsFlags = struct {
	jid        string
	name       string
	jsonOutput bool
}{}

func init() {
	groupsListCmd.Flags().BoolVar(&groupsFlags.jsonOutput, "json", false, "Output in JSON format")

	groupsRenameCmd.Flags().StringVarP(&groupsFlags.jid, "jid", "j", "", "Group JID (required)")
	groupsRenameCmd.Flags().StringVarP(&groupsFlags.name, "name", "n", "", "New group name (required)")
	groupsRenameCmd.Flags().BoolVar(&groupsFlags.jsonOutput, "json", false, "Output in JSON format")

	groupsRenameCmd.MarkFlagRequired("jid")
	groupsRenameCmd.MarkFlagRequired("name")

	groupsCmd.AddCommand(groupsListCmd)
	groupsCmd.AddCommand(groupsRenameCmd)
}

// GroupInfo represents formatted group information
type GroupInfo struct {
	JID          string   `json:"jid"`
	Name         string   `json:"name"`
	ParticipantCount int   `json:"participant_count"`
	IsCommunity  bool     `json:"is_community"`
}

// GroupsListResult represents the result of listing groups
type GroupsListResult struct {
	Success bool        `json:"success"`
	Count   int         `json:"count"`
	Groups  []*GroupInfo `json:"groups"`
	Error   string      `json:"error,omitempty"`
}

// GroupRenameResult represents the result of renaming a group
type GroupRenameResult struct {
	Success    bool   `json:"success"`
	OldName    string `json:"old_name,omitempty"`
	NewName    string `json:"new_name,omitempty"`
	JID        string `json:"jid"`
	Error      string `json:"error,omitempty"`
}

func runGroupsList(cmd *cobra.Command, args []string) error {
	waClient := GetClient()
	if waClient == nil {
		return fmt.Errorf("client not initialized. Run 'wacli auth' first")
	}

	if !waClient.IsConnected() {
		return fmt.Errorf("client not connected. Run 'wacli auth' first")
	}

	ctx := context.Background()
	groups, err := waClient.GetGroups(ctx)
	if err != nil {
		result := GroupsListResult{Success: false, Error: err.Error()}
		if groupsFlags.jsonOutput {
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
		} else {
			fmt.Fprintf(os.Stderr, "✗ Failed to get groups: %s\n", err.Error())
		}
		return err
	}

	groupInfos := make([]*GroupInfo, 0, len(groups))
	for _, g := range groups {
		groupInfos = append(groupInfos, &GroupInfo{
			JID:             g.JID.String(),
			Name:            g.Name,
			ParticipantCount: len(g.Participants),
		})
	}

	result := GroupsListResult{
		Success: true,
		Count:   len(groupInfos),
		Groups:  groupInfos,
	}

	if groupsFlags.jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		printGroupsList(&result)
	}

	return nil
}

func printGroupsList(result *GroupsListResult) {
	if result.Count == 0 {
		fmt.Println("No groups found.")
		return
	}

	fmt.Printf("Found %d group(s)\n\n", result.Count)
	for _, g := range result.Groups {
		participantStr := "participant"
		if g.ParticipantCount != 1 {
			participantStr = "participants"
		}
		fmt.Printf("👥 %s\n", g.Name)
		fmt.Printf("   JID: %s\n", g.JID)
		fmt.Printf("   %d %s\n\n", g.ParticipantCount, participantStr)
	}
}

func runGroupsRename(cmd *cobra.Command, args []string) error {
	waClient := GetClient()
	if waClient == nil {
		return fmt.Errorf("client not initialized. Run 'wacli auth' first")
	}

	if !waClient.IsConnected() {
		return fmt.Errorf("client not connected. Run 'wacli auth' first")
	}

	ctx := context.Background()

	// Get current group info
	groupInfo, err := waClient.GetGroupByJID(ctx, groupsFlags.jid)
	if err != nil {
		result := GroupRenameResult{Success: false, JID: groupsFlags.jid, Error: err.Error()}
		if groupsFlags.jsonOutput {
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
		} else {
			fmt.Fprintf(os.Stderr, "✗ Failed to get group: %s\n", err.Error())
		}
		return err
	}

	oldName := groupInfo.Name

	// Update group name via client
	err = waClient.UpdateGroupName(ctx, groupsFlags.jid, groupsFlags.name)
	if err != nil {
		result := GroupRenameResult{Success: false, JID: groupsFlags.jid, Error: err.Error()}
		if groupsFlags.jsonOutput {
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
		} else {
			fmt.Fprintf(os.Stderr, "✗ Failed to rename group: %s\n", err.Error())
		}
		return err
	}

	result := GroupRenameResult{
		Success: true,
		OldName: oldName,
		NewName: groupsFlags.name,
		JID:     groupsFlags.jid,
	}

	if groupsFlags.jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("✓ Group renamed successfully\n")
		fmt.Printf("  JID: %s\n", groupsFlags.jid)
		fmt.Printf("  Old name: %s\n", oldName)
		fmt.Printf("  New name: %s\n", groupsFlags.name)
	}

	return nil
}