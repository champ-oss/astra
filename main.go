package main

import (
	"context"
	"encoding/base64"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"github.com/shurcooL/githubv4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type checkSuite struct {
	ID          githubv4.ID
	Conclusion  githubv4.String
	WorkflowRun struct {
		DatabaseId int64
		Workflow   struct {
			Name string
		}
	}
}

type pullRequest struct {
	ID               githubv4.ID
	Number           githubv4.Int
	Title            githubv4.String
	Url              githubv4.String
	AutoMergeRequest struct {
		EnabledAt githubv4.String
	}
	Author struct {
		Login githubv4.String
	}
	Commits struct {
		Nodes []struct {
			Commit struct {
				CheckSuites struct {
					Nodes []checkSuite
				} `graphql:"checkSuites(last:100)"`
			}
		}
	} `graphql:"commits(last:1)"`
}

type branchProtectionRule struct {
	RequiredApprovingReviewCount githubv4.Int
	Pattern                      githubv4.String
	RequiresStatusChecks         githubv4.Boolean
	RequiresStrictStatusChecks   githubv4.Boolean
	RequiresApprovingReviews     githubv4.Boolean
	RequiredStatusChecks         []struct {
		Context githubv4.String
	}
}

type repoPullRequestsQuery struct {
	Repository struct {
		ID           githubv4.ID
		PullRequests struct {
			Nodes []pullRequest
		} `graphql:"pullRequests(states:OPEN, last:20)"`
		BranchProtectionRules struct {
			Nodes []branchProtectionRule
		} `graphql:"branchProtectionRules(last:20)"`
	} `graphql:"repository(owner:$repositoryOwner, name:$repositoryName)"`
}

type repositoryName struct {
	Name githubv4.String
}

