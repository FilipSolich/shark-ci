// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: pipeline.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPipeline = `-- name: CreatePipeline :one
INSERT INTO "pipeline" (status, clone_url, commit_sha, repo_id)
VALUES ($1, $2, $3, $4)
RETURNING id
`

type CreatePipelineParams struct {
	Status    PipelineStatus
	CloneUrl  string
	CommitSha string
	RepoID    int64
}

func (q *Queries) CreatePipeline(ctx context.Context, arg CreatePipelineParams) (int64, error) {
	row := q.db.QueryRow(ctx, createPipeline,
		arg.Status,
		arg.CloneUrl,
		arg.CommitSha,
		arg.RepoID,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const getPipelineCreationInfo = `-- name: GetPipelineCreationInfo :one
SELECT su.username, su.access_token, su.refresh_token, su.token_type, su.token_expire, r.name
FROM "service_user" su JOIN "repo" r ON su.id = r.service_user_id
WHERE r.id = $1
`

type GetPipelineCreationInfoRow struct {
	Username     string
	AccessToken  string
	RefreshToken pgtype.Text
	TokenType    string
	TokenExpire  pgtype.Timestamp
	Name         string
}

func (q *Queries) GetPipelineCreationInfo(ctx context.Context, id int64) (GetPipelineCreationInfoRow, error) {
	row := q.db.QueryRow(ctx, getPipelineCreationInfo, id)
	var i GetPipelineCreationInfoRow
	err := row.Scan(
		&i.Username,
		&i.AccessToken,
		&i.RefreshToken,
		&i.TokenType,
		&i.TokenExpire,
		&i.Name,
	)
	return i, err
}

const getPipelineStateChangeInfo = `-- name: GetPipelineStateChangeInfo :one
SELECT p.url, p.commit_sha, p.started_at, r.service, r.owner, r.name, su.access_token, su.refresh_token, su.token_type, su.token_expire
FROM public.pipeline p JOIN public.repo r ON p.repo_id = r.id JOIN public.service_user su ON r.service_user_id = su.id
WHERE p.id = $1
`

type GetPipelineStateChangeInfoRow struct {
	Url          pgtype.Text
	CommitSha    string
	StartedAt    pgtype.Timestamp
	Service      Service
	Owner        string
	Name         string
	AccessToken  string
	RefreshToken pgtype.Text
	TokenType    string
	TokenExpire  pgtype.Timestamp
}

func (q *Queries) GetPipelineStateChangeInfo(ctx context.Context, id int64) (GetPipelineStateChangeInfoRow, error) {
	row := q.db.QueryRow(ctx, getPipelineStateChangeInfo, id)
	var i GetPipelineStateChangeInfoRow
	err := row.Scan(
		&i.Url,
		&i.CommitSha,
		&i.StartedAt,
		&i.Service,
		&i.Owner,
		&i.Name,
		&i.AccessToken,
		&i.RefreshToken,
		&i.TokenType,
		&i.TokenExpire,
	)
	return i, err
}

const pipelineFinished = `-- name: PipelineFinished :exec
UPDATE "pipeline"
SET status = $1, finished_at = $2
WHERE id = $3
`

type PipelineFinishedParams struct {
	Status     PipelineStatus
	FinishedAt pgtype.Timestamp
	ID         int64
}

func (q *Queries) PipelineFinished(ctx context.Context, arg PipelineFinishedParams) error {
	_, err := q.db.Exec(ctx, pipelineFinished, arg.Status, arg.FinishedAt, arg.ID)
	return err
}

const pipelineStarted = `-- name: PipelineStarted :exec
UPDATE "pipeline"
SET status = $1, started_at = $2
WHERE id = $3
`

type PipelineStartedParams struct {
	Status    PipelineStatus
	StartedAt pgtype.Timestamp
	ID        int64
}

func (q *Queries) PipelineStarted(ctx context.Context, arg PipelineStartedParams) error {
	_, err := q.db.Exec(ctx, pipelineStarted, arg.Status, arg.StartedAt, arg.ID)
	return err
}

const setPipelineUrl = `-- name: SetPipelineUrl :exec
UPDATE "pipeline"
SET url = $1
WHERE id = $2
`

type SetPipelineUrlParams struct {
	Url pgtype.Text
	ID  int64
}

func (q *Queries) SetPipelineUrl(ctx context.Context, arg SetPipelineUrlParams) error {
	_, err := q.db.Exec(ctx, setPipelineUrl, arg.Url, arg.ID)
	return err
}
