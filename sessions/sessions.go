package sessions

import (
	"github.com/gorilla/sessions"

	"github.com/FilipSolich/ci-server/configs"
)

const SessionKey = "id"

var Store = sessions.NewCookieStore([]byte(configs.SessionSecret))
