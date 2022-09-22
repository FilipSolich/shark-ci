# CI Server

## Env variables

| Key                    | Type     | Description                                                         |
|------------------------|----------|---------------------------------------------------------------------|
| `SESSION_KEY`          | `string` | Random key for session encryption.                                  |
| `CSRF_KEY`             | `string` | Random key for csrf encryption.                                     |
| `GITHUB_SERVICE`       | `bool`   | Allow GitHub repositories for CI.                                   |
| `GITHUB_CLIENT_ID`     | `string` | GitHub client ID (Required only if `GITHUB_SERVICE` == `true`).     |
| `GITHUB_CLIENT_SECRET` | `string` | GitHub client secret (Required only if `GITHUB_SERVICE` == `true`). |
| `GITLAB_SERVICE`       | `bool`   | Allow GitLab repositories for CI.                                   |
| `GITLAB_CLIENT_ID`     | `string` | GitLab client ID (Required only if `GITLAB_SERVICE` == `true`).     |
| `GITLAB_CLIENT_SECRET` | `string` | GitLab client secret (Required only if `GITLAB_SERVICE` == `true`). |
