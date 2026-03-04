# CI/CD Integration

Integrate UniFi Site Manager CLI into your CI/CD pipelines.

## GitHub Actions

### Basic Workflow

**File**: `.github/workflows/unifi-monitor.yml`

```yaml
name: UniFi Network Monitor

on:
  schedule:
    - cron: '*/15 * * * *'  # Every 15 minutes
  workflow_dispatch:

env:
  USM_API_KEY: ${{ secrets.USM_API_KEY }}
  USM_SITE_ID: ${{ vars.USM_SITE_ID }}

jobs:
  health-check:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Install USM CLI
        run: |
          curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
          chmod +x usm
          sudo mv usm /usr/local/bin/
      
      - name: Check Site Health
        run: |
          usm sites health $USM_SITE_ID --output json > health.json
          cat health.json
      
      - name: Verify All Devices Online
        run: |
          OFFLINE=$(usm sites health $USM_SITE_ID --output json | jq '.devices.offline')
          if [ "$OFFLINE" -gt 0 ]; then
            echo "❌ $OFFLINE devices are offline"
            exit 1
          fi
          echo "✅ All devices online"
      
      - name: Upload Health Report
        uses: actions/upload-artifact@v4
        with:
          name: health-report
          path: health.json
```

### Deployment Workflow

```yaml
name: UniFi Configuration Deployment

on:
  push:
    branches: [ main ]
    paths:
      - 'unifi-configs/**'
  workflow_dispatch:

env:
  USM_API_KEY: ${{ secrets.USM_API_KEY }}
  USM_SITE_ID: ${{ vars.USM_SITE_ID }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install USM CLI
        run: |
          curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
          chmod +x usm
          sudo mv usm /usr/local/bin/
      
      - name: Pre-deployment Health Check
        run: |
          usm sites health $USM_SITE_ID --output json | jq -e '.status == "healthy"'
          echo "✅ Site is healthy, proceeding with deployment"
      
      - name: Apply WLAN Configurations
        run: |
          for config in unifi-configs/wlans/*.json; do
            echo "Applying $config"
            # Parse config and apply via usm
            NAME=$(jq -r '.name' "$config")
            SSID=$(jq -r '.ssid' "$config")
            usm wlans create $USM_SITE_ID "$NAME" "$SSID" \
              --password "$(jq -r '.password' "$config")" \
              --security "$(jq -r '.security' "$config")" \
              --vlan "$(jq -r '.vlan' "$config")" || true
          done
      
      - name: Post-deployment Verification
        run: |
          sleep 30  # Wait for changes to apply
          usm sites health $USM_SITE_ID --output json | jq -e '.devices.offline == 0'
          echo "✅ Deployment successful"
      
      - name: Notify Slack
        if: always()
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "UniFi Deployment: ${{ job.status }}"
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Pull Request Validation

```yaml
name: PR Validation

on:
  pull_request:
    paths:
      - 'unifi-configs/**'

env:
  USM_API_KEY: ${{ secrets.USM_API_KEY }}

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install USM CLI
        run: |
          curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
          chmod +x usm
          sudo mv usm /usr/local/bin/
      
      - name: Validate Configuration Files
        run: |
          # Check JSON validity
          for file in unifi-configs/**/*.json; do
            echo "Validating $file"
            jq empty "$file" || exit 1
          done
      
      - name: Dry Run Changes
        run: |
          # Test connectivity
          usm whoami
          echo "✅ Configuration valid and API accessible"
```

## GitLab CI

### Basic Pipeline

**File**: `.gitlab-ci.yml`

```yaml
stages:
  - validate
  - deploy
  - verify

variables:
  USM_API_KEY: $USM_API_KEY
  USM_SITE_ID: $USM_SITE_ID

before_script:
  - apk add --no-cache curl jq
  - curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
  - chmod +x usm
  - mv usm /usr/local/bin/

validate:
  stage: validate
  script:
    - usm whoami
    - usm sites health $USM_SITE_ID
  rules:
    - if: $CI_PIPELINE_SOURCE == "schedule"

