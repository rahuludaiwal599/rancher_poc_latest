/*
Copyright © 2023 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package cmd implements the rdctl commands
package cmd

import (
	"fmt"

	"github.com/rancher-sandbox/rancher-desktop/src/go/rdctl/pkg/client"
	"github.com/rancher-sandbox/rancher-desktop/src/go/rdctl/pkg/config"
	"github.com/spf13/cobra"
)

// installCmd represents the 'rdctl extensions install' command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install an RDX extension",
	Long: `rdctl extension install [--force] <image-id>
--force: avoid any interactivity.
The <image-id> is an image reference, e.g. splatform/epinio-docker-desktop:latest (the tag is optional).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return installExtension(args)
	},
}

func init() {
	extensionCmd.AddCommand(installCmd)
}

func installExtension(args []string) error {
	connectionInfo, err := config.GetConnectionInfo(false)
	if err != nil {
		return fmt.Errorf("failed to get connection info: %w", err)
	}
	rdClient := client.NewRDClient(connectionInfo)
	imageID := args[0]
	endpoint := fmt.Sprintf("/%s/extensions/install?id=%s", client.ApiVersion, imageID)
	// https://stackoverflow.com/questions/20847357/golang-http-client-always-escaped-the-url
	// Looks like http.NewRequest(method, url) escapes the URL

	result, errorPacket, err := client.ProcessRequestForAPI(rdClient.DoRequest("POST", endpoint))
	if errorPacket != nil || err != nil {
		return displayAPICallResult(result, errorPacket, err)
	}
	msg := "no output from server"
	if result != nil {
		msg = string(result)
	}
	fmt.Printf("Installing image %s: %s\n", imageID, msg)
	return nil
}
