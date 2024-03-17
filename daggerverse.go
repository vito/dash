package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/exp/slog"
)

type Daggerverse struct{}

const (
	GoVersion       = "1.21"
	DaggerVersion   = "0.9.3"
	PostgresVersion = "15.3"
)

func (m *Daggerverse) Source() *Directory {
	// FIXME: remove when Host.Directory supports ..
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return dag.Host().Directory(filepath.Dir(wd))
}

func (m *Daggerverse) Bin() *File {
	return dag.Go().FromVersion(GoVersion).
		WithSource(m.Source()).
		WithEnvVariable("CGO_ENABLED", "0").
		Exec([]string{"go", "build", "-o", "/bin/daggerverse", "."}).
		File("/bin/daggerverse")
}

func (m *Daggerverse) RiverBin() *File {
	return dag.Go().FromVersion(GoVersion).
		WithSource(m.Source()).
		WithEnvVariable("CGO_ENABLED", "0").
		// use version from go.mod
		Exec([]string{"go", "build", "-o", "/bin/river", "github.com/riverqueue/river/cmd/river"}).
		File("/bin/river")
}

func (m *Daggerverse) GoremanBin() *File {
	return dag.Go().FromVersion(GoVersion).
		WithSource(m.Source()).
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GOBIN", "/bin").
		Exec([]string{"go", "install", "github.com/mattn/goreman@v0.3.15"}).
		File("/bin/goreman")
}

func (m *Daggerverse) App() *Service {
	return m.AppContainer().
		WithServiceBinding("db", m.Database()).
		WithEnvVariable("DATABASE_URL", "postgres://dagger:dagger@db/daggerverse?sslmode=disable").
		// just use local workers for dev since this uses nesting anyway
		WithEnvVariable("WORKERS", "3").
		WithMountedFile("/tmp/refs", dag.Host().File("refs")).
		WithFocus().
		WithExec([]string{"load-modules", "/tmp/refs"}, ContainerWithExecOpts{
			ExperimentalPrivilegedNesting: true,
		}).
		WithExec(nil, ContainerWithExecOpts{
			ExperimentalPrivilegedNesting: true,
		}).
		AsService()
}

func (m *Daggerverse) AppContainer() *Container {
	return dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/daggerverse", m.Bin()).
		WithFile("/usr/local/bin/river", m.RiverBin()).
		WithEnvVariable("PORT", "8080").
		WithEnvVariable("MIGRATE", "true").
		WithExposedPort(8080).
		WithDefaultArgs().
		WithEntrypoint([]string{"daggerverse"})
}

func (m *Daggerverse) FlyWorker() *Container {
	return dag.Container().
		From(fmt.Sprintf("registry.dagger.io/engine:v%s", DaggerVersion)).
		WithEntrypoint(nil).
		WithEnvVariable("BIN_DIR", "/usr/local/bin").
		WithEnvVariable("DAGGER_VERSION", DaggerVersion).
		WithFile("/opt/dagger-install.sh", dag.HTTP("https://dl.dagger.io/dagger/install.sh")).
		WithExec([]string{"sh", "/opt/dagger-install.sh"}).
		WithEnvVariable("_EXPERIMENTAL_DAGGER_RUNNER_HOST", "unix:///var/run/buildkit/buildkitd.sock").
		WithEnvVariable("_EXPERIMENTAL_DAGGER_CLI_BIN", "/usr/local/bin/dagger").
		WithFile("/usr/local/bin/daggerverse", m.Bin()).
		WithFile("/usr/local/bin/river", m.RiverBin()).
		WithFile("/usr/local/bin/goreman", m.GoremanBin()).
		WithWorkdir("/app").
		WithFile("/app/Procfile", m.Source().File("workers/Procfile")).
		WithEnvVariable("PORT", "8080").
		WithEnvVariable("MIGRATE", "true").
		WithExposedPort(8080).
		WithDefaultArgs().
		WithEntrypoint([]string{"goreman", "--set-ports=false", "start"})
}

func (m *Daggerverse) Database() *Service {
	return dag.Postgres().
		WithVersion(PostgresVersion).
		WithCredential(
			dag.SetSecret("pg_user", "dagger"),
			dag.SetSecret("pg_password", "dagger"),
		).
		WithDatabaseName("daggerverse").
		WithInitScript("initdb.sql", dag.Host().File("initdb.sql")).
		Database().
		AsService()
}

const WorkerVMScale = "performance-4x"

var nonAlnum = regexp.MustCompile("[^a-zA-Z0-9]+")

