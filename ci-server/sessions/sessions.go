package sessions

import (
	"github.com/gorilla/sessions"
)

const SessionKey = "id"

var Store *sessions.CookieStore

func InitSessionStore(secretKey string) {
	Store = sessions.NewCookieStore([]byte(secretKey))
}
