# Buildctl Usage Guide for CI/CD Pipelines

## Overview

`buildctl` is a CLI tool for tracking build progress and uploading artifacts in CI/CD pipelines. It sends events to a Build Assistant backend to track build status, logs, and artifacts.

## Basic Workflow Pattern

The typical CI/CD integration follows this pattern:

1. **Start**: Send `started` event at the beginning of the build
2. **Progress**: Send `in_progress` events after major build steps with log messages
3. **End**: Send `success` or `failed` event based on build outcome

## Commands

### Event Command

Sends build status events to the backend.

#### Syntax

```bash
buildctl event \
  --token="<AUTH_TOKEN>" \
  --backend="<BACKEND_URL>" \
  --project="<PROJECT_NAME>" \
  --commit="<COMMIT_HASH>" \
  --branch="<BRANCH_NAME>" \
  --status="<STATUS>" \
  [--log="<LOG_MESSAGE>"]
```

#### Required Parameters

- `--token`: Authentication token (Bearer token)
- `--backend`: Backend API URL (e.g., `https://build-assistant.example.com`)
- `--project`: Project name identifier
- `--commit`: Full commit hash (SHA)
- `--branch`: Branch name (e.g., `main`, `develop`)
- `--status`: Build status (see Status Types below)

#### Optional Parameters

- `--log`: Log message describing the current step (useful for `in_progress` events)

#### Status Types

The `--status` parameter accepts one of the following values:

- `started`: Build has started (use at the beginning of pipeline)
- `in_progress`: Build is in progress (use after major steps)
- `success`: Build completed successfully (use at the end on success)
- `failed`: Build failed (use at the end on failure)

#### Example Usage

```bash
# Start of build
buildctl event \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --commit="$COMMIT_SHA" \
  --branch="$BRANCH_NAME" \
  --status="started"

# After completing a build step
buildctl event \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --commit="$COMMIT_SHA" \
  --branch="$BRANCH_NAME" \
  --status="in_progress" \
  --log="Built backend Docker image"

# On success
buildctl event \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --commit="$COMMIT_SHA" \
  --branch="$BRANCH_NAME" \
  --status="success"

# On failure
buildctl event \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --commit="$COMMIT_SHA" \
  --branch="$BRANCH_NAME" \
  --status="failed"
```

### Artifact Upload Command

Uploads build artifacts to S3 storage via presigned URLs.

#### Syntax

```bash
buildctl artifact upload \
  --token="<AUTH_TOKEN>" \
  --backend="<BACKEND_URL>" \
  --project="<PROJECT_NAME>" \
  --branch="<BRANCH_NAME>" \
  --commit="<COMMIT_HASH>" \
  --file="<FILE_PATH>"
```

#### Required Parameters

- `--token`: Authentication token (Bearer token)
- `--backend`: Backend API URL
- `--project`: Project name identifier
- `--branch`: Branch name
- `--commit`: Full commit hash (SHA)
- `--file`: Path to the file to upload

#### Example Usage

```bash
buildctl artifact upload \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --branch="$BRANCH_NAME" \
  --commit="$COMMIT_SHA" \
  --file="./dist/app.tar.gz"
```

## Complete CI/CD Integration Example

Based on Gitea Actions workflow pattern:

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      # 1. START: Notify build started
      - name: Notify build started
        run: |
          buildctl event \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --status="started"
        continue-on-error: true

      # 2. Build steps...
      - name: Build application
        run: |
          # Your build commands here
          make build

      # 3. PROGRESS: Notify after major steps
      - name: Notify build progress
        run: |
          buildctl event \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --log="Application built successfully" \
            --status="in_progress"
        continue-on-error: true

      # 4. Upload artifacts (optional)
      - name: Upload artifact
        run: |
          buildctl artifact upload \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --branch="${{ github.ref_name }}" \
            --commit="${{ github.sha }}" \
            --file="./dist/app.tar.gz"
        continue-on-error: true

      # 5. END: Notify success
      - name: Notify build success
        if: success()
        run: |
          buildctl event \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --status="success"
        continue-on-error: true

      # 6. END: Notify failure
      - name: Notify build failed
        if: failure()
        run: |
          buildctl event \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --status="failed"
        continue-on-error: true
```

## Implementation Notes

### Error Handling

- Always use `continue-on-error: true` in CI/CD steps to prevent build failures if the notification service is unavailable

### Environment Variables

Common CI/CD platform variables:

- **Gitea Actions**: `${{ github.sha }}`, `${{ github.ref_name }}`
- **GitHub Actions**: `${{ github.sha }}`, `${{ github.ref }}`
- **GitLab CI**: `$CI_COMMIT_SHA`, `$CI_COMMIT_REF_NAME`
- **Jenkins**: `$GIT_COMMIT`, `$GIT_BRANCH`

### Best Practices

1. **Always start with `started`**: Send this event immediately after checkout
2. **Use `in_progress` for milestones**: Send after completing major build steps with descriptive log messages
3. **Always end with status**: Use `success` or `failed` in conditional steps (`if: success()` / `if: failure()`)
4. **Use descriptive logs**: The `--log` parameter helps track progress through the build pipeline
5. **Handle errors gracefully**: Use `continue-on-error: true` to prevent notification failures from breaking builds
