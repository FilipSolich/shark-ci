CREATE TYPE service AS ENUM ('GitHub', 'GitLab');

CREATE TYPE pipeline_status AS ENUM ('success', 'pending', 'running', 'error');

CREATE TABLE "user" (
    id bigserial PRIMARY KEY,
    username text NOT NULL,
    email text NOT NULL
);

CREATE TABLE "service_user" (
    id bigserial PRIMARY KEY,
    service service NOT NULL,
    username text NOT NULL ,
    email text NOT NULL,
    access_token text NOT NULL ,
    refresh_token text,
    token_type text NOT NULL,
    token_expire timestamp,
    user_id bigint NOT NULL,
    UNIQUE (service, username),
    FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE CASCADE
);

CREATE TABLE "oauth2_state" (
    state uuid PRIMARY KEY,
    expire timestamp NOT NULL
);

CREATE TABLE "repo" (
    id bigserial PRIMARY KEY,
    service service NOT NULL,
    owner text NOT NULL,
    name text NOT NULL,
    repo_service_id bigint NOT NULL,
    webhook_id bigint NOT NULL,
    service_user_id bigint NOT NULL,
    UNIQUE (service, repo_service_id),
    UNIQUE (service, webhook_id),
    UNIQUE (service, owner, name),
    FOREIGN KEY (service_user_id) REFERENCES "service_user" (id) ON DELETE CASCADE
);

CREATE TABLE "pipeline" (
    id bigserial PRIMARY KEY,
    url text UNIQUE,
    status pipeline_status NOT NULL,
    clone_url text NOT NULL,
    commit_sha text NOT NULL,
    started_at timestamp,
    finished_at timestamp,
    repo_id bigint NOT NULL,
    FOREIGN KEY (repo_id) REFERENCES "repo" (id) ON DELETE CASCADE
);
