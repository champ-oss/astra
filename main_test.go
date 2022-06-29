package main

import (
	"github.com/google/go-github/v43/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var testUnixTime = int64(1642708819)
var testTime = time.Unix(testUnixTime, 0)
var testOwner = "owner1"
var testRepo = "repo1"
var testBranch = "branch1"
var testSha = "702fe8ee76f422edd4bc257a0a2171af26563063"

var testWorkflowRun = &github.WorkflowRun{
	Name:       github.String("test run"),
	HeadSHA:    github.String(testSha),
	HeadBranch: github.String(testBranch),
	Conclusion: github.String("success"),
	CreatedAt:  &github.Timestamp{Time: testTime},
	UpdatedAt:  &github.Timestamp{Time: testTime},
	RunAttempt: github.Int(1),
}

// Needs to be added to upstream github.com/migueleliasweb/go-github-mock
var GetWorkflowRunByID = mock.EndpointPattern{
	Pattern: "/repos/{owner}/{repo}/actions/runs/{run_id}",
	Method:  "GET",
}

func Test_loadConfig(t *testing.T) {
	_ = os.Setenv("INPUT_APP_ID", "123")
	_ = os.Setenv("INPUT_INSTALLATION_ID", "123")
	_ = os.Setenv("INPUT_PEM", "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUMyVXJtTXorZlhOMkZDWkxDQVRKSkVoSnJZVlp5WG12WWNuTXdyM1ZMTitReE04V0tKCk9rMEhTaXhhaDRJdzQ1bll5c01tRm91R1J2dWtWQ0UrZm0wb0g0ZkVjRFZ0OU1INitsdGVlRUgvMVg1cnRpNmIKVWtXdzBEcktmR09teVhUWTVCWHhSVWQ1Y3pkaWcvSThmMnJoNGk0K3VyV0NHenBETlV1dG1mcDlKd0lEQVFBQgpBb0dBYmpNNks3NU9aMnIxd21lUnR6cVEvaEVZZHNIb1VFbzlqN1hHUW8wWHk1OUlyQWtLZ2Q5WFI1eXhpbFoxCmZvOVRJaElNT2kxT1QrNy9rcWUzSUVyU05uVU5xem9DNzdBeXBZbkRGbDV1VjZSeHlKeW52WVFYbHNCVEUyTEoKVDRKaWNEeFZYbmJKQllKSzJpb0FkcU45dnVnN2FPc2sxU3ltWmFES2YwRjd3dUVDUVFEWVJ2Y0xERUYraFlDQQpXYXk5UGsxdVQ2VjJ3UlNxYjdxMjNlWmlXamhyT1dGd0JRSkRNb0NpNnZ2RkJ6Vm8zSUNZVVh1Y1htRzJVMGYxCjFmR0VSYy9MQWtFQTE4OUtOR1J3UDRWZHg1U25xaTVXVkp6WEQ1Tm92MU8wOUtzNGZneUV0SUp2azZ2V3o1S28KQ05UcWphRXEvcjdCMUF1bDVBSjlOdkZCWEVHL29HQWtsUUpCQU5UK1hvRjAybk5keXNXY2l1LzhrWWtYeXg1KwozSGxWZTQ1b1RtR0I5Sm8wY204OW41TEtBOEZ1cGZET1BwMDh1ekJHM3ZPS1I3U2xvL0xKZGdjTU1hMENRUUN6Cjc3YnNOaTVOR0RMWC9IOUxhclU2ZVViclNyb1VoSU9sV0xtU2gzZUNWaHNYNGpnSi9EcTBtbW95eW9WaHY4VTIKdXJ1SGYvZk0vcHpEZ21KM0lwSjlBa0JEajRQTDlnVWF4d0VqY0crSjNJbmJSWnlGU0NFVHVnajllREpWeFprNApsVjhnS1NpT2JVRFViMkJQM29SekZPTHFZbXhveHNnTUt4RWNxbGVTRzg1MQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=")
	_ = os.Setenv("INPUT_REPO_PREFIXES", "foo")
	_ = os.Setenv("INPUT_ACTORS", "foo")
	_ = os.Setenv("INPUT_DEFAULT_BRANCH", "main")
	_ = os.Setenv("INPUT_WAIT_SECONDS_BETWEEN_REQUESTS", "1")
	_ = os.Setenv("INPUT_MAX_RUN_ATTEMPTS", "1")
	_ = os.Setenv("EXPECT_REQUIRED_APPROVING_REVIEW_COUNT", "1")
	_ = os.Setenv("EXPECT_REQUIRED_STATUS_CHECKS", "1")
	loadConfig()
}

func Test_containsString(t *testing.T) {
	t.Parallel()
	t.Run("does contain", func(t *testing.T) {
		assert.True(t, containsString([]string{"foo", "bar"}, "bar"))
	})
	t.Run("does not contain", func(t *testing.T) {
		assert.False(t, containsString([]string{"foo", "bar"}, "nope"))
	})
	t.Run("is empty", func(t *testing.T) {
		assert.False(t, containsString([]string{}, "nope"))
	})
}

func Test_getEnvString(t *testing.T) {
	t.Parallel()
	key := "TEST6"
	value := "foo"
	expected := "foo"

	t.Run("set correctly", func(t *testing.T) {
		_ = os.Setenv(key, value)
		assert.Equal(t, expected, getEnvString(key))
		_ = os.Unsetenv(key)
	})

	t.Run("not set", func(t *testing.T) {
		_ = os.Unsetenv(key)
		assert.Panicsf(t, func() { getEnvString(key) }, "env variable TEST1 must be set")
	})

	t.Run("set empty", func(t *testing.T) {
		_ = os.Setenv(key, "")
		assert.Panicsf(t, func() {
			getEnvString(key)
		}, "env variable TEST1 must be set")
		_ = os.Unsetenv(key)
	})
}

func Test_getEnvBase64(t *testing.T) {
	t.Parallel()
	key := "TEST1"
	b64 := "dmFsdWUxMjM="
	expected := []byte("value123")

	t.Run("set correctly", func(t *testing.T) {
		_ = os.Setenv(key, b64)
		assert.Equal(t, expected, getEnvBase64(key))
		_ = os.Unsetenv(key)
	})

	t.Run("not set", func(t *testing.T) {
		_ = os.Unsetenv(key)
		assert.Panicsf(t, func() { getEnvBase64(key) }, "env variable TEST1 must be set")
	})

	t.Run("set empty", func(t *testing.T) {
		_ = os.Setenv(key, "")
		assert.Panicsf(t, func() {
			getEnvBase64(key)
		}, "env variable TEST1 must be set")
		_ = os.Unsetenv(key)
	})

	t.Run("parse error", func(t *testing.T) {
		_ = os.Setenv(key, "foo")
		assert.Panicsf(t, func() { getEnvBase64(key) }, "env variable TEST2 must be set")
		_ = os.Unsetenv(key)
	})
}

func Test_getEnvInt(t *testing.T) {
	t.Parallel()
	key := "TEST5"
	value := "293856293239328567"
	expected := 293856293239328567

	t.Run("set correctly", func(t *testing.T) {
		_ = os.Setenv(key, value)
		assert.Equal(t, expected, getEnvInt(key))
		_ = os.Unsetenv(key)
	})

	t.Run("not set", func(t *testing.T) {
		_ = os.Unsetenv(key)
		assert.Panicsf(t, func() { getEnvInt(key) }, "env variable TEST2 must be set")
	})

	t.Run("set empty", func(t *testing.T) {
		_ = os.Setenv(key, "")
		assert.Panicsf(t, func() { getEnvInt(key) }, "env variable TEST2 must be set")
		_ = os.Unsetenv(key)
	})

	t.Run("parse error", func(t *testing.T) {
		_ = os.Setenv(key, "foo")
		assert.Panicsf(t, func() { getEnvInt(key) }, "env variable TEST2 must be set")
		_ = os.Unsetenv(key)
	})
}

func Test_getEnvInt64(t *testing.T) {
	t.Parallel()
	key := "TEST2"
	value := "293856293239328567"
	expected := int64(293856293239328567)

	t.Run("set correctly", func(t *testing.T) {
		_ = os.Setenv(key, value)
		assert.Equal(t, expected, getEnvInt64(key))
		_ = os.Unsetenv(key)
	})

	t.Run("not set", func(t *testing.T) {
		_ = os.Unsetenv(key)
		assert.Panicsf(t, func() { getEnvInt64(key) }, "env variable TEST2 must be set")
	})

	t.Run("set empty", func(t *testing.T) {
		_ = os.Setenv(key, "")
		assert.Panicsf(t, func() { getEnvInt64(key) }, "env variable TEST2 must be set")
		_ = os.Unsetenv(key)
	})

	t.Run("parse error", func(t *testing.T) {
		_ = os.Setenv(key, "foo")
		assert.Panicsf(t, func() { getEnvInt64(key) }, "env variable TEST2 must be set")
		_ = os.Unsetenv(key)
	})
}

func Test_getEnvBool(t *testing.T) {
	t.Parallel()
	key := "TEST3"

	t.Run("true", func(t *testing.T) {
		_ = os.Setenv(key, "true")
		assert.True(t, getEnvBool(key, false))
		_ = os.Unsetenv(key)

		_ = os.Setenv(key, "True")
		assert.True(t, getEnvBool(key, false))
		_ = os.Unsetenv(key)
	})

	t.Run("false", func(t *testing.T) {
		_ = os.Setenv(key, "false")
		assert.False(t, getEnvBool(key, false))
		_ = os.Unsetenv(key)

		_ = os.Setenv(key, "False")
		assert.False(t, getEnvBool(key, false))
		_ = os.Unsetenv(key)
	})

	t.Run("invalid", func(t *testing.T) {
		_ = os.Setenv(key, "nonsense")
		assert.False(t, getEnvBool(key, false))
		_ = os.Unsetenv(key)
	})

	t.Run("unset", func(t *testing.T) {
		_ = os.Unsetenv(key)
		assert.False(t, getEnvBool(key, false))

		_ = os.Unsetenv(key)
		assert.True(t, getEnvBool(key, true))
	})
}

func Test_getEnvStringList(t *testing.T) {
	t.Parallel()
	key := "TEST4"
	expected := []string{"item1", "item2", "item3"}

	t.Run("set correctly", func(t *testing.T) {
		_ = os.Setenv(key, "item1\nitem2\nitem3")
		assert.Equal(t, expected, getEnvStringList(key))
		_ = os.Unsetenv(key)
	})

	t.Run("trailing newline", func(t *testing.T) {
		_ = os.Setenv(key, "item1\nitem2\nitem3\n")
		assert.Equal(t, expected, getEnvStringList(key))
		_ = os.Unsetenv(key)
	})

	t.Run("extra space", func(t *testing.T) {
		_ = os.Setenv(key, "item1\nitem2\nitem3\n ")
		assert.Equal(t, expected, getEnvStringList(key))
		_ = os.Unsetenv(key)
	})

	t.Run("not set", func(t *testing.T) {
		_ = os.Unsetenv(key)
		assert.Panicsf(t, func() { getEnvStringList(key) }, "env variable TEST4 must be set")
	})
}

func Test_getClientV4(t *testing.T) {
	_ = os.Setenv("DUMMY_KEY", "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUMyVXJtTXorZlhOMkZDWkxDQVRKSkVoSnJZVlp5WG12WWNuTXdyM1ZMTitReE04V0tKCk9rMEhTaXhhaDRJdzQ1bll5c01tRm91R1J2dWtWQ0UrZm0wb0g0ZkVjRFZ0OU1INitsdGVlRUgvMVg1cnRpNmIKVWtXdzBEcktmR09teVhUWTVCWHhSVWQ1Y3pkaWcvSThmMnJoNGk0K3VyV0NHenBETlV1dG1mcDlKd0lEQVFBQgpBb0dBYmpNNks3NU9aMnIxd21lUnR6cVEvaEVZZHNIb1VFbzlqN1hHUW8wWHk1OUlyQWtLZ2Q5WFI1eXhpbFoxCmZvOVRJaElNT2kxT1QrNy9rcWUzSUVyU05uVU5xem9DNzdBeXBZbkRGbDV1VjZSeHlKeW52WVFYbHNCVEUyTEoKVDRKaWNEeFZYbmJKQllKSzJpb0FkcU45dnVnN2FPc2sxU3ltWmFES2YwRjd3dUVDUVFEWVJ2Y0xERUYraFlDQQpXYXk5UGsxdVQ2VjJ3UlNxYjdxMjNlWmlXamhyT1dGd0JRSkRNb0NpNnZ2RkJ6Vm8zSUNZVVh1Y1htRzJVMGYxCjFmR0VSYy9MQWtFQTE4OUtOR1J3UDRWZHg1U25xaTVXVkp6WEQ1Tm92MU8wOUtzNGZneUV0SUp2azZ2V3o1S28KQ05UcWphRXEvcjdCMUF1bDVBSjlOdkZCWEVHL29HQWtsUUpCQU5UK1hvRjAybk5keXNXY2l1LzhrWWtYeXg1KwozSGxWZTQ1b1RtR0I5Sm8wY204OW41TEtBOEZ1cGZET1BwMDh1ekJHM3ZPS1I3U2xvL0xKZGdjTU1hMENRUUN6Cjc3YnNOaTVOR0RMWC9IOUxhclU2ZVViclNyb1VoSU9sV0xtU2gzZUNWaHNYNGpnSi9EcTBtbW95eW9WaHY4VTIKdXJ1SGYvZk0vcHpEZ21KM0lwSjlBa0JEajRQTDlnVWF4d0VqY0crSjNJbmJSWnlGU0NFVHVnajllREpWeFprNApsVjhnS1NpT2JVRFViMkJQM29SekZPTHFZbXhveHNnTUt4RWNxbGVTRzg1MQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=")
	dummyKey := getEnvBase64("DUMMY_KEY")

	t.Run("successful", func(t *testing.T) {
		client, err := getClientV4(123, 123, dummyKey)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("failed", func(t *testing.T) {
		client, err := getClientV4(123, 123, []byte("bad key"))
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func Test_getClient(t *testing.T) {
	_ = os.Setenv("DUMMY_KEY", "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUMyVXJtTXorZlhOMkZDWkxDQVRKSkVoSnJZVlp5WG12WWNuTXdyM1ZMTitReE04V0tKCk9rMEhTaXhhaDRJdzQ1bll5c01tRm91R1J2dWtWQ0UrZm0wb0g0ZkVjRFZ0OU1INitsdGVlRUgvMVg1cnRpNmIKVWtXdzBEcktmR09teVhUWTVCWHhSVWQ1Y3pkaWcvSThmMnJoNGk0K3VyV0NHenBETlV1dG1mcDlKd0lEQVFBQgpBb0dBYmpNNks3NU9aMnIxd21lUnR6cVEvaEVZZHNIb1VFbzlqN1hHUW8wWHk1OUlyQWtLZ2Q5WFI1eXhpbFoxCmZvOVRJaElNT2kxT1QrNy9rcWUzSUVyU05uVU5xem9DNzdBeXBZbkRGbDV1VjZSeHlKeW52WVFYbHNCVEUyTEoKVDRKaWNEeFZYbmJKQllKSzJpb0FkcU45dnVnN2FPc2sxU3ltWmFES2YwRjd3dUVDUVFEWVJ2Y0xERUYraFlDQQpXYXk5UGsxdVQ2VjJ3UlNxYjdxMjNlWmlXamhyT1dGd0JRSkRNb0NpNnZ2RkJ6Vm8zSUNZVVh1Y1htRzJVMGYxCjFmR0VSYy9MQWtFQTE4OUtOR1J3UDRWZHg1U25xaTVXVkp6WEQ1Tm92MU8wOUtzNGZneUV0SUp2azZ2V3o1S28KQ05UcWphRXEvcjdCMUF1bDVBSjlOdkZCWEVHL29HQWtsUUpCQU5UK1hvRjAybk5keXNXY2l1LzhrWWtYeXg1KwozSGxWZTQ1b1RtR0I5Sm8wY204OW41TEtBOEZ1cGZET1BwMDh1ekJHM3ZPS1I3U2xvL0xKZGdjTU1hMENRUUN6Cjc3YnNOaTVOR0RMWC9IOUxhclU2ZVViclNyb1VoSU9sV0xtU2gzZUNWaHNYNGpnSi9EcTBtbW95eW9WaHY4VTIKdXJ1SGYvZk0vcHpEZ21KM0lwSjlBa0JEajRQTDlnVWF4d0VqY0crSjNJbmJSWnlGU0NFVHVnajllREpWeFprNApsVjhnS1NpT2JVRFViMkJQM29SekZPTHFZbXhveHNnTUt4RWNxbGVTRzg1MQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=")
	dummyKey := getEnvBase64("DUMMY_KEY")

	t.Run("successful", func(t *testing.T) {
		client, err := getClient(123, 123, dummyKey)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("failed", func(t *testing.T) {
		client, err := getClient(123, 123, []byte("bad key"))
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func Test_processRepos(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data": {"repository": {"id": "abc123"}}}`)
	})
	clientV4 := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
	client := github.NewClient(&http.Client{})
	processRepos([]string{"abc123"}, client, clientV4, testOwner)
}

func Test_runRepoPullRequestsQuery(t *testing.T) {
	// https://github.com/shurcooL/githubv4/blob/master/githubv4_test.go
	t.Run("successful", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data": {"repository": {"id": "abc123"}}}`)
		})
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		results, err := runRepoPullRequestsQuery(client, "owner1", "repo1")
		assert.NoError(t, err)
		assert.NotNil(t, results)
	})

	t.Run("failure", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, ``)
		})
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		results, err := runRepoPullRequestsQuery(client, "owner1", "repo1")
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func Test_enablePullRequestAutoMerge(t *testing.T) {
	// https://github.com/shurcooL/githubv4/blob/master/githubv4_test.go
	t.Run("successful", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data": {"enablePullRequestAutoMerge": {"clientMutationId": "abc123"}}}`)
		})
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		err := enablePullRequestAutoMerge(client, githubv4.ID("abc123"))
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		dryRun = false
		mux := http.NewServeMux()
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, ``)
		})
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		err := enablePullRequestAutoMerge(client, githubv4.ID("abc123"))
		assert.Error(t, err)
	})

	t.Run("dry run", func(t *testing.T) {
		dryRun = true
		client := githubv4.NewClient(&http.Client{})
		err := enablePullRequestAutoMerge(client, githubv4.ID("abc123"))
		assert.NoError(t, err)
		dryRun = false
	})
}

