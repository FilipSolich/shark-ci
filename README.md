# SharkCI

CI server written in Go

## Architecture

![architecture](./docs/architecture.png)

## Env variables CI-Server

| Key                    | Default                         | Description               |
|------------------------|---------------------------------|---------------------------|
| `HOST`                 |                                 | Hostname                  |
| `PORT`                 | `8080`                          | Port                      |
| `SECRET_KEY`           |                                 | Random key for encryption |
| `DB_URI`               | `postgres://localhost/shark-ci` | RabbitMQ URI              |
| `MQ_URI`               | `amqp://guest:guest@localhost`  | RabbitMQ URI              |
| `GITHUB_CLIENT_ID`     |                                 | GitHub client ID          |
| `GITHUB_CLIENT_SECRET` |                                 | GitHub client secret      |
| `GITLAB_CLIENT_ID`     |                                 | GitLab client ID          |
| `GITLAB_CLIENT_SECRET` |                                 | GitLab client secret      |

## Env variables worker

| Key             | Default                        | Description                                              |
|-----------------|--------------------------------|----------------------------------------------------------|
| `CISERVER_HOST` | `localhost`                    | CI server hostname                                       |
| `CISERVER_PORT` | `8000`                         | CI server port                                           |
| `MQ_URI`        | `amqp://guest:guest@localhost` | RabbitMQ URI                                             |
| `MAX_WORKERS`   | `N`                            | Maximum nuber of workers (Default is `runtime.NumCPU()`) |
| `REPOS_PATH`    | `./repos`                      | Path to repositories                                     |
