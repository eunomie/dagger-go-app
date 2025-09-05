// A module to build, test and run Dagger Go App

package main

import (
	"dagger/dagger-go-app/internal/dagger"
)

type DaggerGoApp struct{}

// Create the image containing the application
func (m *DaggerGoApp) Image(
// Source directory of the application
// +optional
// +defaultPath="/"
	src *dagger.Directory,
) *dagger.Container {
	webbuild := dag.Container().
		From("node:24-alpine3.22").
		WithWorkdir("/app/web").
		WithDirectory(
			".",
			src.Directory("web"),
			dagger.ContainerWithDirectoryOpts{
				Include: []string{
					"package.json",
					"package-lock.json",
				},
			}).
		WithExec([]string{"npm", "install", "--no-audit", "--no-fund"}).
		WithDirectory(".", src.Directory("web")).
		WithExec([]string{"npm", "run", "build"})

	gobuild := dag.Container().
		From("golang:1.25-alpine3.22").
		WithWorkdir("/app").
		WithDirectory(
			".",
			src,
			dagger.ContainerWithDirectoryOpts{
				Include: []string{
					"go.mod",
					"go.sum",
				},
			}).
		WithExec([]string{"go", "mod", "download"}).
		WithDirectory(".", src.WithoutDirectory("web")).
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-ldflags", "-s -w", "-o", "server", "./main.go"})

	runtime := dag.Container().
		From("alpine:3.22").
		WithExec([]string{"apk", "add", "--no-cache", "ca-certificates"}).
		WithWorkdir("/app").
		// copy Backend binary
		WithFile("/app/server", gobuild.File("server")).
		// copy Frontend dist
		WithDirectory("/app/web/dist", webbuild.Directory("/app/web/dist")).
		WithEnvVariable("ADDR", ":8080").
		WithExposedPort(8080).
		WithDefaultArgs([]string{"/app/server"})

	return runtime
}

// Create a database container
func (m *DaggerGoApp) DB() *dagger.Container {
	return dag.Container().
		From("postgres:17-alpine3.22").
		WithEnvVariable("POSTGRES_USER", "app").
		WithEnvVariable("POSTGRES_PASSWORD", "app").
		WithEnvVariable("POSTGRES_DB", "appdb").
		WithMountedCache("/var/lib/postgresql/data", dag.CacheVolume("db-data")).
		WithExposedPort(5432)
}

// Return the app as a service, connected to a database
func (m *DaggerGoApp) Service(
// Source directory of the application
// +optional
// +defaultPath="/"
	src *dagger.Directory,
) *dagger.Service {
	return m.Image(src).
		WithServiceBinding("db", m.DB().AsService(dagger.ContainerAsServiceOpts{UseEntrypoint: true})).
		WithEnvVariable("DATABASE_URL", "postgres://app:app@db:5432/appdb?sslmode=disable").
		AsService()
}