type localRoundTripper struct {
	handler http.Handler
}

func (l localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.handler.ServeHTTP(w, req)
	return w.Result(), nil
}

func Test_setLogging(t *testing.T) {
	_ = os.Setenv("DEBUG", "false")
	setLogging("DEBUG")
}

func Test_setDryRun(t *testing.T) {
	_ = os.Setenv("DRY_RUN", "false")
	setDryRun("DRY_RUN")
}

func Test_processPullRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pr := &pullRequest{}
		pr.Author.Login = "dependabot"
		pr.ID = githubv4.ID("abc123")
		pr.AutoMergeRequest.EnabledAt = ""
		dryRun = true
		client := github.NewClient(&http.Client{})
		clientV4 := githubv4.NewClient(&http.Client{})
		processPullRequest(client, clientV4, *pr, "owner1", "repo1", true)
	})

	t.Run("skip author", func(t *testing.T) {
		pr := &pullRequest{}
		pr.Author.Login = "foo"
		pr.ID = githubv4.ID("abc123")
		pr.AutoMergeRequest.EnabledAt = ""
		client := github.NewClient(&http.Client{})
		clientV4 := githubv4.NewClient(&http.Client{})
		processPullRequest(client, clientV4, *pr, "owner1", "repo1", true)
	})
}

