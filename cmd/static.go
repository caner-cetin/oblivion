package cmd

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/fatih/color"
	v1 "github.com/moby/docker-image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed config/static/*
var staticBuildFiles embed.FS

var (
	staticChmodCmd = &cobra.Command{
		Use: "chmod",
		Run: WrapCommandWithResources(staticChmod, ResourceConfig{Resources: []ResourceType{}}),
	}
	staticUpCmd = &cobra.Command{
		Use: "up",
		Run: WrapCommandWithResources(staticUp, ResourceConfig{Resources: []ResourceType{ResourceDocker}}),
	}
	staticCmd = &cobra.Command{
		Use: "static",
	}
)

func getStaticCmd() *cobra.Command {
	staticCmd.AddCommand(staticChmodCmd)
	staticCmd.AddCommand(staticUpCmd)
	return staticCmd
}

func staticUp(cmd *cobra.Command, args []string) {
	app := GetApp(cmd)
	app.Spinner.Prefix = "building static"
	app.Spinner.Start()
	app.Spinner.Prefix = "checking for nginx image"
	defer app.Spinner.Stop()
	exists, err := app.imageExists("cansu.dev-static-nginx")
	if err != nil {
		log.Error().Err(err).Msg("failed to check if volume exists")
		return
	}
	if !exists {
		app.Spinner.Prefix = "building image..."
		if err := app.buildImage(staticBuildFiles, "config/static", "cansu.dev-static-nginx", "nginx.Dockerfile"); err != nil {
			log.Error().Err(err).Send()
			return
		}
	}
	app.Spinner.Prefix = "creating static container"
	resp, err := app.Docker.Client.ContainerCreate(app.Context,
		&container.Config{
			Image:        "cansu.dev-static-nginx",
			AttachStdout: true,
			AttachStderr: true,
			AttachStdin:  false,
			OpenStdin:    false,
			Healthcheck:  nginx_healthcheck,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("80/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "44444"}},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/var/www/servers/cansu.dev/static",
					Target: "/var/www/servers/cansu.dev/static/",
				},
			},
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
		},
		nil,
		nil,
		"file-server",
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create static container")
		return
	}
	app.Spinner.Prefix = "starting container"
	if err := app.Docker.Client.ContainerStart(app.Context, resp.ID, container.StartOptions{}); err != nil {
		log.Error().Err(err).Send()
		return
	}
	if err := app.waitForContainerHealthWithConfig(resp.ID, nginx_healthcheck); err != nil {
		log.Error().Err(err).Send()
		return
	}

}

var nginx_healthcheck = &v1.HealthcheckConfig{
	Test:     []string{"CMD-SHELL", "wget -O /dev/null http://localhost || exit 1"},
	Interval: 5 * time.Second,
	Timeout:  10 * time.Second,
	Retries:  10,
}

func staticChmod(cmd *cobra.Command, args []string) {
	path := "/var/www/servers/cansu.dev/static"
	uploader, err := user.Lookup("caner")
	if err != nil {
		log.Error().Err(err).Msgf("failed to get user %s", "caner")
		return
	}
	uploader_gid, _ := strconv.Atoi(uploader.Gid)

	// root:uploader_group
	if err := os.Chown(path, 0, uploader_gid); err != nil {
		log.Error().Err(err).Msg("failed to set ownership")
		return
	}

	// 2775:
	// 2000 - setgid bit
	// 7 - full permissions for owner
	// 7 - full permissions for group
	// 5 - read and execute for others (nginx)
	if err := os.Chmod(path, 2775); err != nil {
		log.Error().Err(err).Msg("failed to set permissions")
		return
	}

	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := os.Chown(path, 0, uploader_gid); err != nil {
			return fmt.Errorf("chown failed for %s: %w", path, err)
		}
		// for directories, same permissions as 2775 but without setgid
		if info.IsDir() {
			return os.Chmod(path, 0775)
		}
		// for files, rw for owner, rw for group, wx for anyone else
		return os.Chmod(path, 0664)
	}); err != nil {
		log.Error().Err(err).Msg("failed to set recursive permissions")
		return
	}

	acl_cmd := exec.Command("sudo", "setfacl", "-R", "-d", "-m", "u::rwx,g::rwx,o::r-x", path)
	acl_cmd.Stdout = os.Stdout
	acl_cmd.Stderr = os.Stderr
	if err := acl_cmd.Run(); err != nil {
		log.Error().Err(err).Msg("failed to set ACL list")
		return
	}

	color.Green("Permissions and ownership set for %s", path)
}
