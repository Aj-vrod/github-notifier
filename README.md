# git-notifier
Have granular control over what GitHub events you are notified of. No more broad repository events calling your entire team or having to subscribe to all issues creation.

## How does it work?
This service runs in an isolated Docker container in the user's machine. That way it can use the credentials from the user and all information will never leave their local machine.

With a simple POST REST request, the user can subscribe to a single PR by providing the URL in the body and this service will start the check process:
1. requests to GitHub current state of the PR
2. stores that state in Redis memory
3. Sleeps for 30s
4. requests to Github the latest state of that same PR
5. Compares latest with previous state
6. If changed, sends a message to a given Slack channel
7. If it didn't change, sleep for 30s

