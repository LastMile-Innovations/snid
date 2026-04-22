# CI/CD Integration Guide

This guide covers integrating the SNID Benchmarking Platform with CI/CD pipelines.

## GitHub Actions

### Nightly Benchmark Runs

Create `.github/workflows/nightly-benchmarks.yml`:

```yaml
name: Nightly Benchmarks

on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight UTC
  workflow_dispatch:  # Allow manual trigger

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build benchmark image
        run: |
          docker build -f benchmarks/Dockerfile -t snid-benchmarks:latest .

      - name: Run benchmarks
        run: |
          docker run --rm \
            -v ${{ github.workspace }}/benchmarks/results:/app/results \
            -e BENCH_MODE=cli \
            -e BENCH_SUITES=all \
            snid-benchmarks:latest

      - name: Upload results
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: benchmarks/results/*.json
          retention-days: 90

      - name: Check for regressions
        run: |
          python benchmarks/check_regressions.py \
            --current benchmarks/results/full_run_*.json \
            --baseline benchmarks/baselines/latest.json \
            --threshold 10
```

### PR Benchmark Validation

Add to your existing PR workflow:

```yaml
name: PR Benchmark Check

on:
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build and run benchmarks
        run: |
          docker build -f benchmarks/Dockerfile -t snid-benchmarks:latest .
          docker run --rm \
            -v ${{ github.workspace }}/benchmarks/results:/app/results \
            -e BENCH_MODE=cli \
            -e BENCH_SUITES=go,rust,python \
            snid-benchmarks:latest

      - name: Compare with baseline
        run: |
          python benchmarks/compare_benchmarks.py \
            --current benchmarks/results/performance_*.json \
            --baseline benchmarks/baselines/main.json

      - name: Comment on PR
        if: failure()
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '⚠️ Benchmark regression detected. Check the workflow logs for details.'
            })
```

## Railway CI/CD

### Automatic Deployment on Push

```yaml
# .github/workflows/railway-deploy.yml
name: Deploy to Railway

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Railway CLI
        run: npm install -g @railway/cli

      - name: Login to Railway
        run: railway login --token ${{ secrets.RAILWAY_TOKEN }}

      - name: Deploy
        run: railway up

      - name: Run benchmarks
        run: |
          railway run --env BENCH_MODE=cli --env BENCH_SUITES=all \
            python benchmarks/runner.py all
```

### Trigger Benchmark from CI

```yaml
- name: Trigger Railway benchmark
  run: |
    curl -X POST \
      -H "Authorization: Bearer ${{ secrets.RAILWAY_TOKEN }}" \
      -H "Content-Type: application/json" \
      -d '{"service_id": "${{ secrets.RAILWAY_SERVICE_ID }}"}' \
      https://api.railway.app/v1/projects/${{ secrets.RAILWAY_PROJECT_ID }}/services/${{ secrets.RAILWAY_SERVICE_ID }}/triggers
```

## GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - benchmark