func Test_filterRepoNames(t *testing.T) {
	repoPrefixes = []string{"terraform-", "foo"}
	repos := []repositoryName{
		{
			Name: githubv4.String("terraform-aws-foo"),
		},
		{
			Name: githubv4.String("something-else"),
		},
		{
			Name: githubv4.String("foo"),
		},
		{
			Name: githubv4.String("terraform-blah"),
		},
	}
	results := filterRepoNames(repos)
	expected := []string{"terraform-aws-foo", "foo", "terraform-blah"}
	assert.Equal(t, expected, results)
}

func Test_getAllOrgRepos(t *testing.T) {
	// https://github.com/shurcooL/githubv4/blob/master/githubv4_test.go
	t.Run("successful", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data": {"organization": {"repositories": { "nodes": [{"name":"foo"}]}}}}}`)
		})
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		results, err := getAllOrgRepos(client, "owner1")
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

// Test_getWorkflowRunAttempt tests getting a workflow run attempt
func Test_GetWorkflowRunAttempt(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				GetWorkflowRunByID,
				&testWorkflowRun,
			))
		client := github.NewClient(mockedHTTPClient)
		result, err := getWorkflowRunAttempt(client, testOwner, testRepo, 333)
		assert.NoError(t, err)
		assert.Equal(t, testWorkflowRun.GetRunAttempt(), result)
	})

	t.Run("error", func(t *testing.T) {
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatchHandler(
				GetWorkflowRunByID,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mock.WriteError(
						w,
						http.StatusInternalServerError,
						"test error",
					)
				}),
			))
		client := github.NewClient(mockedHTTPClient)
		result, err := getWorkflowRunAttempt(client, testOwner, testRepo, 333)
		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})
}

