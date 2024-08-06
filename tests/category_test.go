package tests

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var opts = godog.Options{
	Format:      "pretty",
	Paths:       []string{"features"},
	Output:      colors.Colored(os.Stdout),
	Concurrency: 1,
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestFeatures(t *testing.T) {
	o := opts
	o.TestingT = t

	status := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options:             &o,
	}.Run()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}

// Steps
const featureKey CtxKeyType = "feature"

type feature struct {
	URL    string
	Body   string
	Status int
}

var state = NewState[feature](featureKey)

func aCategory(ctx context.Context, categoryType string) (context.Context, error) {
	feat := state.retrieve(ctx)

	if categoryType == "valid" {
		feat.Body = `{"title": "Test Category", "description": "Test Description"}`
	} else {
		feat.Body = `{"title": "Test Category"}`
	}

	return state.enrich(ctx, feat), nil
}

func iCreateTheCategory(ctx context.Context) (context.Context, error) {
	feat := state.retrieve(ctx)

	token, err := generateToken(uuid.NewString(), time.Minute*10)
	if err != nil {
		return ctx, err
	}

	route := fmt.Sprintf("%s/categories", feat.URL)
	req, err := http.NewRequest(http.MethodPost, route, bytes.NewBuffer([]byte(feat.Body)))
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ctx, err
	}

	feat.Status = resp.StatusCode

	return state.enrich(ctx, feat), nil
}

func theCategoryIsCreated(ctx context.Context) (context.Context, error) {
	feat := state.retrieve(ctx)

	if feat.Status != http.StatusCreated {
		return ctx, fmt.Errorf("Expected status code 201, got %d", feat.Status)
	}

	return ctx, nil
}

func theCategoryIsNotCreated(ctx context.Context) (context.Context, error) {
	feat := state.retrieve(ctx)

	if feat.Status != http.StatusUnprocessableEntity {
		return ctx, fmt.Errorf("Expected status code 422, got %d", feat.Status)
	}

	return ctx, nil
}

type testContext struct {
	network    *testcontainers.DockerNetwork
	containers []testcontainers.Container
}

var (
	containers = make(map[string]testContext)
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		network, err := network.New(ctx, network.WithCheckDuplicate(), network.WithDriver("bridge"))
		if err != nil {
			return ctx, err
		}

		localstackContainer, ctx, err := createLocalstackContainer(ctx, network)
		if err != nil {
			return ctx, err
		}

		mongoContainer, ctx, err := createMongoContainer(ctx, network)
		if err != nil {
			return ctx, err
		}

		apiContainer, ctx, err := createApiContainer(ctx, network)
		if err != nil {
			return ctx, err
		}

		containers[sc.Id] = testContext{
			network: network,
			containers: []testcontainers.Container{
				mongoContainer,
				apiContainer,
				localstackContainer,
			},
		}

		return ctx, nil
	})

	ctx.Step(`^a "([^"]*)" category$`, aCategory)
	ctx.Step(`^I create the category$`, iCreateTheCategory)
	ctx.Step(`^the category is created$`, theCategoryIsCreated)
	ctx.Step(`^the category is not created$`, theCategoryIsNotCreated)

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if err != nil {
			return ctx, err
		}

		tc := containers[sc.Id]

		for _, c := range tc.containers {
			err := c.Terminate(ctx)
			if err != nil {
				return ctx, err
			}
		}

		err = tc.network.Remove(ctx)

		return ctx, err
	})
}

func generateToken(userId string, expire time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(expire).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("my-secret"))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", tokenString), nil
}

func createMongoContainer(ctx context.Context, network *testcontainers.DockerNetwork) (testcontainers.Container, context.Context, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "mongo:7",
			ExposedPorts: []string{
				"27017",
			},
			Env: map[string]string{
				"MONGO_INITDB_ROOT_USERNAME": "product",
				"MONGO_INITDB_ROOT_PASSWORD": "product",
			},
			Networks: []string{
				network.Name,
			},
			NetworkAliases: map[string][]string{
				network.Name: {"mongo"},
			},
			WaitingFor: wait.ForLog("Waiting for connections"),
		},
		Started: true,
	})
	if err != nil {
		return nil, ctx, err
	}

	mongoIp, err := container.Host(ctx)
	if err != nil {
		return nil, ctx, err
	}

	mongoPort, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return nil, ctx, err
	}

	connStr := fmt.Sprintf("mongodb://product:product@%s:%s/", mongoIp, mongoPort.Port())

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		return nil, ctx, err
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		return nil, ctx, err
	}

	return container, ctx, nil
}

func createApiContainer(ctx context.Context, network *testcontainers.DockerNetwork) (testcontainers.Container, context.Context, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    "../",
				Dockerfile: "Dockerfile",
				KeepImage:  true,
			},
			ExposedPorts: []string{
				"5000",
			},
			Env: map[string]string{
				"API_PORT":              "5000",
				"API_ENV_NAME":          "development",
				"API_VERSION":           "v1",
				"DB_NAME":               "product_db",
				"DB_URL":                "todo",
				"DB_URL_SECRET_NAME":    "db-secret-url",
				"AWS_ACCESS_KEY_ID":     "test",
				"AWS_SECRET_ACCESS_KEY": "test",
				"AWS_SESSION_TOKEN":     "test",
				"AWS_BASE_ENDPOINT":     "http://test:4566",
			},
			Networks: []string{
				network.Name,
			},
			WaitingFor: wait.ForLog("Server started").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})

	if err != nil {
		return nil, ctx, err
	}

	ports, err := container.Ports(ctx)
	if err != nil {
		return nil, ctx, err
	}

	if len(ports["5000/tcp"]) == 0 {
		return nil, ctx, fmt.Errorf("Port 5000/tcp not found")
	}

	port := ports["5000/tcp"][0].HostPort

	// Run health check
	res, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
	if err != nil {
		return nil, ctx, err
	}

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, ctx, err
		}
		defer res.Body.Close()

		fmt.Printf("Body: %s", string(body))

		return nil, ctx, fmt.Errorf("API health check failed with status: %d", res.StatusCode)
	}

	return container, state.enrich(ctx, &feature{
		URL: fmt.Sprintf("http://localhost:%s/api/v1", port),
	}), nil
}

func createLocalstackContainer(ctx context.Context, network *testcontainers.DockerNetwork) (testcontainers.Container, context.Context, error) {
	secretsScript, err := filepath.Abs(filepath.Join(".", "testdata", "init-sm.sh"))
	if err != nil {
		return nil, ctx, err
	}

	secretsScriptReader, err := os.Open(secretsScript)
	if err != nil {
		return nil, ctx, err
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "localstack/localstack:latest",
			ExposedPorts: []string{
				"4566",
			},
			Env: map[string]string{
				"SERVICES":       "secretsmanager",
				"DEFAULT_REGION": "us-east-1",
				"DOCKER_HOST":    "unix:///var/run/docker.sock",
			},
			Networks: []string{
				network.Name,
			},
			NetworkAliases: map[string][]string{
				network.Name: {
					"test",
				},
			},
			Files: []testcontainers.ContainerFile{
				{
					Reader:            secretsScriptReader,
					ContainerFilePath: "/etc/localstack/init/ready.d/init-sns.sh",
					FileMode:          0777,
				},
			},
			WaitingFor: wait.ForListeningPort("4566/tcp").WithStartupTimeout(120 * time.Second),
		},
		Started: true,
	})

	if err != nil {
		return nil, ctx, err
	}

	return container, ctx, nil
}
