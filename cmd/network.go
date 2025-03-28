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
	exists, err := app.networkExists(cfg.Networks.DatabaseNetworkName)
	if err != nil {
		log.Error().Err(err).Msgf("failed to check existence of network %s", cfg.Networks.DatabaseNetworkName)
		return
	}
	if exists {
		color.Cyan("network %s already exists", cfg.Networks.DatabaseNetworkName)
	} else {
		resp, err := app.Docker.Client.NetworkCreate(app.Context, cfg.Networks.DatabaseNetworkName, network.CreateOptions{Driver: "bridge"})
		if err != nil {
			log.Error().Err(err).Msgf("failed to create network %s", cfg.Networks.DatabaseNetworkName)
			return
		}
		if resp.Warning != "" {
			log.Warn().Str("msg", resp.Warning).Msgf("warning from network %s", cfg.Networks.DatabaseNetworkName)
		}
		color.Green("created network %s", cfg.Networks.DatabaseNetworkName)
	}

	exists, err = app.networkExists(cfg.Networks.UptimeNetworkName)
	if err != nil {
		log.Error().Err(err).Msgf("failed to check existence of network %s", cfg.Networks.UptimeNetworkName)
		return
	}
	if exists {
		color.Cyan("network %s already exists", cfg.Networks.UptimeNetworkName)
	} else {
		resp, err := app.Docker.Client.NetworkCreate(app.Context, cfg.Networks.UptimeNetworkName, network.CreateOptions{Driver: "bridge"})
		if err != nil {
			log.Error().Err(err).Msgf("failed to create network %s", cfg.Networks.UptimeNetworkName)
			return
		}
		if resp.Warning != "" {
			log.Warn().Str("msg", resp.Warning).Msgf("warning from network %s", cfg.Networks.UptimeNetworkName)
		}
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
