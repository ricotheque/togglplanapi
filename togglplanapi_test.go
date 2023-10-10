package togglplanapi

import (
	"fmt"
	"testing"
)

const username = "[Toggl Plan username]"
const password = "[Toggl Plan password]"
const clientId = "[Toggl Plan App Key]"
const clientSecret = "[Toggl Plan Secret]"

func TestNormalRequest(t *testing.T) {
	pa := New(username, password, clientId, clientSecret, "")

	result, err := Request(pa, "https://api.plan.toggl.com/api/v5/me", "GET", []byte{}, map[string]string{})

	fmt.Println(result, err)

	result2, err2 := Request(pa, "https://api.plan.toggl.com/api/v5/me", "GET", []byte{}, map[string]string{})

	fmt.Println(result2, err2)
}