health_check:
  stage: validate
  script:
    - |
      HEALTH=$(usm sites health $USM_SITE_ID --output json)
      OFFLINE=$(echo $HEALTH | jq '.devices.offline')
      if [ "$OFFLINE" -gt 0 ]; then
        echo "Warning: $OFFLINE devices offline"
      fi
  rules:
    - if: $CI_PIPELINE_SOURCE == "schedule"
  
deploy_wlans:
  stage: deploy
  script:
    - |
      for config in configs/wlans/*.json; do
        NAME=$(jq -r '.name' "$config")
        SSID=$(jq -r '.ssid' "$config")
        usm wlans create $USM_SITE_ID "$NAME" "$SSID" \
          --password "$(jq -r '.password' "$config")" \
          --security "$(jq -r '.security' "$config")"
      done
  rules:
    - if: $CI_COMMIT_BRANCH == "main"

verify:
  stage: verify
  script:
    - sleep 30
    - usm sites health $USM_SITE_ID --output json | jq -e '.status == "healthy"'
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
```

## Jenkins

### Jenkinsfile

```groovy
pipeline {
    agent any
    
    environment {
        USM_API_KEY = credentials('usm-api-key')
        USM_SITE_ID = '60abcdef1234567890abcdef'
    }
    
    stages {
        stage('Install CLI') {
            steps {
                sh '''
                    curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
                    chmod +x usm
                    sudo mv usm /usr/local/bin/
                '''
            }
        }
        
        stage('Health Check') {
            steps {
                sh 'usm sites health $USM_SITE_ID --output json'
            }
        }
        
        stage('Deploy Configuration') {
            when {
                branch 'main'
            }
            steps {
                sh '''
                    # Apply configurations
                    for config in configs/*.json; do
                        echo "Processing $config"
                        # Apply config
                    done
                '''
            }
        }
        
        stage('Verify') {
            steps {
                sh '''
                    sleep 30
                    usm sites health $USM_SITE_ID --output json | jq -e '.devices.offline == 0'
                '''
            }
        }
    }
    
    post {
        always {
            script {
                if (currentBuild.result == 'SUCCESS') {
                    slackSend(color: 'good', message: "UniFi deployment successful: ${env.JOB_NAME} ${env.BUILD_NUMBER}")
                } else {
                    slackSend(color: 'danger', message: "UniFi deployment failed: ${env.JOB_NAME} ${env.BUILD_NUMBER}")
                }
            }
        }
    }
}
```

## Azure DevOps

### Azure Pipelines

**File**: `azure-pipelines.yml`

```yaml
trigger:
  branches:
    include:
      - main
  paths:
    include:
      - unifi-configs/**

pr:
  branches:
    include:
      - main

variables:
  USM_API_KEY: $(usmApiKey)
  USM_SITE_ID: $(usmSiteId)

stages:
- stage: Validate
  jobs:
  - job: HealthCheck
    pool:
      vmImage: 'ubuntu-latest'
    steps:
    - script: |
        curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
        chmod +x usm
        sudo mv usm /usr/local/bin/
        usm sites health $(USM_SITE_ID)
      displayName: 'Health Check'

- stage: Deploy
  condition: and(succeeded(), eq(variables['Build.SourceBranch'], 'refs/heads/main'))
  jobs:
  - deployment: DeployToProduction
    pool:
      vmImage: 'ubuntu-latest'
    environment: 'production'
    strategy:
      runOnce:
        deploy:
          steps:
          - script: |
              # Apply configurations
              echo "Deploying to site: $(USM_SITE_ID)"
            displayName: 'Deploy Configurations'
          
          - script: |
              sleep 30
              usm sites health $(USM_SITE_ID) --output json | jq -e '.status == "healthy"'
            displayName: 'Verify Deployment'
```

## CircleCI

### config.yml

```yaml
version: 2.1

executors:
  default:
    docker:
      - image: cimg/base:stable

jobs:
  health-check:
    executor: default
    steps:
      - run:
          name: Install USM
          command: |
            curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
            chmod +x usm
            sudo mv usm /usr/local/bin/
      
      - run:
          name: Check Health
          command: |
            usm sites health $USM_SITE_ID --output json | jq '.status'
    
  deploy:
    executor: default
    steps:
      - checkout
      - run:
          name: Deploy Configurations
          command: |
            # Deploy logic here
            echo "Deploying..."

workflows:
  version: 2
  monitor-and-deploy:
    jobs:
      - health-check:
          filters:
            branches:
              only: main
      - deploy:
          requires:
            - health-check
          filters:
            branches:
              only: main
```

## Travis CI

### .travis.yml

```yaml
language: minimal

env:
  global:
    - USM_API_KEY="${USM_API_KEY}"
    - USM_SITE_ID="60abcdef1234567890abcdef"

before_install:
  - curl -L https://github.com/dl-alexandre/UniFi-Site-Manager-CLI/releases/latest/download/usm-linux-amd64 -o usm
  - chmod +x usm
  - sudo mv usm /usr/local/bin/

script:
  - usm sites health $USM_SITE_ID --output json
  - usm devices list $USM_SITE_ID --output json | jq '.devices | length'

branches:
  only:
    - main
```

## Configuration as Code

### Example Configurations

**File**: `unifi-configs/wlans/corporate.json`

```json
{
  "name": "Corporate WiFi",
  "ssid": "CORP-5G",
  "security": "wpapsk",
  "password": "CorporateSecure123!",
  "vlan": 10,
  "band": "both",
  "wpa3": true,
  "hide_ssid": false
}
```

**File**: `unifi-configs/wlans/guest.json`

```json
{
  "name": "Guest WiFi",
  "ssid": "Guest-Network",
  "security": "wpapsk",
  "password": "GuestAccess2024",
  "vlan": 20,
  "band": "both",
  "is_guest": true,
  "bandwidth_limit": {
    "download": 50,
    "upload": 25
  }
}
```

## Security Best Practices

### Secret Management

**GitHub Actions**:
```yaml
env:
  USM_API_KEY: ${{ secrets.USM_API_KEY }}
```

**GitLab CI**:
```yaml
variables:
  USM_API_KEY: $USM_API_KEY  # From CI/CD variables
```

**Jenkins**:
```groovy
environment {
    USM_API_KEY = credentials('usm-api-key')
}
```

### IP Whitelisting

```yaml
# GitHub Actions - Restrict to specific runners
jobs:
  deploy:
    runs-on: [self-hosted, unifi-deployer]
    # Only run on authorized runners
```

### Audit Logging

```bash
#!/bin/bash
# audit-deploy.sh

echo "[$(date)] Deployment by $GITHUB_ACTOR" >> deploy-audit.log
usm sites health $USM_SITE_ID >> deploy-audit.log
```

## Monitoring CI/CD

### Deployment Notifications

```yaml
- name: Notify on Failure
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {
        "text": "🚨 UniFi Deployment Failed",
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "*Site:* $USM_SITE_ID\n*Status:* Failed\n*Commit:* $GITHUB_SHA"
            }
          }
        ]
      }
```

### Deployment Dashboard

Create a status page showing:
- Last deployment time
- Current site health
- Device status
- Active alerts

## Rollback Procedures

### Automated Rollback

```bash
#!/bin/bash
# rollback.sh

SITE_ID="$1"
BACKUP_FILE="$2"

# Verify current state is bad
if ! usm sites health $SITE_ID --output json | jq -e '.status == "healthy"'; then
  echo "Site unhealthy, initiating rollback..."
  
  # Restore from backup
  # Apply previous configuration
  
  # Verify rollback
  sleep 30
  usm sites health $SITE_ID --output json | jq -e '.status == "healthy"'
fi
```

---

For more examples, see:
- [Basic Usage](../basic/)
- [Automation Scripts](../automation/)
- [Monitoring Setup](../monitoring/)
