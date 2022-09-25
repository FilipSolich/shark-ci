# CI Server

## Env variables

At least one git service must be enabled (`GITHUB_SERVICE` or `GITLAB_SERVICE`)

| Key                    | Type            | Default           | Description                                                         |
|------------------------|-----------------|-------------------|---------------------------------------------------------------------|
| `HOST`                 | `string`        |                   | Hostname.                                                           |
| `PORT`                 | `int`\|`string` | 8080              | Port                                                                |
| `SESSION_SECRET`       | `string`        | "insecure-secret" | Random key for session encryption.                                  |
| `CSRF_SECRET`          | `string`        | "insecure-secret" | Random key for csrf encryption.                                     |
| `WEBHOOK_SECRET`       | `string`        | "insecure-secret" | Random key for webhook encryption.                                  |
| `RABBITMQ_HOST`        | `string`        | "localhost"       | RabbtiMQ hostname.                                                  |
| `RABBITMQ_PORT`        | `int`\|`string` | 5672              | RabbitMQ port.                                                      |
| `RABBITMQ_USERNAME`    | `string`        | "guest"           | RabbitMQ username.                                                  |
| `RABBITMQ_PASSWORD`    | `string`        | "guest"           | RabbitMQ password.                                                  |
| `GITHUB_SERVICE`       | `bool`          | `false`           | Allow GitHub repositories for CI.                                   |
| `GITHUB_CLIENT_ID`     | `string`        |                   | GitHub client ID (Required only if `GITHUB_SERVICE` == `true`).     |
| `GITHUB_CLIENT_SECRET` | `string`        |                   | GitHub client secret (Required only if `GITHUB_SERVICE` == `true`). |
| `GITLAB_SERVICE`       | `bool`          | `false`           | Allow GitLab repositories for CI.                                   |
| `GITLAB_CLIENT_ID`     | `string`        |                   | GitLab client ID (Required only if `GITLAB_SERVICE` == `true`).     |
| `GITLAB_CLIENT_SECRET` | `string`        |                   | GitLab client secret (Required only if `GITLAB_SERVICE` == `true`). |
