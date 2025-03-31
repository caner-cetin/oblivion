package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/fatih/color"
	v1 "github.com/moby/docker-image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type userPasswordPair struct {
	User     string
	Password string
}

type postgresCredentials struct {
	Bouncer    userPasswordPair
	Replicator userPasswordPair
	Postgres   userPasswordPair
}

var (
	postgresUpCmd = &cobra.Command{
		Use: "up",
		Run: WrapCommandWithResources(postgresUp, ResourceConfig{[]ResourceType{ResourceDocker, ResourceOnePassword}, []Network{NetworkDatabase}}),
	}
	postgresCmd = &cobra.Command{
		Use: "postgres",
	}
)

func getPostgresCmd() *cobra.Command {
	postgresCmd.AddCommand(postgresUpCmd)
	return postgresCmd
}

func postgresUp(cmd *cobra.Command, args []string) {
	app := GetApp(cmd)
	credentials, err := app.loadPostgresSecrets()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	if err := credentials.startPrimary(&app); err != nil {
		log.Error().Err(err).Send()
		return
	}
	if err := credentials.startReplica(&app); err != nil {
		log.Error().Err(err).Send()
		return
	}
	if err := credentials.startBouncer(&app); err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (a *AppCtx) loadPostgresSecrets() (*postgresCredentials, error) {
	secrets, err := a.resolveSecrets(
		[]string{
			"/Postgres/Replicator/username",
			"/Postgres/Replicator/password",
			"/Postgres/Root/username",
			"/Postgres/Root/password",
			"/Postgres/Bouncer/username",
			"/Postgres/Bouncer/password",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve postgres credentials: %w", err)
	}
	var credentials = postgresCredentials{
		Replicator: userPasswordPair{
			User:     secrets[0],
			Password: secrets[1],
		},
		Postgres: userPasswordPair{
			User:     secrets[2],
			Password: secrets[3],
		},
		Bouncer: userPasswordPair{
			User:     secrets[4],
			Password: secrets[5],
		},
	}
	return &credentials, nil
}

func (c *postgresCredentials) startPrimary(app *AppCtx) error {
	exists, err := app.containerExists(cfg.Postgres.Primary.Name)
	if err != nil {
		return fmt.Errorf("failed to check if primary container exists: %w", err)
	}
	if exists {
		color.Green("primary pg container running")
		return nil
	}
	response, err := app.Docker.Client.ContainerCreate(app.Context,
		&container.Config{
			AttachStdout: true,
			AttachStderr: true,
			AttachStdin:  false,
			OpenStdin:    false,
			Image:        cfg.Postgres.Primary.Image,
			Cmd: []string{
				"-c",
				"wal_level=replica",
				"-c",
				"max_wal_senders=10",
				"-c",
				"max_replication_slots=10",
			},
			Env: []string{
				fmt.Sprintf("POSTGRES_DB=%s", cfg.Postgres.DB),
				fmt.Sprintf("POSTGRES_USER=%s", c.Postgres.User),
				fmt.Sprintf("POSTGRES_PASSWORD=%s", c.Postgres.Password),
				"POSTGRES_HOST_AUTH_METHOD=scram-sha-256",
				fmt.Sprintf("POSTGRES_REPLICATION_USER=%s", c.Replicator.User),
				fmt.Sprintf("POSTGRES_REPLICATION_PASSWORD=%s", c.Replicator.Password),
			},
			Healthcheck: postgres_healthcheck,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("5432/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Postgres.Primary.Port}},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: cfg.Postgres.Primary.Volume,
					Target: "/var/lib/postgresql/data",
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: app.getNetworks(cfg.Networks.DatabaseNetworkName),
		},
		nil,
		cfg.Postgres.Primary.Name)
	if err != nil {
		return fmt.Errorf("failed to create primary postgres container: %w", err)
	}
	log.Info().Str("id", response.ID).Msg("created primary postgres container")
	err = app.Docker.Client.ContainerStart(app.Context, response.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start primary postgres container: %w", err)
	}
	cancel := app.spawnLogs(response.ID)
	defer cancel()
	if err := app.waitForContainerHealthWithConfig(response.ID, postgres_healthcheck); err != nil {
		return fmt.Errorf("start of primary postgres failed: %w", err)
	}
	return nil
}

func (c *postgresCredentials) startReplica(app *AppCtx) error {
	exists, err := app.containerExists(cfg.Postgres.Replica.Name)
	if err != nil {
		return fmt.Errorf("failed to check if replica container exists: %w", err)
	}
	if exists {
		color.Green("replica pg container running")
		return nil
	}
	response, err := app.Docker.Client.ContainerCreate(app.Context,
		&container.Config{
			AttachStdout: true,
			AttachStderr: true,
			AttachStdin:  false,
			OpenStdin:    false,
			Image:        cfg.Postgres.Replica.Image,
			Cmd: []string{
				"-c",
				"wal_level=replica",
				"-c",
				"max_wal_senders=10",
				"-c",
				"max_replication_slots=10",
			},
			Env: []string{
				fmt.Sprintf("POSTGRES_DB=%s", cfg.Postgres.DB),
				fmt.Sprintf("POSTGRES_USER=%s", c.Postgres.User),
				fmt.Sprintf("POSTGRES_PASSWORD=%s", c.Postgres.Password),
				"POSTGRES_HOST_AUTH_METHOD=scram-sha-256",
				"PGDATA=/var/lib/postgresql/data",
			},
			Healthcheck: postgres_healthcheck,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: cfg.Postgres.Replica.Volume,
					Target: "/var/lib/postgresql/data",
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: app.getNetworks(cfg.Networks.DatabaseNetworkName),
		},
		nil,
		cfg.Postgres.Replica.Name)
	if err != nil {
		return fmt.Errorf("failed to create replica postgres container: %w", err)
	}

	err = app.Docker.Client.ContainerStart(app.Context, response.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start replica postgres container: %w", err)
	}
	cancel := app.spawnLogs(response.ID)
	defer cancel()
	if err := app.waitForContainerHealthWithConfig(response.ID, postgres_healthcheck); err != nil {
		return fmt.Errorf("start of replica postgres failed: %w", err)
	}
	return nil
}

func (c *postgresCredentials) startBouncer(app *AppCtx) error {
	exists, err := app.containerExists(cfg.Postgres.Bouncer.Name)
	if err != nil {
		return fmt.Errorf("failed to check if bouncer container exists: %w", err)
	}
	if exists {
		color.Green("pgbouncer container running")
		return nil
	}
	response, err := app.Docker.Client.ContainerCreate(app.Context,
		&container.Config{
			AttachStdout: true,
			AttachStderr: true,
			Image:        cfg.Postgres.Bouncer.Image,
			Env: []string{
				fmt.Sprintf("DB_HOST=%s", cfg.Postgres.Primary.Name),
				"DB_PORT=5432",
				"AUTH_USER=" + c.Bouncer.User,
				"AUTH_FILE=/etc/pgbouncer/userlist.txt",
				"AUTH_TYPE=scram-sha-256",
				"AUTH_QUERY='SELECT p_user, p_password FROM public.lookup($1)'",
				fmt.Sprintf("LISTEN_PORT=%s", cfg.Postgres.Bouncer.Port),
				"LISTEN_ADDR=0.0.0.0",
				"POOL_MODE=session",
				"MAX_CLIENT_CONN=250",
				"DEFAULT_POOL_SIZE=20",
				"MIN_POOL_SIZE=5",
				"RESERVE_POOL_SIZE=10",
				"SERVER_RESET_QUERY=DISCARD ALL",
				"SERVER_CHECK_QUERY=SELECT 1",
				"SERVER_CHECK_DELAY=30",
				"IGNORE_STARTUP_PARAMETERS=extra_float_digits",
			},
			ExposedPorts: nat.PortSet{
				"6432/tcp": struct{}{},
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("6432/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Postgres.Bouncer.Port}},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: app.getNetworks(cfg.Networks.DatabaseNetworkName),
		},
		nil,
		cfg.Postgres.Bouncer.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create bouncer container: %w", err)
	}
	err = app.Docker.Client.ContainerStart(app.Context, response.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start bouncer container: %w", err)
	}
	save_user_list, err := app.Docker.Client.ContainerExecCreate(app.Context, response.ID, container.ExecOptions{
		Cmd: []string{"/bin/sh", "-c", fmt.Sprintf("echo '%s %s' > /etc/pgbouncer/userlist.txt", c.Bouncer.User, c.Bouncer.Password)},
	})
	if err != nil {
		return fmt.Errorf("failed to create exec command for bouncer: %w", err)
	}
	err = app.Docker.Client.ContainerExecStart(app.Context, save_user_list.ID, container.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start exec command for bouncer: %w", err)
	}
	return nil
}

var postgres_healthcheck = &v1.HealthcheckConfig{
	Test:     []string{"CMD-SHELL", "pg_isready"},
	Interval: time.Second * 10,
	Timeout:  time.Second * 5,
	Retries:  5,
}