func Test_rerunWorkflow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				mock.PostReposActionsRunsRerunByOwnerByRepoByRunId,
				&testWorkflowRun,
			))
		client := github.NewClient(mockedHTTPClient)
		dryRun = false
		err := rerunWorkflow(client, testOwner, testRepo, 333)
		assert.NoError(t, err)
	})

	t.Run("dry run", func(t *testing.T) {
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				mock.PostReposActionsRunsRerunByOwnerByRepoByRunId,
				&testWorkflowRun,
			))
		client := github.NewClient(mockedHTTPClient)
		dryRun = true
		err := rerunWorkflow(client, testOwner, testRepo, 333)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatchHandler(
				mock.PostReposActionsRunsRerunByOwnerByRepoByRunId,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mock.WriteError(
						w,
						http.StatusInternalServerError,
						"test error",
					)
				}),
			))
		client := github.NewClient(mockedHTTPClient)
		dryRun = false
		err := rerunWorkflow(client, testOwner, testRepo, 333)
		assert.Error(t, err)
	})
}

func Test_processPullRequestCheckSuites(t *testing.T) {
	dryRun = true
	t.Run("success", func(t *testing.T) {
		checks := []checkSuite{
			{
				ID:         githubv4.ID("123"),
				Conclusion: "FAILURE",
				WorkflowRun: struct {
					DatabaseId int64
					Workflow   struct {
						Name string
					}
				}{},
			},
		}
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				GetWorkflowRunByID,
				&testWorkflowRun,
			))
		client := github.NewClient(mockedHTTPClient)
		processPullRequestCheckSuites(client, checks, testOwner, testRepo)
	})

	t.Run("too many failures", func(t *testing.T) {
		checks := []checkSuite{
			{
				ID:         githubv4.ID("123"),
				Conclusion: "FAILURE",
				WorkflowRun: struct {
					DatabaseId int64
					Workflow   struct {
						Name string
					}
				}{},
			},
		}
		testWorkflowRun.RunAttempt = github.Int(5)
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				GetWorkflowRunByID,
				&testWorkflowRun,
			))
		client := github.NewClient(mockedHTTPClient)
		processPullRequestCheckSuites(client, checks, testOwner, testRepo)
	})

}

