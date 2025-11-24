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

Uploads build artifacts (files, archives, binaries) to S3 storage via presigned URLs and registers them in the build system.

**What it does:**
1. Requests a presigned S3 upload URL from the backend
2. Uploads the file directly to S3
3. Confirms the upload with the backend, creating an artifact record
4. Returns a public download URL

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
- `--file`: Path to the file to upload (can be any file type: .tar.gz, .zip, .bin, .apk, etc.)

#### Example Usage

```bash
# Upload a tarball
buildctl artifact upload \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --branch="$BRANCH_NAME" \
  --commit="$COMMIT_SHA" \
  --file="./dist/app.tar.gz"

# Upload a binary
buildctl artifact upload \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --branch="$BRANCH_NAME" \
  --commit="$COMMIT_SHA" \
  --file="./build/myapp"

# Upload a mobile app
buildctl artifact upload \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --branch="$BRANCH_NAME" \
  --commit="$COMMIT_SHA" \
  --file="./app/build/outputs/apk/release/app-release.apk"
```

### Container Register Command

Registers container images (Docker, OCI) in the build system. Can optionally upload a tar archive of the image to S3.

**What it does:**
1. (Optional) If `--file` is provided: uploads the image tar file to S3
2. Registers the container image metadata (name, tag, digest) in the build system
3. Associates the image with the build for tracking and notifications

**Use cases:**
- Track Docker images pushed to registries
- Store image tar files as backup artifacts
- Monitor which images were built from which commits

#### Syntax

```bash
buildctl container register \
  --token="<AUTH_TOKEN>" \
  --backend="<BACKEND_URL>" \
  --project="<PROJECT_NAME>" \
  --branch="<BRANCH_NAME>" \
  --commit="<COMMIT_HASH>" \
  --image="<IMAGE_NAME_WITH_TAG>" \
  --digest="<IMAGE_DIGEST>" \
  [--file="<IMAGE_TAR_PATH>"]
```

#### Required Parameters

- `--token`: Authentication token (Bearer token)
- `--backend`: Backend API URL
- `--project`: Project name identifier
- `--branch`: Branch name
- `--commit`: Full commit hash (SHA)
- `--image`: Full image name with tag (e.g., `registry.example.com/myapp:v1.0.0`)
- `--digest`: Image digest/hash (e.g., `sha256:abc123...`)

#### Optional Parameters

- `--file`: Path to image tar file (if you want to store the image as an artifact)

#### Example Usage

```bash
# Register image after pushing to registry (most common)
buildctl container register \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --branch="$BRANCH_NAME" \
  --commit="$COMMIT_SHA" \
  --image="registry.example.com/myapp:v1.0.0" \
  --digest="sha256:abc123def456..."

# Register image AND upload tar file
buildctl container register \
  --token="$BCTL_TOKEN" \
  --backend="$BCTL_BACKEND" \
  --project="$BCTL_PROJECT_NAME" \
  --branch="$BRANCH_NAME" \
  --commit="$COMMIT_SHA" \
  --image="registry.example.com/myapp:v1.0.0" \
  --digest="sha256:abc123def456..." \
  --file="./myapp-image.tar"
```

#### Getting Image Digest

Different ways to obtain the image digest:

```bash
# After building with Docker
DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' myimage:tag | cut -d'@' -f2)

# After pushing to registry
docker push myimage:tag
DIGEST=$(docker inspect myimage:tag --format='{{index .RepoDigests 0}}' | cut -d'@' -f2)

# Using docker buildx
docker buildx build --push -t myimage:tag . --iidfile imageid.txt
DIGEST=$(cat imageid.txt)
```

## Complete CI/CD Integration Examples

### Example 1: Simple Build with Artifact Upload

