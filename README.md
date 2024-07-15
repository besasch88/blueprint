
![Cover image](/assets/images/cover.png)

# Blueprint application!
Welcome to the GO + GIN blueprint application to get all your projects started! Fork this repository and start building on it following the best practices already implemented and avoiding boilerplates like Postgres DB connection, logger, project structure, env variables, and scripts.

## Developer Experience
Below you can find instructions on how to start developing natively your project based on the Blueprint, leveraging a dockerized external Database.

#### Install GO
First of all, let's install go version `1.22.4` or higher from this link: https://go.dev/doc/install

### Check the Go version
Before proceed, ensure your version is correct. Run this command in your terminal:
``` sh
go version
```
The answer should be something like this according to your installed version and arch:
``` sh
go version go1.22.4 darwin/amd64
```

### Start PostgreSql
Navigate in the `build` folder and start the Postgres DB inside Docker:
``` sh
cd build
docker compose up bp-database -d
```
It contains only a PostgresQL database server mapped on the local port `54322`. Feel free to take a look to the docker-compose file to retrieve DB credentials if you want to use an external tool to connect with.

### Migration Tool
The Migration Tool is a command that help you in creating migrations, apply or revert thanks to migration versioning. Let's start by installing the migration tool:
``` sh
brew install golang-migrate
```
and with the following command you can create your first migration:
``` sh
migrate create -ext sql -dir ./scripts/migrations -seq init schema
```
Thanks to it, the tool will create two empty sql files in the `scripts/migrations` folder to apply a new changes to the Database or to revert it.
Once your migrations are defined, you can apply them locally with this command:
``` sh
migrate -path "./scripts/migrations" -database "postgres://blueprint:blueprint@127.0.0.1:54322/blueprint?sslmode=disable" up
```
or simply running this file as a shortcut:
``` sh
./scripts/migrate_local.sh
```
Looking to the docker-compose file, you will notice that there is a dedicated service aims to apply migrations each time the project is deployed in your production environment. Basically it starts, applies all the migrations and shutdown.

### Start the webapp locally
Now we have all the migration setup, the DB running and updated and we can run your local webapp locally via this command:
``` sh
go run cmd/webapp/main.go
```
If everything is fine, you will see in logs that the webapp is up and running, waiting incoming API requests.

### Start dockerized application
If you want. you can run the webapp application in docker, useful for testing/demo purposes. So from the root folder of your project run:
``` sh
bash build/scripts/dev/start.sh
```
The webapp is mapped on the port `8003`.

### Test the webapp
To test the webapp, please open Postman and call this endpoint:
```
POST http://0.0.0.0:8003/api/v1/companies
```

### Env variables
This project is configured via environment variables that are declared and expected in the repository.

Please use:
- `.env` file to change configs of the app while working natively
- Check out `docker-compose.yaml` to override configs of the app when it's run as docker container

### Commands
To see the list of available commands run the following scripts from the home directory:
``` sh
go run ./cmd/cli/cli.go
```
The CLI will prompt all the available commands and you can select one of them to be run, accepting input parameters. E.g.
``` sh
go run ./cmd/cli/cli.go default-command --user-id 29382
```

## Style
This section helps in understanding the applied style of coding. Please follow it carefully. The golden rule is `consistency`!
### Export structs and methods
Avoid the exposure of services, repositories, routers outside their own package. The best way is to implement an Init() method of the package
that initialize all the components without the need of exposing them. E.g.
``` go
// > OK
func (r auditorRepository) getAuditorByID(...) (auditor, error)

// > NOT OK
func (r AuditorRepository) getAuditorByID(...) (Auditor, error)
```

### Method declaration
To keep consistency across the entire project, the method declaration leverage the value instead of the reference of the struct. E.g.
``` go
// > OK
func (r auditorRepository) getAuditorByID(...) (auditor, error)

// > NOT OK
func (r *auditorRepository) getAuditorByID(...) (auditor, error)
```
### Error Handling
Errors need to be returned as an expicit item. Avoid any definition of errors inside a struct for returing statement. E.g.
``` go
// > OK
func (r auditorRepository) getAuditorByID(...) (auditor, error)

// > NOT OK
type auditorDto struct {
  item   auditor
  err    error
}
func (r auditorRepository) getAuditorByID(...) auditorDto
```

### Return by Reference or Value
Avoid the return by reference if not really needed. E.g.
``` go
// > OK
func (r auditorRepository) getAuditorByID(...) (auditor, error) {
  // In case of error
  return auditor{}, err
}

// > NOT OK
func (r auditorRepository) getAuditorByID(...) (*auditor, error) {
  // In case of error
  return nil, err
}
```

### Env variables
Never access env variables directly form Controllers/Services/Repositories, but inject their values in the constructor. E.g.
``` go
// > OK
func newTextAnalyzerService(isServiceActive bool) {
  return TextAnalyzerService{isActive: isServiceActive}
}
// Inject in constructor
textAnalyzerService := newTextAnalyzerService(envs.TEXT_ANALYZER_ENABLED)

// > NOT OK
func newTextAnalyzerService() {
  return TextAnalyzerService{}
}
// definition without injecting
textAnalyzerService := newTextAnalyzerService()
...
func (s TextAnalyzerService) AnalyzeText(text string) (string, error) {
  // Access global env variable directly from the business logic
  s.Start(text, tc_env.ENVS.TEXT_ANALYZER_ENABLED)
}
```

### Context (Gin Context)
We should avoid passing the GIN Context from routers to underlayers (services, repositories) unless it is really needed. The idea is to ensure services and repositories are indipendent from the framework and ensure developers cannot access directly information from the request that has not been passed by router.
At the end, the router is the only component that can access the GIN Context and prepare DTOs to pass information to layer under itself. E.g. 
``` go
// > OK
func (s TextAnalyzerService).AnalyzeText(text string) (string, error) {
  // Access env variable directly from the business logic
  s.Start(text, s.isActive)
}
// > NOT OK
func (s TextAnalyzerService).AnalyzeText(_ *gin.Context, text string) (string, error) {
  // Access env variable directly from the business logic
  s.Start(text, s.isActive)
}
```

### External Service Call
In case of external call API are performed, ensure to define a timeout. E.g. 
``` go
client := http.Client{Timeout: 5 * time.Second}
client.Do(req)
```

### Package boundaries
A package must take into account its boundaries. When we need to access a specific model (e.g. DB table) that does not fall within its boundaries, it is important to re-declare the model with only the necessary fields it needs to access and not make any writes (read-only mode). In this way, it will be easier to avoid circular dependencies and facilitate migration to a microservice approach.
If, as a result of a change to one entity there is a subsequent change to another not of the same scope, it is appropriate to leverage the pubsub service to notify the package owner that it must react to a change made by another package.