func Test_shouldEnableAutoMerge(t *testing.T) {
	expectRequiredApprovingReviewCount = 1
	expectRequiresStatusChecks = true
	expectRequiresStrictStatusChecks = true
	expectRequiresApprovingReviews = true
	expectRequiredStatusChecks = 1

	t.Run("no default branch match", func(t *testing.T) {
		rules := []branchProtectionRule{
			{
				Pattern:                      "foo",
				RequiredApprovingReviewCount: 1,
				RequiresStatusChecks:         true,
				RequiresStrictStatusChecks:   true,
				RequiresApprovingReviews:     true,
				RequiredStatusChecks: []struct{ Context githubv4.String }{
					{
						Context: "foo",
					},
				},
			},
			{
				Pattern:                      "bar",
				RequiredApprovingReviewCount: 1,
				RequiresStatusChecks:         true,
				RequiresStrictStatusChecks:   true,
				RequiresApprovingReviews:     true,
				RequiredStatusChecks: []struct{ Context githubv4.String }{
					{
						Context: "foo",
					},
				},
			},
		}
		assert.False(t, shouldEnableAutoMerge(rules))
	})

	t.Run("low count of required approving reviews", func(t *testing.T) {
		rules := []branchProtectionRule{
			{
				Pattern:                      githubv4.String(defaultBranch),
				RequiredApprovingReviewCount: 0,
				RequiresStatusChecks:         true,
				RequiresStrictStatusChecks:   true,
				RequiresApprovingReviews:     true,
				RequiredStatusChecks: []struct{ Context githubv4.String }{
					{
						Context: "foo",
					},
				},
			},
		}
		assert.False(t, shouldEnableAutoMerge(rules))
	})

	t.Run("requires status checks is false", func(t *testing.T) {
		rules := []branchProtectionRule{
			{
				Pattern:                      githubv4.String(defaultBranch),
				RequiresStatusChecks:         false,
				RequiredApprovingReviewCount: 1,
				RequiresStrictStatusChecks:   true,
				RequiresApprovingReviews:     true,
				RequiredStatusChecks: []struct{ Context githubv4.String }{
					{
						Context: "foo",
					},
				},
			},
		}
		assert.False(t, shouldEnableAutoMerge(rules))
	})

	t.Run("requires strict status checks is false", func(t *testing.T) {
		rules := []branchProtectionRule{
			{
				Pattern:                      githubv4.String(defaultBranch),
				RequiredApprovingReviewCount: 1,
				RequiresStatusChecks:         true,
				RequiresApprovingReviews:     true,
				RequiredStatusChecks: []struct{ Context githubv4.String }{
					{
						Context: "foo",
					},
				},
				RequiresStrictStatusChecks: false,
			},
		}
		assert.False(t, shouldEnableAutoMerge(rules))
	})

	t.Run("requires approving reviews is false", func(t *testing.T) {
		rules := []branchProtectionRule{
			{
				Pattern:                      githubv4.String(defaultBranch),
				RequiredApprovingReviewCount: 1,
				RequiresStatusChecks:         true,
				RequiresStrictStatusChecks:   true,
				RequiredStatusChecks: []struct{ Context githubv4.String }{
					{
						Context: "foo",
					},
				},
				RequiresApprovingReviews: false,
			},
		}
		assert.False(t, shouldEnableAutoMerge(rules))
	})

	t.Run("low count of required status checks", func(t *testing.T) {
		rules := []branchProtectionRule{
			{
				Pattern:                      githubv4.String(defaultBranch),
				RequiredApprovingReviewCount: 1,
				RequiresStatusChecks:         true,
				RequiresStrictStatusChecks:   true,
				RequiresApprovingReviews:     true,
				RequiredStatusChecks:         []struct{ Context githubv4.String }{},
			},
		}
		assert.False(t, shouldEnableAutoMerge(rules))
	})

	t.Run("should return true", func(t *testing.T) {
		rules := []branchProtectionRule{
			{
				Pattern:                      githubv4.String(defaultBranch),
				RequiredApprovingReviewCount: 1,
				RequiresStatusChecks:         true,
				RequiresStrictStatusChecks:   true,
				RequiresApprovingReviews:     true,
				RequiredStatusChecks: []struct{ Context githubv4.String }{
					{
						Context: "foo",
					},
				},
			},
		}
		assert.True(t, shouldEnableAutoMerge(rules))
	})

}
