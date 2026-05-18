# git-notifier
Have granular control over what GitHub events you are notified of. No more broad repository events calling your entire team or having to subscribe to all issues creation.

## How does it work?
This service runs in an isolated Docker container in the user's machine. That way it can use the credentials from the user and all information will never leave their local machine.

With a simple POST REST request, the user can subscribe to a single PR by providing the URL in the body and this service will start the check process:
1. requests to GitHub current state of the PR
2. stores that state in memory
3. Sleeps for 60s
4. requests to Github the latest state of that same PR
5. Compares latest with previous state
6. If changed, sends a message to a given Slack channel
7. If it didn't change, sleep for 60s

## Usage (Recommended)
1. Build the image with Docker
``` bash
docker build -t gn .
```
2. Run the github-notifier image passing the required environmental variables
``` bash
docker run \
  -e SLACK_WEBHOOK_URL="MY_SLACK_TOKEN" \
  -e GITHUB_TOKEN="MY_GH_TOKEN" \
  -e USER_EMAIL="change@me.com" \
  -e USERNAME="change-me" \
  -p 8001:8001 -d gn:latest
```
3. Subscribe to a PR
``` bash
curl -X POST http://localhost:8001/api/v1/subscribe \
-d '{"pr_url": "https://github.com/my-org/my-repo/pull/1"}'
```

## TODOs
- [x] CLI command for subscribe
- [x] /unsubscribe
- [x] CLI command for unsubscribe
- [ ] Refactor client types and tests
- [x] Ignore my own comments and commits
- [x] Include approvals
- [x] Add mutex lock for storage
- [ ] Add lint and test workflows
- [x] Create Dockerfile
- [ ] Add Docker build and deploy workflow
- [ ] Add k8s manifests
- [ ] Use better logger
