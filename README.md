# SharkCI

CI server written in Go.

> [!IMPORTANT]
> This project is under development and is not in working state yet.

## Download

### Go install

```
go install github.com/shark-ci/shark-ci/cmd/shark-ci-server # Download CI server
go install github.com/shark-ci/shark-ci/cmd/shark-ci-worker # Download CI runner
```

### Docker

```
docker pull ghcr.io/shark-ci/shark-ci-server:latest # Download CI server
docker pull ghcr.io/shark-ci/shark-ci-worker:latest # Download CI runner
```

## Architecture

![architecture](./docs/architecture.png)

## Env variables CI-Server

| Key                    | Default                         | Description               |
|------------------------|---------------------------------|---------------------------|
| `HOST`                 |                                 | Hostname                  |
| `PORT`                 | `8000`                          | Port                      |
| `SECRET_KEY`           |                                 | Random key for encryption |
| `DB_URI`               | `postgres://localhost/shark-ci` | Postgres URI              |
| `MQ_URI`               | `amqp://guest:guest@localhost`  | RabbitMQ URI              |
| `GITHUB_CLIENT_ID`     |                                 | GitHub client ID          |
| `GITHUB_CLIENT_SECRET` |                                 | GitHub client secret      |
| `GITLAB_CLIENT_ID`     |                                 | GitLab client ID          |
| `GITLAB_CLIENT_SECRET` |                                 | GitLab client secret      |

## Env variables worker

| Key             | Default                        | Description              |
|-----------------|--------------------------------|--------------------------|
| `SERVER_HOST`   | `localhost`                    | Server hostname          |
| `SERVER_PORT`   | `8000`                         | Server port              |
| `MQ_URI`        | `amqp://guest:guest@localhost` | RabbitMQ URI             |
| `REPOS_PATH`    | `./repos`                      | Path to repositories     |
