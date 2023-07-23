# SharkCI

CI server written in Go

## Architecture

![architecture](./docs/architecture.png)

## Env variables

| Key                    | Default                              | Description                |
|------------------------|--------------------------------------|----------------------------|
| `HOST`                 |                                      | Hostname                   |
| `PORT`                 | `8080`                               | Port                       |
| `SECRET_KEY`           |                                      | Random key for encryption  |
| `DB_URI`               | `postgres://localhost:5432/shark-ci` | RabbitMQ URI               |
| `MQ_URI`               | `amqp://localhost:5672`              | RabbitMQ URI               |
| `GITHUB_CLIENT_ID`     |                                      | GitHub client ID           |
| `GITHUB_CLIENT_SECRET` |                                      | GitHub client secret       |
| `GITLAB_CLIENT_ID`     |                                      | GitLab client ID           |
| `GITLAB_CLIENT_SECRET` |                                      | GitLab client secret       |
