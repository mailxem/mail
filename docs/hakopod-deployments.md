# Hakopod release deployments

The backend, frontend, MCP, and payments release workflows deploy their published container image to the matching service in the existing Hakopod application. Redis is not changed by these workflows.

Configure these repository or organization settings before a release:

- Variable `HAKOPOD_API_URL`: the public Hakopod origin or API URL.
- Variable `HAKOPOD_APPLICATION_ID`: the existing application containing services named `backend`, `frontend`, `mcp`, and `payments`.
- Secret `HAKOPOD_API_TOKEN`: a scoped machine token with `deployments:read` and `deployments:write` for that application.

Each release publishes its image first and passes the exact `image@sha256:...` reference to `hakopod/deploy@v1`. The reusable deployment waits up to ten minutes for a successful rollout. Deployments for the application share one non-canceling concurrency group with the maximum FIFO queue, so up to GitHub's 100 pending jobs are retained and serialized. Component release queues use the same policy before publication.

When one push changes both backend and frontend release inputs, the frontend workflow waits for the same-commit backend workflow to finish successfully before deploying. This makes database migrations and backend routes available before the corresponding frontend. Client-only releases do not wait for an unrelated backend run. A manually dispatched frontend release deploys independently because it does not imply a paired backend release; when manually releasing paired changes, dispatch and verify the backend release before dispatching the frontend.

GitHub releases created with `GITHUB_TOKEN` produce release events, but those events do not trigger additional Actions workflows. Deployment jobs therefore run directly after successful image publication and release creation. A failed build, manifest publication, GitHub release, backend ordering check, or Hakopod rollout prevents the dependent deployment from being reported successful.

MCP pull requests only validate. Default-branch pushes and manual dispatches from the default branch publish and deploy MCP; `mcp-v*` tags publish versioned images after validation but do not deploy production.

GitHub added `concurrency.queue: max` after the currently pinned Actionlint schema. Until Actionlint releases that schema, validate workflows with the narrow compatibility exclusion below and run the focused contract test, which asserts the lock still uses the documented maximum FIFO queue and does not cancel in-progress deployments:

```sh
actionlint -ignore 'unexpected key "queue" for "concurrency" section' .github/workflows/*.yml
python3 scripts/test_hakopod_workflows.py
```

See [GitHub's concurrency documentation](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency) for the current `queue` behavior.