benchmark:
  stage: benchmark
  image: docker:24
  services:
    - docker:24-dind
  script:
    - docker build -f benchmarks/Dockerfile -t snid-benchmarks:latest .
    - docker run --rm
      -v $CI_PROJECT_DIR/benchmarks/results:/app/results
      -e BENCH_MODE=cli
      -e BENCH_SUITES=all
      snid-benchmarks:latest
  artifacts:
    paths:
      - benchmarks/results/*.json
    expire_in: 30 days
  only:
    - schedules
    - main
```

## CircleCI

```yaml
# .circleci/config.yml
version: 2.1

jobs:
  benchmark:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      - setup_remote_docker
      - run:
          name: Build benchmark image
          command: docker build -f benchmarks/Dockerfile -t snid-benchmarks:latest .
      - run:
          name: Run benchmarks
          command: |
            docker run --rm \
              -v $(pwd)/benchmarks/results:/app/results \
              -e BENCH_MODE=cli \
              -e BENCH_SUITES=all \
              snid-benchmarks:latest
      - store_artifacts:
          path: benchmarks/results
          destination: benchmark-results

workflows:
  nightly:
    triggers:
      - schedule:
          cron: "0 0 * * *"
          filters:
            branches:
              only: main
    jobs:
      - benchmark
```

## Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    triggers {
        cron('0 0 * * *')  // Daily at midnight
    }
    stages {
        stage('Build') {
            steps {
                sh 'docker build -f benchmarks/Dockerfile -t snid-benchmarks:latest .'
            }
        }
        stage('Benchmark') {
            steps {
                sh '''
                    docker run --rm \
                        -v ${WORKSPACE}/benchmarks/results:/app/results \
                        -e BENCH_MODE=cli \
                        -e BENCH_SUITES=all \
                        snid-benchmarks:latest
                '''
            }
        }
        stage('Archive') {
            steps {
                archiveArtifacts artifacts: 'benchmarks/results/*.json', fingerprint: true
            }
        }
    }
    post {
        always {
            sh 'docker system prune -f'
        }
    }
}
```

## Azure Pipelines

```yaml
# azure-pipelines.yml
trigger:
- main

schedules:
- cron: "0 0 * * *"
  displayName: Daily midnight benchmark
  branches:
    include:
    - main

pool:
  vmImage: 'ubuntu-latest'

steps:
- task: Docker@2
  inputs:
    command: 'build'
    dockerFile: 'benchmarks/Dockerfile'
    tags: 'snid-benchmarks:latest'

- script: |
    docker run --rm \
      -v $(System.DefaultWorkingDirectory)/benchmarks/results:/app/results \
      -e BENCH_MODE=cli \
      -e BENCH_SUITES=all \
      snid-benchmarks:latest
  displayName: 'Run benchmarks'

- task: PublishBuildArtifacts@1
  inputs:
    PathtoPublish: 'benchmarks/results'
    ArtifactName: 'benchmark-results'
    publishLocation: 'Container'
```

## Regression Alerting

### Slack Notification

```yaml
- name: Notify on regression
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {
        "text": "⚠️ Benchmark regression detected in ${{ github.repository }}",
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "Benchmark regression detected in ${{ github.repository }}\nCommit: ${{ github.sha }}\nWorkflow: ${{ github.workflow }}"
            }
          }
        ]
      }
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Email Notification

```yaml
- name: Send email on regression
  if: failure()
  uses: dawidd6/action-send-mail@v3
  with:
    server_address: smtp.gmail.com
    server_port: 465
    username: ${{ secrets.EMAIL_USERNAME }}
    password: ${{ secrets.EMAIL_PASSWORD }}
    subject: "Benchmark Regression - ${{ github.repository }}"
    to: team@example.com
    from: ci@example.com
    body: |
      Benchmark regression detected.
      Repository: ${{ github.repository }}
      Commit: ${{ github.sha }}
      Workflow: ${{ github.workflow }}
```

## Baseline Management

### Update Baseline on Success

```yaml
- name: Update baseline
  if: success()
  run: |
    cp benchmarks/results/performance_*.json benchmarks/baselines/latest.json
    git config user.name "CI Bot"
    git config user.email "ci@example.com"
    git add benchmarks/baselines/latest.json
    git commit -m "Update benchmark baseline"
    git push
```

## Performance Trends

### Store Results in Database

```yaml
- name: Store results
  run: |
    python benchmarks/store_results.py \
      --file benchmarks/results/performance_*.json \
      --database-url ${{ secrets.DATABASE_URL }}
```

### Generate Trend Report

```yaml
- name: Generate trend report
  run: |
    python benchmarks/generate_trend_report.py \
      --database-url ${{ secrets.DATABASE_URL }} \
      --output benchmarks/trend_report.html

- name: Upload trend report
  uses: actions/upload-artifact@v4
  with:
    name: trend-report
    path: benchmarks/trend_report.html
```

## Secrets Required

Configure these secrets in your CI platform:

- `RAILWAY_TOKEN` - Railway API token
- `RAILWAY_PROJECT_ID` - Railway project ID
- `RAILWAY_SERVICE_ID` - Railway service ID
- `SLACK_WEBHOOK_URL` - Slack webhook for notifications
- `EMAIL_USERNAME` / `EMAIL_PASSWORD` - Email credentials
- `DATABASE_URL` - Database for storing benchmark results

## Best Practices

1. **Run benchmarks nightly** - Don't block PRs on full benchmark suite
2. **Use artifacts** - Store results for later analysis
3. **Set baselines** - Compare against known-good runs
4. **Alert on regression** - Notify team when performance degrades
5. **Keep history** - Retain results for trend analysis
6. **Parallelize** - Run different language benchmarks in parallel jobs
7. **Cache Docker layers** - Speed up builds with layer caching