This example shows a basic build pipeline that compiles code and uploads the resulting artifact.

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

      # 2. Build application
      - name: Build application
        run: |
          # Your build commands here
          make build
          tar -czf dist/app.tar.gz -C build .

      # 3. PROGRESS: Notify after build
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

      # 4. Upload artifact
      - name: Upload build artifact
        if: success()
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

### Example 2: Docker Build and Push with Container Registration

This example shows how to build a Docker image, push it to a registry, and register it with the build system.

```yaml
jobs:
  docker-build:
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

      # 2. Login to registry
      - name: Login to Docker Registry
        uses: docker/login-action@v3
        with:
          registry: registry.example.com
          username: ${{ secrets.REGISTRY_USERNAME }}
          password: ${{ secrets.REGISTRY_PASSWORD }}

      # 3. Build and push Docker image
      - name: Build and push Docker image
        run: |
          IMAGE_NAME="registry.example.com/${{ vars.BCTL_PROJECT_NAME }}:${{ github.sha }}"

          # Build the image
          docker build -t "$IMAGE_NAME" .

          # Push to registry
          docker push "$IMAGE_NAME"

          # Get the digest
          DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' "$IMAGE_NAME" | cut -d'@' -f2)

          # Save for later steps
          echo "IMAGE_NAME=$IMAGE_NAME" >> $GITHUB_ENV
          echo "IMAGE_DIGEST=$DIGEST" >> $GITHUB_ENV

      # 4. PROGRESS: Notify after Docker build
      - name: Notify Docker build progress
        run: |
          buildctl event \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --log="Docker image built and pushed: ${{ env.IMAGE_NAME }}" \
            --status="in_progress"
        continue-on-error: true

      # 5. Register container image
      - name: Register container image
        if: success()
        run: |
          buildctl container register \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --branch="${{ github.ref_name }}" \
            --commit="${{ github.sha }}" \
            --image="${{ env.IMAGE_NAME }}" \
            --digest="${{ env.IMAGE_DIGEST }}"
        continue-on-error: true

      # 6. END: Notify success
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

      # 7. END: Notify failure
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

### Example 3: Multi-Step Build with Multiple Artifacts

This example shows a complex pipeline with multiple build steps and artifact uploads.

```yaml
jobs:
  full-build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      # 1. START
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

      # 2. Build backend
      - name: Build backend
        run: |
          cd backend
          go build -o ../dist/backend-server ./cmd/server
          cd ..

      - name: Notify backend built
        run: |
          buildctl event \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --log="Backend server built" \
            --status="in_progress"
        continue-on-error: true

      # 3. Upload backend binary
      - name: Upload backend artifact
        if: success()
        run: |
          buildctl artifact upload \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --branch="${{ github.ref_name }}" \
            --commit="${{ github.sha }}" \
            --file="./dist/backend-server"
        continue-on-error: true

      # 4. Build frontend
      - name: Build frontend
        run: |
          cd frontend
          npm ci
          npm run build
          tar -czf ../dist/frontend.tar.gz -C dist .
          cd ..

      - name: Notify frontend built
        run: |
          buildctl event \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --log="Frontend built" \
            --status="in_progress"
        continue-on-error: true

      # 5. Upload frontend artifact
      - name: Upload frontend artifact
        if: success()
        run: |
          buildctl artifact upload \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --branch="${{ github.ref_name }}" \
            --commit="${{ github.sha }}" \
            --file="./dist/frontend.tar.gz"
        continue-on-error: true

      # 6. Build and push Docker images
      - name: Login to registry
        uses: docker/login-action@v3
        with:
          registry: registry.example.com
          username: ${{ secrets.REGISTRY_USERNAME }}
          password: ${{ secrets.REGISTRY_PASSWORD }}

      - name: Build and push backend image
        run: |
          IMAGE_NAME="registry.example.com/${{ vars.BCTL_PROJECT_NAME }}-backend:${{ github.sha }}"
          docker build -t "$IMAGE_NAME" -f backend/Dockerfile .
          docker push "$IMAGE_NAME"
          DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' "$IMAGE_NAME" | cut -d'@' -f2)

          buildctl container register \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --branch="${{ github.ref_name }}" \
            --commit="${{ github.sha }}" \
            --image="$IMAGE_NAME" \
            --digest="$DIGEST"
        continue-on-error: true

      - name: Build and push frontend image
        run: |
          IMAGE_NAME="registry.example.com/${{ vars.BCTL_PROJECT_NAME }}-frontend:${{ github.sha }}"
          docker build -t "$IMAGE_NAME" -f frontend/Dockerfile .
          docker push "$IMAGE_NAME"
          DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' "$IMAGE_NAME" | cut -d'@' -f2)

          buildctl container register \
            --token="${{ vars.BCTL_TOKEN }}" \
            --backend="${{ vars.BCTL_BACKEND }}" \
            --project="${{ vars.BCTL_PROJECT_NAME }}" \
            --branch="${{ github.ref_name }}" \
            --commit="${{ github.sha }}" \
            --image="$IMAGE_NAME" \
            --digest="$DIGEST"
        continue-on-error: true

      # 7. END: Success
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

      # 8. END: Failure
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
6. **Upload artifacts conditionally**: Use `if: success()` to only upload artifacts when build succeeds
7. **Register containers after push**: Always push to registry first, then register with build system

