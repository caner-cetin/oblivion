package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caner-cetin/oblivion/internal"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	playgroundUpCmd = &cobra.Command{
		Use: "up",
		Run: WrapCommandWithResources(playgroundUp, ResourceConfig{Resources: []ResourceType{ResourceDocker, ResourceOnePassword}, Networks: []Network{NetworkDatabase, NetworkLoki}}),
	}

	playgroundCmd = &cobra.Command{
		Use: "playground",
	}
)

func getPlaygroundCmd() *cobra.Command {
	playgroundCmd.AddCommand(playgroundUpCmd)
	return playgroundCmd
}

func playgroundUp(cmd *cobra.Command, args []string) {
	app := GetApp(cmd)
	exists, err := app.containerExists(cfg.Playground.Backend.ContainerName)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	if exists {
		color.Cyan("playground backend running")
		return
	}
	tmp_repo_dir := filepath.Join(os.TempDir(), "code.cansu.dev")
	if err := app.pullRepo(tmp_repo_dir); err != nil {
		log.Error().Err(err).Send()
		return
	}
	backend_dir := filepath.Join(tmp_repo_dir, "backend")
	repo_fs := os.DirFS(backend_dir)
	if err := app.buildImage(repo_fs, backend_dir, cfg.Playground.Backend.ImageName, "Dockerfile"); err != nil {
		log.Error().Err(err).Msg("failed to build image")
		return
	}

	pg_secrets, err := app.loadPostgresSecrets(internal.Ptr("/Postgres/Playground/username"), internal.Ptr("/Postgres/Playground/password"))
	if err != nil {
		log.Error().Err(err).Msg("failed to get postgres secrets")
		return
	}
	redis_ref := app.Vault.Prefix + "/Redis/password"
	hf_key_ref := app.Vault.Prefix + "/Hugging Face/API Key"
	secrets, err := app.Vault.Client.Secrets().ResolveAll(app.Context, []string{redis_ref, hf_key_ref})
	if err != nil {
		log.Error().Err(err).Msg("failed to get secrets")
		return
	}
	resp, err := app.Docker.Client.ContainerCreate(app.Context,
		&container.Config{
			Image:        cfg.Playground.Backend.ImageName,
			ExposedPorts: nat.PortSet{nat.Port("6767/tcp"): struct{}{}},
			Env: []string{
				// sorry for this sequence
				"HF_TOKEN=" + secrets.IndividualResponses[hf_key_ref].Content.Secret,
				"HF_MODEL_URL=" + cfg.Playground.Backend.HFModelUrl,
				"REDIS_URL=" + fmt.Sprintf("redis://%s:%s/2", cfg.Dragonfly.ContainerName, cfg.Dragonfly.Port),
				"REDIS_PASSWORD=" + secrets.IndividualResponses[redis_ref].Content.Secret,
				"DATABASE_URL=" + fmt.Sprintf("postgres://%s:%s@%s:%s/playground?sslmode=disable",
					pg_secrets.Role.User,
					pg_secrets.Role.Password,
					cfg.Postgres.Primary.Name,
					cfg.Postgres.Primary.Port),
				"LOKI_URL=" + fmt.Sprintf("http://%s:%s/loki/api/v1/push", cfg.Observer.ContainerNames.Loki, cfg.Observer.Ports.Loki),
			},

			Cmd: []string{"/app"},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/var/run/docker.sock",
					Target: "/var/run/docker.sock",
				},
				{
					Type:   mount.TypeBind,
					Source: "/sys/fs/cgroup",
					Target: "/sys/fs/cgroup",
				},
				{
					Type:   mount.TypeBind,
					Source: "/run/systemd/private",
					Target: "/run/systemd/private",
				},
				{
					Type:   mount.TypeBind,
					Source: "/tmp",
					Target: "/tmp",
				},
			},
			PortBindings: nat.PortMap{
				nat.Port("6767/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Playground.Backend.Port}},
			},
			SecurityOpt: []string{"seccomp:unconfined"},
			CapAdd:      []string{"SYS_ADMIN", "DAC_OVERRIDE", "SYS_RESOURCE"},
			Cgroup:      container.CgroupSpec("host"),
			PidMode:     container.PidMode("host"),
		},
		&network.NetworkingConfig{
			EndpointsConfig: app.getNetworks(cfg.Networks.LokiNetworkName, cfg.Networks.DatabaseNetworkName),
		},
		nil,
		cfg.Playground.Backend.ContainerName,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create container")
		return
	}
	if err := app.Docker.Client.ContainerStart(app.Context, resp.ID, container.StartOptions{}); err != nil {
		log.Error().Err(err).Msg("failed to start container")
		return
	}
}
