package token

import "net/http"

func ParseUserNameFromToken(req *http.Request) string {
	// TODO
	return req.Header.Get("user")
}
