package cmd

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	networkUpCmd = &cobra.Command{
		Use: "up",
		Run: WrapCommandWithResources(networkUp, ResourceConfig{Resources: []ResourceType{ResourceDocker}, Networks: []Network{}}),
	}
	networkCmd = &cobra.Command{
		Use: "networks",
	}
)

func getNetworkCmd() *cobra.Command {
	networkCmd.AddCommand(networkUpCmd)
	return networkCmd
}

func networkUp(cmd *cobra.Command, args []string) {
	app := GetApp(cmd)
	if err := app.createNetworkIfNotExists(cfg.Networks.DatabaseNetworkName, nil); err != nil {
		log.Error().Err(err).Str("network_name", cfg.Networks.DatabaseNetworkName).Msg("failed to create network")
		return
	}
	if err := app.createNetworkIfNotExists(cfg.Networks.UptimeNetworkName, nil); err != nil {
		log.Error().Err(err).Str("network_name", cfg.Networks.UptimeNetworkName).Msg("failed to create network")
		return
	}
	if err := app.createNetworkIfNotExists(cfg.Networks.GrafanaNetworkName, nil); err != nil {
		log.Error().Err(err).Str("network_name", cfg.Networks.GrafanaNetworkName).Msg("failed to create network")
		return
	}
	if err := app.createNetworkIfNotExists(cfg.Networks.LokiNetworkName, nil); err != nil {
		log.Error().Err(err).Str("network_name", cfg.Networks.LokiNetworkName).Msg("failed to create network")
		return
	}

	color.Green("created required networks")
}

func (a *AppCtx) getNetworkIDByName(ctx context.Context, networkName string) (string, error) {
	networks, err := a.Docker.Client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %w", err)
	}

	for _, network := range networks {
		if network.Name == networkName {
			return network.ID, nil
		}
	}

	return "", fmt.Errorf("network with name '%s' not found", networkName)
}

func (a *AppCtx) getNetworks(networkNames ...string) map[string]*network.EndpointSettings {
	es := make(map[string]*network.EndpointSettings)
	for _, nw := range networkNames {
		es[nw] = a.Docker.Networks[nw]
	}
	return es
}
