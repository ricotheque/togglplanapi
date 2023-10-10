# togglplanpi

A golang client interface for the [Toggl Plan API](https://developers.plan.toggl.com/api-v5.html).

## Setup
1. Register your application at the [Toggl Plan Developers Application Page](https://developers.plan.toggl.com/applications).
2. Load the following securely into your application:
  a. Toggl plan username
  b. Toggl plan password
  c. App key
  d. App secret
3. Pass these into `New()`:

```go
import (
    "fmt"
    "github.com/ricotheque/togglplanapi"
)

func main() {
    // Load your information securely
    username := getSecret("Toggl plan username")
    password := getSecret("Toggl plan password")
    clientId := getSecret("App key")
    clientSecret := getSecret("App secret")

    // Initialize a new Toggl Plan API client
    pa := togglplanapi.New(username, password, clientId, clientSecret, "")

    // Send a GET request to the Toggl Plan API
    result, err := togglplanapi.Request(pa, "https://api.plan.toggl.com/api/v5/me", "GET", []byte{}, map[string]string{})

    fmt.Println(result, err)

    // You can reuse the same client instance for another request
    result2, err2 := togglplanapi.Request(pa, "https://api.plan.toggl.com/api/v5/me", "GET", []byte{}, map[string]string{})

    fmt.Println(result2, err2)
}
```

## Response handling

`Request()` returns a string and an error. You'll need to unmarshall the string into a struct.

