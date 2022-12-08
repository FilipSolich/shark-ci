package sessions

import (
	"github.com/gorilla/sessions"

	"github.com/shark-ci/shark-ci/configs"
)

const SessionKey = "id"

var Store = sessions.NewCookieStore([]byte(configs.SessionSecret))
