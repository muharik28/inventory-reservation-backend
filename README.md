## Prerequisite

- Install [go](https://golang.org/doc/install) in local machine for convenience development experience (auto complete, code sugestion, etc)
- Install golang plugin to your editor choice (ie. VSCode, Atom, Goland, Intellij IDE)
- Install [PostgresSQL](https://www.postgresql.org/download/)
- [Docker](https://docs.docker.com/install/) and [docker-compose](https://docs.docker.com/compose/)
- [swag](https://github.com/swaggo/swag) CLI: `go install github.com/swaggo/swag/cmd/swag@latest` (for API documentation)

## Install make

### You first need Makefile

```sh
# install windows
C:\> choco install make
```

```sh
# install linux
$ sudo apt-get install make
```

```sh
# install MacOS with Homebrew
$ brew install make
```

### Verify Version Makefile

```sh
# verify that Makefile is installed correctly by checking the version
$ make --version
```

## How to run

Despite, it is possible to run this project in local machine Please follow this steps:
- Changes file .env to your config `.cp .env.example to .env`.

```sh
# run rest api server
$ make rest

# run scheduler (optional)
$ make sch

# General help
$ make help

# Specific help for sub-commands
$ make help-rest

$ make help-sch

# Using --help flag
$ make help-flag

# Example for PowerShell
$ go run main.go completion powershell | Out-String | Invoke-Expression

# See other shell options
$ make help-completion

# 1. Build binary
$ make app-build

# 2. Run commands
$ .\inventory.exe rest

$ .\inventory.exe scheduler
```

## How to run unit test

Despite, it is possible to run unit test this project in local machine Please follow this steps:

> **Note:** the `test/` suite contains unit tests **and** PostgreSQL-backed integration/concurrency tests.
> Integration tests create a throwaway `inventory_test_db` database, so PostgreSQL must be reachable
> (defaults match Docker Compose: `localhost:5432`, `postgres/postgres`). Start it first with:
> `$ make compose-db`
> Without PostgreSQL, the integration tests are skipped automatically.

```sh
# Run all tests with the race detector (required by the Definition of Done)
$ make test-all

# Run unit test on folder test
$ make test-unit

# Run all tests in the project
$ make test-full

# Run specific test
$ make test-specific

# Run concurrency / stress test
$ make test-concurrency

# Run all tests without cache (re-run fresh)
$ make test-cache

# Run all tests with cache
$ make test-all-cache
```

## How to run docker compose

Despite, it is possible to run docker compose this project in local machine Please follow this steps:
```sh
# Build image and run all containers in the background
$ make compose

# Only run postgres database container
$ make compose-db

# Restart postgres database container
$ make compose-db-restart

# View Container Status
$ make compose-ps

# View all logs in realtime
$ make compose-logs

# View only REST API logs
$ make compose-log-app

# View only Scheduler logs
$ make compose-log-sch

# Stop container
$ make compose-down

# Stop container + reset/delete volume data database
$ make compose-down-volume

# Even if an error still occurs in `go mod download`
# Notice the error message that appears before exit code: 1
$ make compose-cache-app

$ make compose-cache-sch
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
$ make swag
```

# Project-Structure


    inventory-reservation-backend/
    ├── README.md                # Project description
    ├── go.mod                   # Go module
    ├── go.sum                   # Dependency list
    ├── Makefile                 # Automating the build process (compilation and integration) of program code
    ├── .env.example             # Environment variables example
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

