# Writing tests for Cloudinary SDK

When you are creating a new feature it is highly recommended adding some tests for it.
There are two types of tests in Cloudinary SDK: E2E and acceptance.

### End-to-End tests (E2E)
You can find E2E tests in `*_test.go` files.
The main purpose of these tests is to check the main scenarios in the SDK integration with Cloudinary server.
These tests performing some HTTP queries to the Cloudinary server and requires `CLOUDINARY_URL` to be given to run.

#### Writing an E2E test
Check [api/admin/api_test.go](api/admin/api_test.go) for an example of E2E test.
These tests are using a basic Go test approach [https://golang.org/pkg/testing/](https://golang.org/pkg/testing/).

*Example of E2E test:*
```go

func TestAPI_Timeout(t *testing.T) {
	var originalTimeout = adminAPI.Config.API.Timeout

	adminAPI.Config.API.Timeout = 0 // should timeout immediately

	_, err := adminAPI.Ping(ctx)

	if err == nil || !strings.HasSuffix(err.Error(), "context deadline exceeded") {
		t.Error("Expected context timeout did not happen")
	}

	adminAPI.Config.API.Timeout = originalTimeout
}

```


### Acceptance tests
These tests are placed in `*_acceptance_test.go` files.
The main purpose of these tests is to check whether SDK generates proper HTTP requests or not with mocked HTTP-server.

#### Writing an acceptance test
Check [api/admin/asset_acceptance_test.go](api/admin/assets_acceptance_test.go) for an example of acceptance test.
These tests are using a basic Go test approach with [httptest](https://golang.org/pkg/net/http/httptest/) as a mocking HTTP server.


A basic function to run in test has this reference:
```go
func testApiByTestCases(cases []ApiAcceptanceTestCase, t *testing.T)
```

ApiAcceptanceTestCase is a structure that defines a test case for the test:

```go
// Acceptance test case definition. See `TEST.md` for an additional information.
type ApiAcceptanceTestCase struct {
	Name              string                // Name of the test case
	RequestTest       ApiRequestTest        // Function which will be called as a API request. Put SDK calls here.
	ResponseTest      ApiResponseTest       // Function which will be called to test an API response.
	ExpectedRequest   expectedRequestParams // Expected HTTP request to be sent to the server
	JsonResponse      string                // Mock of the JSON response from server. This is used to check JSON parsing.
	ExpectedStatus    string                // Expected HTTP status of the request. This status will be returned from the HTTP mock.
	ExpectedCallCount int                   // Expected call count to the server.
}
```

*Example of test case:*

```go
{
			Name: "Ping",
			RequestTest: func(api *admin.API, ctx context.Context) (interface{}, error) {
				return api.Ping(ctx)
			},
			ResponseTest: func(response interface{}, t *testing.T) {
				_, ok := response.(*admin.PingResult)
				if !ok {
					t.Errorf("Response should be type of PingResult, %s given", reflect.TypeOf(response))
				}
			},
			ExpectedRequest:   expectedRequestParams{Method: "GET", Uri: "/ping"},
			JsonResponse:      "{\"status\": \"OK\"}",
			ExpectedCallCount: 1,
		},
```

You can find more examples in [api/admin/asset_acceptance_test.go](api/admin/assets_acceptance_test.go).