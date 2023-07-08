# SharkCI

CI server written in Go

## Architecture

![architecture](./docs/architecture.png)

## Env variables

At least one git service must be enabled (`GITHUB_ENABLED` or `GITLAB_ENABLED`)

| Key                    | Type            | Default                       | Description                |
|------------------------|-----------------|-------------------------------|----------------------------|
| `HOST`                 | `string`        |                               | Hostname.                  |
| `PORT`                 | `int`\|`string` | `8080`                        | Port.                      |
| `SECRET_SECRET`        | `string`        | `"insecure-secret"`           | Random key for encryption. |
| `MONGO_URI`            | `string`        | `"mongodb://localhost:17017"` | RabbitMQ URI.              |
| `RABBITMQ_URI`         | `string`        | `"amqp://localhost:5672"`     | RabbitMQ URI.              |
| `GITHUB_CLIENT_ID`     | `string`        |                               | GitHub client ID.          |
| `GITHUB_CLIENT_SECRET` | `string`        |                               | GitHub client secret.      |
| `GITLAB_CLIENT_ID`     | `string`        |                               | GitLab client ID.          |
| `GITLAB_CLIENT_SECRET` | `string`        |                               | GitLab client secret.      |
