# SharkCI

CI server written in Go.

> [!IMPORTANT]
> This project is under development and is not in working state yet.

## Download

### Go install

```
go install github.com/shark-ci/shark-ci/cmd/server # Download CI server
go install github.com/shark-ci/shark-ci/cmd/worker # Download CI runner
```

### Docker

```
docker pull ghcr.io/shark-ci/server:latest # Download CI server
docker pull ghcr.io/shark-ci/worker:latest # Download CI runner
```

## Architecture

![architecture](./docs/architecture.png)

## Env variables CI-Server

| Key                    | Default                         | Description               |
|------------------------|---------------------------------|---------------------------|
| `HOST`                 | `localhost`                     | Hostname                  |
| `PORT`                 | `8000`                          | Port                      |
| `GRPC_PORT`            | `9000`                          | GRPc port                 |
| `SECRET_KEY`           |                                 | Random key for encryption |
| `DB_URI`               | `postgres://localhost/shark-ci` | Postgres URI              |
| `MQ_URI`               | `amqp://guest:guest@localhost`  | RabbitMQ URI              |
| `GITHUB_CLIENT_ID`     |                                 | GitHub client ID          |
| `GITHUB_CLIENT_SECRET` |                                 | GitHub client secret      |
| `GITLAB_CLIENT_ID`     |                                 | GitLab client ID          |
| `GITLAB_CLIENT_SECRET` |                                 | GitLab client secret      |

## Env variables worker

| Key           | Default                        | Description           |
|---------------|--------------------------------|-----------------------|
| `HOST`        | `localhost`                    | Server hostname       |
| `GRPC_PORT`   | `9000`                         | Server port           |
| `MAX_WORKERS` | env `GOMAXPROC`                | Max number of workers |
| `MQ_URI`      | `amqp://guest:guest@localhost` | RabbitMQ URI          |
| `REPOS_PATH`  | `./repos`                      | Path to repositories  |