func (m *Daggerverse) Deploy(
	ctx context.Context,
	token *Secret,
	tag Optional[string],
	webAppName Optional[string],
	workersAppName Optional[string],
	workersMachineSize Optional[string],
	dbAppName Optional[string],
	dbForkFrom Optional[string],
	dbName Optional[string],
	dbMachineSize Optional[string],
	dbClusterSize Optional[int],
	dbVolumeSize Optional[int],
	orgName Optional[string],
	primaryRegion Optional[string],
	// TODO support again
	// secondaryRegion Optional[string],
) (string, error) {
	orgName_ := orgName.GetOr("dagger")
	primaryRegion_ := primaryRegion.GetOr("ord")
	// secondaryRegion_ := secondaryRegion.GetOr("iad")
	workersVMSize_ := workersMachineSize.GetOr("performance-4x")

	tag_ := tag.GetOr(time.Now().Format("20060102"))
	webAppName_ := webAppName.GetOr(fmt.Sprintf("daggerverse-%s", tag_))
	workersAppName_ := workersAppName.GetOr(fmt.Sprintf("daggerverse-workers-%s", tag_))
	dbAppName_ := dbAppName.GetOr(fmt.Sprintf("daggerverse-postgres-%s", tag_))
	dbName_ := dbName.GetOr("daggerverse")
	dbVMSize_ := dbMachineSize.GetOr("shared-cpu-1x")
	dbClusterSize_ := dbClusterSize.GetOr(1)
	dbVolumeSize_ := dbVolumeSize.GetOr(30)

	apps, err := flyAppsList(ctx, token)
	if err != nil {
		return "", err
	}

	for _, appName := range []string{webAppName_, workersAppName_} {
		if _, ok := apps[appName]; !ok {
			if err := flyAppsCreate(ctx, token, appName, orgName_); err != nil {
				return "", fmt.Errorf("create app %s: %w", webAppName_, err)
			}
		}
	}

	if _, ok := apps[dbAppName_]; !ok {
		pgArgs := []string{
			"--initial-cluster-size", strconv.Itoa(dbClusterSize_),
			"--vm-size", dbVMSize_,
			"--volume-size", strconv.Itoa(dbVolumeSize_),
		}
		if fork, ok := dbForkFrom.Get(); ok {
			pgArgs = append(pgArgs, "--fork-from", fork)
		}
		if err := flyPostgresCreate(ctx, token, dbAppName_, orgName_, primaryRegion_, pgArgs...); err != nil {
			return "", fmt.Errorf("create postgres %s: %w", dbAppName_, err)
		}
		if err := flyPostgresAttach(ctx, token, dbAppName_, webAppName_, dbName_); err != nil {
			return "", fmt.Errorf("attach postgres %s to %s: %w", dbAppName_, webAppName_, err)
		}
		if err := flyPostgresAttach(ctx, token, dbAppName_, workersAppName_, dbName_); err != nil {
			return "", fmt.Errorf("attach postgres %s to %s: %w", dbAppName_, workersAppName_, err)
		}
	}

	if err := m.DeployWeb(ctx, tag_, token); err != nil {
		return "", fmt.Errorf("deploy web: %w", err)
	}

	if err := m.DeployWorkers(ctx, tag_, token, workersVMSize_); err != nil {
		return "", fmt.Errorf("deploy workers: %w", err)
	}

	return fmt.Sprintf("https://%s.fly.dev", webAppName_), nil
}

func (m *Daggerverse) Undeploy(
	ctx context.Context,
	token *Secret,
	tag Optional[string],
	webAppName Optional[string],
	workersAppName Optional[string],
	dbAppName Optional[string],
	orgName Optional[string],
) error {
	tag_ := tag.GetOr(time.Now().Format("20060102"))
	webAppName_ := webAppName.GetOr(fmt.Sprintf("daggerverse-%s", tag_))
	workersAppName_ := workersAppName.GetOr(fmt.Sprintf("daggerverse-workers-%s", tag_))
	dbAppName_ := dbAppName.GetOr(fmt.Sprintf("daggerverse-postgres-%s", tag_))

	apps, err := flyAppsList(ctx, token)
	if err != nil {
		return err
	}

	for _, app := range []string{webAppName_, workersAppName_, dbAppName_} {
		if _, ok := apps[app]; !ok {
			slog.Info("app not found", "app", app)
			continue
		}
		if err := flyAppsDestroy(ctx, token, app, orgName.GetOr("dagger")); err != nil {
			return fmt.Errorf("destroy app %s: %w", app, err)
		}
	}

	return nil
}

func (m *Daggerverse) DeployWeb(ctx context.Context, tag string, token *Secret) error {
	return publishAndDeploy(ctx, tag, token,
		m.Source(),
		m.AppContainer(),
	)
}

func (m *Daggerverse) DeployWorkers(ctx context.Context, tag string, token *Secret, vmSize string) error {
	return publishAndDeploy(ctx, tag, token,
		m.Source().Directory("workers"),
		m.FlyWorker(),
		"--vm-size", vmSize,
	)
}
