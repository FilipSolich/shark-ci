package types

import "golang.org/x/oauth2"

type ServiceUserRepoFetchInfo struct {
	ID      int64
	Service string
	Token   oauth2.Token
}
