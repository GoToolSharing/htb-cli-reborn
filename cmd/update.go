package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GoToolSharing/htb-cli/config"
	"github.com/GoToolSharing/htb-cli/utils"
	"github.com/spf13/cobra"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func updateTool(newVersion string) error {
	githubVersion := "https://api.github.com/repos/GoToolSharing/htb-cli/releases/latest"

	resp, err := utils.HTTPRequest(http.MethodGet, githubVersion, "", nil)
	if err != nil {
		return err
	}
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return err
	}
	if release.TagName != config.Version {
		message := fmt.Sprintf("A new update is now available ! (%s)", release.TagName)
		fmt.Println(message)
		fmt.Println("Update with : go install github.com/GoToolSharing/htb-cli@latest")
	} else {
		message := fmt.Sprintf("You're up to date ! (%s)", config.Version)
		fmt.Println(message)
	}

	return nil

}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		err := updateTool(config.Version)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
