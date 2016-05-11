package auth

import (
	"net/http"

	"github.com/gorilla/context"
)

func GetTeam(r *http.Request) (string, int, bool, bool) {
	teamName, namePresent := context.GetOk(r, teamNameKey)
	teamID, idPresent := context.GetOk(r, teamIDKey)
	isAdmin, adminPresent := context.GetOk(r, isAdminKey)

	if !(namePresent && idPresent && adminPresent) {
		return "", 0, false, false
	}

	return teamName.(string), teamID.(int), isAdmin.(bool), true
}
