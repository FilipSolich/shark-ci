CREATE TABLE IF NOT EXISTS "user" (
    id bigserial PRIMARY KEY,
    username text,
    email text NOT NULL
);

CREATE TABLE IF NOT EXISTS "service_user" (
    id bigserial PRIMARY KEY,
    service text NOT NULL,
    username text NOT NULL ,
    email text NOT NULL,
    access_token text NOT NULL ,
    refresh_token text,
    token_type text,
    token_expire timestamp,
    user_id bigint NOT NULL ,
    FOREIGN KEY (user_id) REFERENCES "user" (id)
);

CREATE TABLE IF NOT EXISTS "oauth2_state" (
    state uuid PRIMARY KEY,
    expire timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS "repo" (
    id bigserial PRIMARY KEY,
    repo_service_id bigint NOT NULL ,
    name text NOT NULL,
    service text NOT NULL,
    webhook_id bigint,
    webhook_active boolean,
    service_user_id bigint NOT NULL ,
    FOREIGN KEY (service_user_id) REFERENCES "service_user" (id)
);

CREATE TABLE IF NOT EXISTS "pipeline" (
    id bigserial PRIMARY KEY,
    commit_sha text NOT NULL,
    clone_url text NOT NULL,
    status text,
    started_at timestamp,
    finished_at timestamp,
    access_token text NOT NULL,
    refresh_token text,
    token_type text,
    token_expire timestamp,
    repo_id bigint NOT NULL,
    FOREIGN KEY (repo_id) REFERENCES "repo" (id)
);

CREATE TYPE log_line AS (
    line bigint,
    file text,
    content text
);

CREATE TABLE IF NOT EXISTS "pipeline_log" (
    id bigserial PRIMARY KEY,
    started_at timestamp,
    finished_at timestamp,
    cmd text,
    output log_line[],
    return_code int,
    pipeline_id bigint,
    FOREIGN KEY (pipeline_id) REFERENCES "pipeline" (id)
);
