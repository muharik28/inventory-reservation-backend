## Prerequisite

- Install [go](https://golang.org/doc/install) in local machine for convenience development experience (auto complete, code sugestion, etc)
- Install golang plugin to your editor choice (ie. VSCode, Atom, Goland, Intellij IDE)
- Install [PostgresSQL](https://www.postgresql.org/download/)
- [Docker](https://docs.docker.com/install/) and [docker-compose](https://docs.docker.com/compose/)
- [swag](https://github.com/swaggo/swag) CLI: `go install github.com/swaggo/swag/cmd/swag@latest` (for API documentation)

## How to run

Despite, it is possible to run this project in local machine Please follow this steps:
- Changes file .env to your config `.cp .env.example to .env`.

```sh
# run rest api server
$ go run main.go rest

# run scheduler (optional)
$ go run main.go scheduler

# General help
$ go run main.go help

# Specific help for sub-commands
$ go run main.go help rest

$ go run main.go help scheduler

# Using --help flag
$ go run main.go rest --help

# Example for PowerShell
$ go run main.go completion powershell | Out-String | Invoke-Expression

# See other shell options
$ go run main.go completion --help

# 1. Build binary
$ go build -o inventory.exe main.go

# 2. Run commands
$ .\inventory.exe rest

$ .\inventory.exe scheduler
```

## How to run unit test

Despite, it is possible to run unit test this project in local machine Please follow this steps:

> **Note:** the `test/` suite contains unit tests **and** PostgreSQL-backed integration/concurrency tests.
> Integration tests create a throwaway `inventory_test_db` database, so PostgreSQL must be reachable
> (defaults match Docker Compose: `localhost:5432`, `postgres/postgres`). Start it first with:
> `$ docker compose -f deployments/docker-compose.yml up -d postgres`
> Without PostgreSQL, the integration tests are skipped automatically.

```sh
# Run all tests with the race detector (required by the Definition of Done)
$ go test -race -v ./...

# Run unit test on folder test
$ go test -v ./test/...

# Run all tests in the project
$ go test -v ./...

# Run specific test
$ go test -v ./test -run TestDomainLogic

# Run concurrency / stress test
$ go test -race -v ./test -run TestConcurrentReservationsDoNotOversell

# Run all tests without cache (re-run fresh)
$ go test -count=1 -race -v ./test/...
```

## How to run docker compose

Despite, it is possible to run docker compose this project in local machine Please follow this steps:
```sh
# Build image and run all containers in the background
$ docker compose -f deployments/docker-compose.yml up -d --build

# Only run postgres database container
$ docker compose -f deployments/docker-compose.yml up -d postgres

# Restart postgres database container
$ docker compose -f deployments/docker-compose.yml restart postgres

# View Container Status
$ docker compose -f deployments/docker-compose.yml ps

# View all logs in realtime
$ docker compose -f deployments/docker-compose.yml logs -f

# View only REST API logs
$ docker compose -f deployments/docker-compose.yml logs -f app

# View only Scheduler logs
$ docker compose -f deployments/docker-compose.yml logs -f scheduler

# Stop container
$ docker compose -f deployments/docker-compose.yml down

# Stop container + reset/delete volume data database
$ docker compose -f deployments/docker-compose.yml down -v

# Even if an error still occurs in `go mod download`
# Notice the error message that appears before exit code: 1
$ docker build --no-cache --progress=plain -t app .

$ docker build --no-cache --progress=plain -t scheduler .
```

## API Documentation (Swagger)

API documentation is generated using **Swagger / OpenAPI** and served interactively via Swagger UI.

### Access Swagger UI
When the REST server is running (`go run main.go rest`), open:
- **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- **OpenAPI JSON**: [http://localhost:8080/swagger/doc.json](http://localhost:8080/swagger/doc.json)

### Regenerate Swagger Documentation
Whenever routes, handler annotations, or request/response DTOs are updated, regenerate the documentation files to `docs/api`:
```sh
# Generate Swagger docs into docs/api/
$ swag init -g cmd/rest/rest.go -o docs/api
```

# Project-Structure


    inventory-reservation-backend/
    ├── README.md                # Project description
    ├── go.mod                   # Go module
    ├── go.sum                   # Dependency list
    ├── api/                     # API definitions
    │   ├── rest/                # REST API definitions
    │   │   ├── api/             # REST API implementation
    │   │   │   ├── routes/      # Routes for REST API
    │   │   │   └── server.go    # Server setup for REST API
    ├── build/                   # Build-related files
    │   ├── package/             # Package-related files
    │   |   └── Dockerfile       # Dockerfile for building the application
    ├── cmd/                     # Contains main applications
    │   ├── rest/                # REST application
    │   │   └── rest.go          # Entry point Cobra Command for REST app
    │   │   └── scheduler/       # Scheduler for tasks
    │   │       └── scheduler.go # Entry point Cobra Command for scheduler
    │   ├── root.go              # Entry point for the root Cobra Command
    ├── configs/                 # Configuration files
    │   ├── configs.go           # Configuration Global
    ├── deployments/             # Deployment-related files
    │   ├── docker-compose.yml   # Docker Compose configuration file
    ├── docs/                    # Project documentation
    │   ├── api/                 # API documentation (Swagger generated docs)
    │   ├── architecture/        # Architecture documentation
    │   ├── images/              # Images related to documentation
    ├── internal/                # Code that can only be used by this project
    │   ├── application/         # Application logic
    │   │   ├── adapters/        # Adapters for external services
    │   │   │   └── info/        # Info related adapters
    │   │   │       └── info.go  # Info adapter
    │   │   ├── handlers/        # Handlers for handling requests
    │   │   │   └── info/        # Info related handlers
    │   │   │       ├── interface.go # Info handlers interface
    │   │   │       └── impl/    # Implementation handlers
    │   │   │           └── info.go # Info handler (Adapter (Hexagonal))
    │   │   ├── presenters/      # Presenters for formatting responses
    │   │   │   ├── request/     # Request related presenters
    │   │   │   │   └── FindInfoDetailByIDRequest.go # Find info detail request
    │   │   │   └── response/    # Response related presenters
    │   │   │       └── InfoDetailResponse.go # Info detail response
    │   │   ├── repositories/    # Repositories for data access
    │   │   │   └── info/        # Info related repositories
    │   │   │       ├── interface.go # Info repository interface
    │   │   │       └── impl/    # Implementation repository
    │   │   │           └── model_info.go  # ORM (Object-Relational Mapping) model
    │   │   │           └── repo_info.go   # Info repository (Adapter (Hexagonal))
    │   │   │           └── mapper_info.go # Info mapping from domain entities
    │   │   ├── usecases/        # Use cases for application logic
    │   │   │   └── info/        # Info related use cases
    │   │   │       ├── interface.go # Info use cases interface
    │   │   │       └── impl/    # Implementation use cases
    │   │   │           └── info.go # Info use case (Adapter (Hexagonal))
    │   ├── domain/              # Domain models
    │   │   ├── entities/        # Domain entities
    │   │   |   ├── response.go  # Response standardization entities
    │   │   │   └── info/        # Info related entities group
    │   │   │       └── info.go  # Info entity definition
    │   ├── pkg/                 # Internal packages
    │   │   ├── database/        # Database related code
    │   │   │   ├── nosql/       # NoSQL database connection code
    │   │   │   └── sql/         # SQL database connection code
    │   │   ├── logger/          # Logging utilities
    │   │   └── utils/           # Utility functions
    ├── pkg/                     # Code that can be used by other projects
    │   ├── pkg1/                # Package global for used all application
    ├── scripts/                 # Scripts for development and deployment
    │   ├── setup.sh             # Setup script
    ├── test/                    # Contains test files
    │   ├── testdata/            # Test data
    │   ├── integration_test.go  # Integration test
    └── migrations/              # Database migration scripts
        └── 20261010_init.sql    # Initial migration script

