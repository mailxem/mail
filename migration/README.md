# Monorepo migration and CI

## Source import

All eight former submodules are ordinary tracked directories. `component-sources.json` records the exact published source commit for each import. This is a snapshot migration: the earlier component commit histories remain in their original repositories and are linked by that manifest. They have not been rewritten, deleted, or archived. One monorepo commit now identifies the matching frontend and backend source used by the installer.

The import uses published trees, not the contents of a developer's working directory. Local `.env` files, private keys, dependencies, and uncommitted work are excluded. The backend's previously tracked `bin/server` executable is intentionally omitted; CI rebuilds binaries from source. Original frontend/backend workflow files are relocated/adapted as described below.

Go modules and JavaScript dependency files remain component-local. This migration does not rename module import paths, combine unrelated lockfiles, or change SDK package identities. Other components' nested `.github/` files are retained as historical reference and are not active GitHub workflows. This migration activates frontend and backend CI at the root.

## Existing developer checkouts

The safest migration is a fresh clone after this PR merges:

```bash
git clone https://github.com/mailxem/mail.git xem-monorepo
cd xem-monorepo
make help
```

Before retiring an old submodule checkout, commit or separately back up each component's local changes, untracked files, and environment configuration. Do not run a forced submodule deinitialization, clean, or reset over ongoing work. The new checkout no longer needs `git submodule update`.

## Workflow mapping

| Previous component workflow | Root workflow | Behavior |
| --- | --- | --- |
| `server/.github/workflows/ci.yml` | `backend-ci.yml` | Go vet, race tests, coverage, and binary build with `server/` working directory. |
| `server/.github/workflows/managed-sending.yml` | `backend-integration.yml` | Real PostgreSQL/SMTP tests and CloudFormation lint, scoped to relevant backend paths. |
| `server/.github/workflows/release.yml` | `backend-release.yml` + `component-release.yml` | Independent amd64/arm64 image publication followed by a combined manifest and backend release. |
| `client/.github/workflows/docker.yml` | `frontend-release.yml` + `component-release.yml` | Native amd64/arm64 app builds and a frontend release. |
| Frontend validation | `frontend-ci.yml` | TypeScript, Jest, and Next.js build on frontend changes. |
| `server/.github/workflows/deploy-eks.yml` | `backend-eks.yml` | Manual, default-branch-only, gated by `EKS_DEPLOY_ENABLED=true` and the `production` environment. No deployment on ordinary pushes. |
| `server/.github/workflows/code-review.yml` | `backend-review.yml` | Optional Cori integration, gated by `CORI_REVIEW_ENABLED=true`; same-repo PRs only. |
| Parent Swarm starter | `swarm.yml` | Builds the monorepo app/backend, deploys disposable services, exercises signup/login/upload and repeat installation, then removes the stack. |
| Workflow/release validation | `monorepo.yml` | Actionlint, release-version tests, and a check preventing new Git submodules. |

GitHub only runs workflows under the repository root `.github/workflows/`. Path filters keep a frontend-only change from running backend tests or publishing a new backend image. Changes to the shared release workflow intentionally affect both release pipelines.

## Releases

The repository's default branch is currently `undefined`. CI and publication use that branch; update the workflow filters together if the branch is renamed.

Frontend and backend get separate Git tags: `frontend-v0.1.0`, `backend-v0.1.0`, then independent patch increments. Older unprefixed component tags are never mixed into this sequence. Release notes come from the changed component's Git history and do not depend on an external AI service. The release is created only after image publication succeeds. Per-component concurrency serializes version selection/publication, while architecture jobs publish independently as soon as they finish.

Existing Docker Hub repositories remain:

- `theboringhumane/xemapp`
- `theboringhumane/xemgo`

Each receives version and full-SHA tags, plus the existing `sudo` and `latest` aliases. Architecture tags include `-amd64`/`-arm64`. A combined manifest is published after both architectures succeed. Release dispatch from a feature branch skips publication; use the default branch.

## Repository configuration

Configure in **mailxem/mail**, not only the former component repositories:

| Name | Kind | Purpose |
| --- | --- | --- |
| `DOCKER_HUB_TOKEN` | Actions secret | Publish both existing Docker Hub image repositories. |
| `DOCKER_HUB_USERNAME` | Actions variable | Defaults to `theboringhumane`. |
| `NEXT_PUBLIC_API_URL` | Actions variable | Frontend's build-time public API URL, including `/api/v1`. |
| `NEXT_PUBLIC_PAYWALL_URL` | Actions variable | Frontend's build-time payment API URL. |
| `CORI_REVIEW_ENABLED` | Actions variable | Optional; enable only after repairing/configuring the GitHub App. |
| `CORI_APP_ID`, `CORI_APP_PRIVATE_KEY`, `OPENAI_API_KEY`, `OPENAI_BASE_URL` | Actions secrets | Needed only for the optional reviewer. |
| `EKS_DEPLOY_ENABLED` | Actions variable | Optional legacy EKS deployment gate. |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | Actions secrets | Needed only for the legacy EKS workflow. Review its target settings and protect the `production` environment first. |

Secrets cannot be read back through GitHub's API. Never print them in Actions logs. For authorized transfers, encrypt on the source runner with the destination repository's GitHub public key, then submit only the sealed ciphertext to the destination Secrets API. Alternatively, generate a new scoped credential and save it directly in the destination.

## Cutover

1. Review and merge this monorepo PR, with the Docker Hub credential and public build variables configured.
2. Confirm the root validation/release workflows and published image manifests.
3. Change branch protection to the new root workflow checks. Path-filtered checks are absent on unrelated changes; do not require them unconditionally without a checks-aggregation policy.
4. After successful publication from this repo, disable the former repositories' publishing workflows so two pipelines cannot race to update `sudo`/`latest`. Until cutover, avoid merging publication-triggering changes into both locations.
5. Point deployment/build integrations at `mailxem/mail`, using `server/` and `client/` as build contexts. Keeping the Docker Hub image names preserves image-based deployment references; external source-build webhooks still need an explicit repository/context change.

The original repositories are left intact as history and rollback references. The migration does not archive them or alter production deployment integrations automatically.
