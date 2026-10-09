## Prerequisite

- Install [go](https://golang.org/doc/install) in local machine for convenience development experience (auto complete, code sugestion, etc)
- Install golang plugin to your editor choice (ie. VSCode, Atom, Goland, Intellij IDE)
- Install [PostgresSQL](https://www.postgresql.org/download/)
- [Docker](https://docs.docker.com/install/) and [docker-compose](https://docs.docker.com/compose/)

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
    │   ├── api/                 # API documentation
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
        └── 20210101_init.sql    # Initial migration script

