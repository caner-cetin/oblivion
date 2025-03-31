package cmd

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	redisUpCmd = &cobra.Command{
		Use: "up",
		Run: WrapCommandWithResources(redisUp, ResourceConfig{Resources: []ResourceType{ResourceDocker, ResourceOnePassword}, Networks: []Network{NetworkDatabase}}),
	}
	redisCmd = &cobra.Command{
		Use: "redis",
	}
)

func getRedisCmd() *cobra.Command {
	redisCmd.AddCommand(redisUpCmd)
	return redisCmd
}

func redisUp(cmd *cobra.Command, args []string) {
	app := GetApp(cmd)
	exists, err := app.containerExists(cfg.Dragonfly.ContainerName)
	if err != nil {
		log.Error().Err(err).Msg("failed to check existence of redis container")
	}
	if exists {
		color.Cyan("redis running")
		return
	}
	password, err := app.Vault.Client.Secrets().Resolve(app.Context, app.Vault.Prefix+"/Redis/password")
	if err != nil {
		log.Error().Err(err).Msg("failed to get redis password")
		return
	}
	resp, err := app.Docker.Client.ContainerCreate(app.Context,
		&container.Config{
			Image: cfg.Dragonfly.Image,
			Cmd:   []string{"dragonfly", "--requirepass", password},
			Env: []string{
				"REDIS_PASSWORD=" + password,
			},
		},
		&container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: "dragonflydata",
					Target: "/data",
				},
			},
			PortBindings: nat.PortMap{
				nat.Port("6379/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Dragonfly.Port}},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: app.getNetworks(cfg.Networks.DatabaseNetworkName),
		},
		nil,
		cfg.Dragonfly.ContainerName)
	if err != nil {
		log.Error().Err(err).Msg("failed to create redis container")
		return
	}
	if err := app.Docker.Client.ContainerStart(app.Context, resp.ID, container.StartOptions{}); err != nil {
		log.Error().Err(err).Msg("failed to start redis container")
		return
	}
}