## Guide for AI Agents

This section is specifically designed to help AI coding assistants integrate `buildctl` into CI/CD pipelines.

### Quick Decision Tree

When modifying a CI/CD pipeline, follow this decision tree:

1. **Does the pipeline build/compile code?**
   - YES → Add `buildctl event` calls with `started`, `in_progress`, and `success`/`failed` statuses
   - NO → Consider if this is a deployment-only pipeline (see below)

2. **Does the pipeline produce downloadable files (binaries, archives, APKs)?**
   - YES → Add `buildctl artifact upload` after successful build
   - NO → Skip artifact upload

3. **Does the pipeline build and push Docker/container images?**
   - YES → Add `buildctl container register` after pushing to registry
   - NO → Skip container registration

### Required Environment Variables

Before adding `buildctl` commands, ensure these variables exist in the CI/CD configuration:

```yaml
vars:
  BCTL_TOKEN: "authentication-token"
  BCTL_BACKEND: "https://build-assistant.example.com"
  BCTL_PROJECT_NAME: "project-name"
```

### Integration Template

Use this template structure when adding `buildctl` to any pipeline:

```yaml
# At the very beginning (after checkout)
- name: Notify build started
  run: buildctl event --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --commit="${{ github.sha }}" --branch="${{ github.ref_name }}" --status="started"
  continue-on-error: true

# After each major build step
- name: Notify [step description]
  run: buildctl event --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --commit="${{ github.sha }}" --branch="${{ github.ref_name }}" --log="[Descriptive message]" --status="in_progress"
  continue-on-error: true

# After successful artifact creation
- name: Upload [artifact name]
  if: success()
  run: buildctl artifact upload --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --branch="${{ github.ref_name }}" --commit="${{ github.sha }}" --file="[path/to/file]"
  continue-on-error: true

# After pushing container to registry
- name: Register container
  if: success()
  run: |
    DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' "$IMAGE_NAME" | cut -d'@' -f2)
    buildctl container register --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --branch="${{ github.ref_name }}" --commit="${{ github.sha }}" --image="$IMAGE_NAME" --digest="$DIGEST"
  continue-on-error: true

# At the very end (success case)
- name: Notify build success
  if: success()
  run: buildctl event --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --commit="${{ github.sha }}" --branch="${{ github.ref_name }}" --status="success"
  continue-on-error: true

# At the very end (failure case)
- name: Notify build failed
  if: failure()
  run: buildctl event --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --commit="${{ github.sha }}" --branch="${{ github.ref_name }}" --status="failed"
  continue-on-error: true
```

