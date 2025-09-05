// A module to build, test and run Dagger Go App

package main

import (
	"context"
	"fmt"

	"dagger/dagger-go-app/internal/dagger"
)

type DaggerGoApp struct {
	Src *dagger.Directory
}

// Creates a new DaggerGoApp instance
func New(
// Source directory of the application
// +optional
// +defaultPath="/"
	src *dagger.Directory,
) *DaggerGoApp {
	return &DaggerGoApp{
		Src: src,
	}
}

// Environment for the frontend part of the application
func (m *DaggerGoApp) FrontendEnv() *dagger.Container {
	return dag.Container().
		From("node:24-alpine3.22").
		WithWorkdir("/app/web").
		WithDirectory(
			".",
			m.Src.Directory("web"),
			dagger.ContainerWithDirectoryOpts{
				Include: []string{
					"package.json",
					"package-lock.json",
				},
			}).
		WithMountedCache("/root/.npm", dag.CacheVolume("npm-cache")).
		WithExec([]string{"npm", "install", "--no-audit", "--no-fund"}).
		WithDirectory(".", m.Src.Directory("web"))
}

// Build frontend application
func (m *DaggerGoApp) FrontentBuild() *dagger.Container {
	return m.FrontendEnv().
		WithExec([]string{"npm", "run", "build"})
}

// Distribution of the frontend
func (m *DaggerGoApp) FrontendDist() *dagger.Directory {
	return m.FrontentBuild().Directory("/app/web/dist")
}

// Run frontend tests
func (m *DaggerGoApp) FrontendTest(ctx context.Context) (string, error) {
	ctr := m.FrontendEnv().
		WithExec([]string{"npm", "ci"}).
		WithExec([]string{"npm", "run", "test:run"}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny})
	out, err := ctr.CombinedOutput(ctx)
	if err != nil {
		return "", err
	}
	if e, err := ctr.ExitCode(ctx); err != nil {
		return "", err
	} else if e != 0 {
		return "", fmt.Errorf("frontend tests failed:\n%s", out)
	}
	return out, nil
}

// Environment for the backend part of the application
func (m *DaggerGoApp) BackendEnv() *dagger.Container {
	return dag.Container().
		From("golang:1.25-alpine3.22").
		WithWorkdir("/app").
		WithDirectory(
			".",
			m.Src,
			dagger.ContainerWithDirectoryOpts{
				Include: []string{
					"go.mod",
					"go.sum",
				},
			}).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithExec([]string{"go", "mod", "download"}).
		WithDirectory(".", m.Src.WithoutDirectory("web")).
		WithEnvVariable("CGO_ENABLED", "0")
}

// Build backend application
func (m *DaggerGoApp) BackendBuild() *dagger.Container {
	return m.BackendEnv().
		WithExec([]string{"go", "build", "-ldflags", "-s -w", "-o", "server", "./main.go"})
}

// Distribution of the backend
func (m *DaggerGoApp) BackendDist() *dagger.File {
	return m.BackendBuild().File("server")
}

// Run backend tests
func (m *DaggerGoApp) BackendTest(ctx context.Context) (string, error) {
	ctr := m.BackendEnv().
		WithExec([]string{"go", "test", "./..."}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny})
	out, err := ctr.CombinedOutput(ctx)
	if err != nil {
		return "", err
	}
	if e, err := ctr.ExitCode(ctx); err != nil {
		return "", err
	} else if e != 0 {
		return "", fmt.Errorf("backend tests failed:\n%s", out)
	}
	return out, nil
}

// Run tests
func (m *DaggerGoApp) Test(ctx context.Context) (string, error) {
	fout, err := m.FrontendTest(ctx)
	if err != nil {
		return "", err
	}
	bout, err := m.BackendTest(ctx)
	if err != nil {
		return "", err
	}
	return fout + "\n" + bout, nil
}

// Create the image containing the application
func (m *DaggerGoApp) Image() *dagger.Container {
	runtime := dag.Container().
		From("alpine:3.22").
		WithExec([]string{"apk", "add", "--no-cache", "ca-certificates"}).
		WithWorkdir("/app").
		// copy Backend binary
		WithFile("/app/server", m.BackendDist()).
		// copy Frontend dist
		WithDirectory("/app/web/dist", m.FrontendDist()).
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
func (m *DaggerGoApp) Service() *dagger.Service {
	return m.Image().
		WithServiceBinding("db", m.DB().AsService(dagger.ContainerAsServiceOpts{UseEntrypoint: true})).
		WithEnvVariable("DATABASE_URL", "postgres://app:app@db:5432/appdb?sslmode=disable").
		AsService()
}
