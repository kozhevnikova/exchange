package main

import (
	"net/http"
	"strconv"
)

func (user *StoreUser) SetCookieValues(w http.ResponseWriter) error {
	value := map[string]string{
		"login":  user.Login,
		"userid": strconv.Itoa(user.Userid),
		"role":   user.Role,
	}
	if encoded, err := sc.Encode(cookieName, value); err == nil {
		cookie := &http.Cookie{
			Name:     cookieName,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
		}
		if err != nil {
			return err
		}
		http.SetCookie(w, cookie)
	}
	return nil
}

func readCookies(r *http.Request) (int, string, string, error) {
	var userid int
	var login string
	var role string

	if cookie, err := r.Cookie(cookieName); err == nil {
		value := make(map[string]string)
		if err = sc.Decode(cookieName, cookie.Value, &value); err == nil {
			login = value["login"]
			userid, err = strconv.Atoi(value["userid"])
			if err != nil {
				return 0, "", "", err
			}
			role = value["role"]
		}
	} else {
		return 0, "", "", err
	}
	return userid, login, role, nil
}
