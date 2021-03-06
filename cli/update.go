package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/coder/coder/codersdk"
)

func update() *cobra.Command {
	return &cobra.Command{
		Annotations: workspaceCommand,
		Use:         "update",
		Short:       "Update a workspace to the latest template version",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient(cmd)
			if err != nil {
				return err
			}
			workspace, err := namedWorkspace(cmd, client, args[0])
			if err != nil {
				return err
			}
			if !workspace.Outdated {
				_, _ = fmt.Printf("Workspace isn't outdated!\n")
				return nil
			}
			template, err := client.Template(cmd.Context(), workspace.TemplateID)
			if err != nil {
				return nil
			}
			before := time.Now()
			build, err := client.CreateWorkspaceBuild(cmd.Context(), workspace.ID, codersdk.CreateWorkspaceBuildRequest{
				TemplateVersionID: template.ActiveVersionID,
				Transition:        workspace.LatestBuild.Transition,
			})
			if err != nil {
				return err
			}
			logs, err := client.WorkspaceBuildLogsAfter(cmd.Context(), build.ID, before)
			if err != nil {
				return err
			}
			for {
				log, ok := <-logs
				if !ok {
					break
				}
				_, _ = fmt.Printf("Output: %s\n", log.Output)
			}
			return nil
		},
	}
}
