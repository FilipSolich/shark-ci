package sessions

import (
	"github.com/FilipSolich/ci-server/configs"
	"github.com/gorilla/sessions"
)

const SessionKey = "id"

var Store = sessions.NewCookieStore([]byte(configs.SessionSecret))