type organizationRepos struct {
	Organization struct {
		Repositories struct {
			Nodes    []repositoryName
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"repositories(first:100, after:$repositoriesCursor)"`
	} `graphql:"organization(login:$login)"`
}

var debug bool
var dryRun bool
var repoPrefixes []string
var actors []string
var defaultBranch string
var waitSecondsBetweenRequests int
var maxRunAttempts int

func main() {
	setLogging("INPUT_DEBUG")
	setDryRun("INPUT_DRY_RUN")

	owner := os.Getenv("INPUT_OWNER")
	appId := getEnvInt64("INPUT_APP_ID")
	installationId := getEnvInt64("INPUT_INSTALLATION_ID")
	pem := getEnvBase64("INPUT_PEM")
	repoPrefixes = getEnvStringList("INPUT_REPO_PREFIXES")
	actors = getEnvStringList("INPUT_ACTORS")
	defaultBranch = getEnvString("INPUT_DEFAULT_BRANCH")
	waitSecondsBetweenRequests = getEnvInt("INPUT_WAIT_SECONDS_BETWEEN_REQUESTS")
	maxRunAttempts = getEnvInt("INPUT_MAX_RUN_ATTEMPTS")

	client, err := getClient(appId, installationId, pem)
	if err != nil {
		panic(err)
	}

	clientV4, err := getClientV4(appId, installationId, pem)
	if err != nil {
		panic(err)
	}

	allRepos, err := getAllOrgRepos(clientV4, owner)
	if err != nil {
		panic(err)
	}
	processRepos(filterRepoNames(allRepos), client, clientV4, owner)
}

// setLogging sets up logging configuration
func setLogging(envKey string) {
	log.SetLevel(log.DebugLevel)
	debug = getEnvBool(envKey, true)
	if !debug {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		DisableQuote:  true,
		FullTimestamp: true,
	})
}

// setDryRun sets up dry run configuration
func setDryRun(envKey string) {
	dryRun = getEnvBool(envKey, true)
	log.Warningf("%s is set to: %v", envKey, dryRun)
}

// processRepos process each repo
func processRepos(repos []string, client *github.Client, clientV4 *githubv4.Client, owner string) {
	for _, repo := range repos {
		log.Infof("Started processing repo: %s", repo)
		results, err := runRepoPullRequestsQuery(clientV4, owner, repo)
		if err != nil {
			log.Errorf("error running runRepoPullRequestsQuery: %s", err)
			continue
		}

		enableAutoMerge := shouldEnableAutoMerge(results.Repository.BranchProtectionRules.Nodes)
		if !enableAutoMerge {
			log.Warningf("Will not enable auto-merge for %s repo due to lack of branch protections", repo)
		}

		for _, pr := range results.Repository.PullRequests.Nodes {
			processPullRequest(client, clientV4, pr, owner, repo, enableAutoMerge)
		}

		log.Infof("Done processing repo: %s", repo)
	}
}

// processPullRequest processes a single pull request
func processPullRequest(client *github.Client, clientV4 *githubv4.Client, pr pullRequest, owner, repo string, enableAutoMerge bool) {
	log.Infof("Processing pull request: ID:%s Number:%d Author:%s Title:%s\n", pr.ID, pr.Number, pr.Author.Login, pr.Title)
	log.Infof("Pull request URL: %s", pr.Url)

	if !containsString(actors, string(pr.Author.Login)) {
		log.Infof("Skipping pull request by author: %s", pr.Author.Login)
		return
	}

	if pr.AutoMergeRequest.EnabledAt == "" && enableAutoMerge {
		log.Warningf("Enabling auto merge for PR: %s\n", pr.Title)
		err := enablePullRequestAutoMerge(clientV4, pr.ID)
		if err != nil {
			log.Errorf("error running enablePullRequestAutoMerge: %s", err)
		}
	}

	if len(pr.Commits.Nodes) > 0 {
		processPullRequestCheckSuites(client, pr.Commits.Nodes[0].Commit.CheckSuites.Nodes, owner, repo)
	}
}

func processPullRequestCheckSuites(client *github.Client, checkSuites []checkSuite, owner, repo string) {
	for _, check := range checkSuites {
		log.Debugf("Processing check suite: name:%s conclusion:%s id:%d", check.WorkflowRun.Workflow.Name, check.Conclusion, check.WorkflowRun.DatabaseId)
		if check.Conclusion == "FAILURE" || check.Conclusion == "CANCELLED" {

			attempts, err := getWorkflowRunAttempt(client, owner, repo, check.WorkflowRun.DatabaseId)
			if err != nil {
				log.Errorf("error running getWorkflowRunAttempt: %s", err)
			}
			log.Infof("Workflow run attempts: %d", attempts)
			if attempts >= maxRunAttempts {
				log.Warningf("Not restarting workflow due to too many failed attempts")
				continue
			}

			log.Warningf("Restarting %s workflow for repo: %s", check.WorkflowRun.Workflow.Name, repo)
			err = rerunWorkflow(client, owner, repo, check.WorkflowRun.DatabaseId)
			if err != nil {
				log.Errorf("error running rerunWorkflow: %s", err)
			}
		}
	}
}

// getEnvString parses an environment variable as a string
func getEnvString(envKey string) string {
	log.Debugf("Parsing env %s=%s as string", envKey, os.Getenv(envKey))

	envValue := os.Getenv(envKey)
	if envValue == "" {
		log.Panicf("env variable %s must be set", envKey)
	}
	return envValue
}

// getEnvBool parses an environment variable as a boolean value
func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); strings.ToLower(value) == "true" {
		return true
	}
	if value := os.Getenv(key); strings.ToLower(value) == "false" {
		return false
	}
	return fallback
}

// getEnvInt parses an environment variable as int
func getEnvInt(envKey string) int {
	log.Debugf("Parsing env %s=%s as int", envKey, os.Getenv(envKey))

	envValue := os.Getenv(envKey)
	if envValue == "" {
		log.Panicf("env variable %s must be set", envKey)
	}

	parsed, err := strconv.Atoi(envValue)
	if err != nil {
		log.Panicf("unable to parse %s as int", envKey)
	}
	return parsed
}

// getEnvInt64 parses an environment variable as int64
func getEnvInt64(envKey string) int64 {
	log.Debugf("Parsing env %s=%s as int64", envKey, os.Getenv(envKey))

	envValue := os.Getenv(envKey)
	if envValue == "" {
		log.Panicf("env variable %s must be set", envKey)
	}

	parsed, err := strconv.ParseInt(envValue, 10, 64)
	if err != nil {
		log.Panicf("unable to parse %s as int64", envKey)
	}
	return parsed
}

// getEnvBase64 parses an environment variable as base64 bytes
func getEnvBase64(envKey string) []byte {
	log.Debugf("Parsing env %s as base64", envKey)

	envValue := os.Getenv(envKey)
	if envValue == "" {
		log.Panicf("env variable %s must be set", envKey)
	}

	decoded, err := base64.StdEncoding.DecodeString(envValue)
	if err != nil {
		log.Panicf("unable to decode %s as base64", envKey)
	}
	return decoded
}

// getEnvStringList parses an environment variable as a list of strings
func getEnvStringList(envKey string) []string {
	log.Debugf("Parsing env %s as a list of strings", envKey)

	envValue := os.Getenv(envKey)
	if envValue == "" {
		log.Panicf("env variable %s must be set", envKey)
	}

	var items []string
	for _, f := range strings.Split(envValue, "\n") {
		if strings.TrimSpace(f) != "" {
			items = append(items, f)
		}
	}
	return items
}

// getClientV4 creates a mew githubv4 graphql client instance
func getClientV4(appID, installationID int64, pem []byte) (*githubv4.Client, error) {
	log.Debugf("Creating GitHub v4 client using appID:%d installationID:%d", appID, installationID)

	transport, err := ghinstallation.New(http.DefaultTransport, appID, installationID, pem)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{Transport: transport}
	return githubv4.NewClient(httpClient), nil
}

// getClient creates a new github REST client instance
func getClient(appID, installationID int64, pem []byte) (*github.Client, error) {
	log.Debugf("Creating GitHub v4 client using appID:%d installationID:%d", appID, installationID)

	transport, err := ghinstallation.New(http.DefaultTransport, appID, installationID, pem)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{Transport: transport}
	return github.NewClient(httpClient), nil
}

// getAllOrgRepos fetches a list of all repositories for the organization
func getAllOrgRepos(client *githubv4.Client, organization string) ([]repositoryName, error) {
	log.Infof("Getting all repos for organization: %s", organization)

	// Create variable from the query struct to hold results
	queryInstance := &organizationRepos{}

	variables := map[string]interface{}{
		"login":              githubv4.String(organization),
		"repositoriesCursor": (*githubv4.String)(nil), // get first page.
	}

	var allRepositories []repositoryName
	for {
		err := client.Query(context.Background(), &queryInstance, variables)
		time.Sleep(time.Duration(waitSecondsBetweenRequests) * time.Second)
		if err != nil {
			return nil, err
		}
		allRepositories = append(allRepositories, queryInstance.Organization.Repositories.Nodes...)
		log.Debugf("retrieved %d repos", len(allRepositories))
		if !queryInstance.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["repositoriesCursor"] = queryInstance.Organization.Repositories.PageInfo.EndCursor
	}
	return allRepositories, nil
}

// filterRepoNames returns a filtered list of repositories
func filterRepoNames(repos []repositoryName) []string {
	log.Infof("Filtering %d repositories to process based on repoPrefixes list", len(repos))
	var filtered []string
	for _, repo := range repos {
		for _, allowedPrefix := range repoPrefixes {
			if strings.HasPrefix(string(repo.Name), allowedPrefix) {
				filtered = append(filtered, string(repo.Name))
			}
		}
	}
	log.Infof("Filtered to %d repos", len(filtered))
	return filtered
}

// runRepoPullRequestsQuery executes the query and returns results
func runRepoPullRequestsQuery(client *githubv4.Client, owner, repo string) (*repoPullRequestsQuery, error) {

	// Create variable from the query struct to hold results
	queryInstance := &repoPullRequestsQuery{}

	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(repo),
	}

	err := client.Query(context.Background(), queryInstance, variables)
	time.Sleep(time.Duration(waitSecondsBetweenRequests) * time.Second)
	if err != nil {
		return nil, err
	}
	return queryInstance, nil
}

// enablePullRequestAutoMerge enables the Auto-Merge feature on a pull request
func enablePullRequestAutoMerge(client *githubv4.Client, pullRequestID githubv4.ID) error {
	var mutate struct {
		EnablePullRequestAutoMerge struct {
			ClientMutationId githubv4.String
		} `graphql:"enablePullRequestAutoMerge(input: $input)"`
	}

	input := githubv4.EnablePullRequestAutoMergeInput{
		PullRequestID: pullRequestID,
	}

	if dryRun {
		log.Warningf("Dry-run is enabled. Will not execute enablePullRequestAutoMerge")
		return nil
	}

	err := client.Mutate(context.Background(), &mutate, input, nil)
	time.Sleep(time.Duration(waitSecondsBetweenRequests) * time.Second)
	if err != nil {
		return err
	}
	return nil
}

// containsString checks if the container contains the given string
func containsString(container []string, name string) bool {
	for _, v := range container {
		if v == name {
			return true
		}
	}
	return false
}

// getWorkflowRunAttempt returns the run attempts for a workflow
func getWorkflowRunAttempt(client *github.Client, owner, repo string, runID int64) (int, error) {
	run, _, err := client.Actions.GetWorkflowRunByID(context.Background(), owner, repo, runID)
	time.Sleep(time.Duration(waitSecondsBetweenRequests) * time.Second)
	if err != nil {
		return 0, err
	}
	return run.GetRunAttempt(), nil
}

// rerunWorkflow requests the given workflow to be run again
func rerunWorkflow(client *github.Client, owner, repo string, runID int64) error {
	if dryRun {
		log.Warningf("Dry-run is enabled. Will not execute rerunWorkflow")
		return nil
	}
	_, err := client.Actions.RerunWorkflowByID(context.Background(), owner, repo, runID)
	time.Sleep(time.Duration(waitSecondsBetweenRequests) * time.Second)
	if err != nil {
		return err
	}
	return nil
}

// shouldEnableAutoMerge evaluates if it is safe to enable auto-merge of a pull request based on branch protection rules
func shouldEnableAutoMerge(rules []branchProtectionRule) bool {
	log.Debugf("validating branch protection rules for %s branch", defaultBranch)
	for _, rule := range rules {

		// Only evaluate rules for the default branch (ex: main or master)
		if string(rule.Pattern) != defaultBranch {
			continue
		}
		if rule.RequiredApprovingReviewCount < 1 {
			log.Debugf("RequiredApprovingReviewCount is < 1")
			return false
		}
		if !rule.RequiresStatusChecks {
			log.Debugf("RequiresStatusChecks is false")
			return false
		}
		if !rule.RequiresStrictStatusChecks {
			log.Debugf("RequiresStrictStatusChecks is false")
			return false
		}
		if !rule.RequiresApprovingReviews {
			log.Debugf("RequiresApprovingReviews is false")
			return false
		}
		if len(rule.RequiredStatusChecks) < 1 {
			log.Debugf("RequiredStatusChecks < 1")
			return false
		}
		log.Debugf("successfully validated branch protection rules for %s branch", defaultBranch)
		return true
	}
	log.Debugf("no branch protection rules found for %s branch", defaultBranch)
	return false
}