### Common Patterns for AI Agents

#### Pattern 1: Identify Build Outputs

When analyzing a pipeline, look for these patterns to identify what artifacts to upload:

```yaml
# Compiled binaries
go build -o ./dist/myapp
# → Upload: ./dist/myapp

# Compressed archives
tar -czf release.tar.gz ./build
# → Upload: release.tar.gz

# npm/node builds
npm run build  # usually outputs to dist/ or build/
# → Upload: dist/ or build/ as tarball

# Mobile apps
./gradlew assembleRelease  # outputs APK
# → Upload: app/build/outputs/apk/release/*.apk

# Python packages
python setup.py sdist bdist_wheel
# → Upload: dist/*.whl or dist/*.tar.gz
```

#### Pattern 2: Docker Image Registration

When you see Docker build/push commands, add container registration:

```yaml
# BEFORE (existing code)
- name: Build and push
  run: |
    docker build -t myimage:latest .
    docker push myimage:latest

# AFTER (with buildctl)
- name: Build and push
  run: |
    IMAGE_NAME="registry.example.com/myimage:${{ github.sha }}"
    docker build -t "$IMAGE_NAME" .
    docker push "$IMAGE_NAME"
    DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' "$IMAGE_NAME" | cut -d'@' -f2)

    buildctl container register \
      --token="${{ vars.BCTL_TOKEN }}" \
      --backend="${{ vars.BCTL_BACKEND }}" \
      --project="${{ vars.BCTL_PROJECT_NAME }}" \
      --branch="${{ github.ref_name }}" \
      --commit="${{ github.sha }}" \
      --image="$IMAGE_NAME" \
      --digest="$DIGEST"
  continue-on-error: true
```

#### Pattern 3: Multi-Stage Pipelines

For pipelines with multiple stages/jobs, only add `buildctl` to the build job, not deployment jobs:

```yaml
jobs:
  build:  # ✓ Add buildctl here
    steps:
      - name: Build
        run: make build
      # ... add buildctl commands

  test:  # ✗ Skip buildctl here (unless it's a test build)
    steps:
      - name: Run tests
        run: npm test

  deploy:  # ✗ Skip buildctl here (deployment, not build)
    steps:
      - name: Deploy
        run: kubectl apply -f k8s/
```

### Minimal Working Example

If user asks for "minimal integration", use this:

```yaml
steps:
  - uses: actions/checkout@v4

  - name: Notify started
    run: buildctl event --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --commit="${{ github.sha }}" --branch="${{ github.ref_name }}" --status="started"
    continue-on-error: true

  # ... existing build steps ...

  - name: Notify success
    if: success()
    run: buildctl event --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --commit="${{ github.sha }}" --branch="${{ github.ref_name }}" --status="success"
    continue-on-error: true

  - name: Notify failure
    if: failure()
    run: buildctl event --token="${{ vars.BCTL_TOKEN }}" --backend="${{ vars.BCTL_BACKEND }}" --project="${{ vars.BCTL_PROJECT_NAME }}" --commit="${{ github.sha }}" --branch="${{ github.ref_name }}" --status="failed"
    continue-on-error: true
```

### Important Notes for AI Agents

1. **Always preserve existing functionality**: Add `buildctl` commands as NEW steps, never replace or modify existing build logic
2. **Use `continue-on-error: true`**: This ensures build notification failures don't break the actual build
3. **Conditional execution**: Use `if: success()` for success/artifact steps and `if: failure()` for failure steps
4. **Variable substitution**: Use `${{ vars.VAR_NAME }}` for GitHub/Gitea Actions, adjust for other CI systems
5. **Commit hash**: Most CI systems provide this as an environment variable (e.g., `${{ github.sha }}`)
6. **Branch name**: Usually available as `${{ github.ref_name }}` or similar
7. **Token security**: Never hardcode tokens, always use secrets/variables from the CI system
