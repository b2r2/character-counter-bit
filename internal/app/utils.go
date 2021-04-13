package app

import "strings"

func verifyUser(user string, accessUsers ...string) (state bool) {
	for _, u := range accessUsers {
		if u == user {
			state = true
			break
		}
	}
	return
}

func verifyLink(msg, medium, domain string) (state bool) {
	line := strings.Split(msg, "://")
	if len(line) == 2 {
		name := strings.Split(line[1], ".")
		switch name[0] {
		case medium, domain:
			state = true
		}
	}
	return
}
