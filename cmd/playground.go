package cmd

import "github.com/spf13/cobra"


var (
	playgroundUpCmd = &cobra.Command{
		Use:   "up",
		Run: WrapCommandWithResources(playgroundUp, ResourceConfig{Resources: []ResourceType{ResourceDocker}, Networks: []Network{NetworkDatabase, NetworkLoki}}),
		
	}

	playgroundCmd = &cobra.Command{
		Use: "playground",
	}
)

func getPlaygroundCmd() *cobra.Command {
	playgroundCmd.AddCommand(playgroundUpCmd)
	return playgroundCmd
}


func playgroundUp(cmd *cobra.Command, args[]string) {
}