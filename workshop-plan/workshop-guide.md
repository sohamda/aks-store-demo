# GitHub Copilot Agentic Workshop — Self-Paced Guide

> **This guide combines the workshop overview and step-by-step instructions into a single document.**
> Whether you're following along live or working through this on your own, everything you need is here.

---

## Workshop at a Glance

|                  |                                                                                 |
| ---------------- | ------------------------------------------------------------------------------- |
| **Duration**     | ~75 min core + take-home exercises                                              |
| **Platform**     | github.com only (no IDE required)                                               |
| **Repo**         | aks-store-demo (8 polyglot microservices, zero Copilot/DevSecOps config)        |
| **Prerequisite** | GitHub account with Copilot Pro/Pro+/Business/Enterprise + Coding Agent enabled |

## What You'll Build

By the end, your fork will have a fully autonomous DevSecOps pipeline:

1. **Custom instructions** that shape all Copilot behavior across the repo
2. **Dependabot** scanning dependencies daily across 6 ecosystems
3. **CodeQL** running SAST on every push with Copilot Autofix
4. **An automated workflow** that audits for vulnerabilities → creates issues → assigns to Copilot → agent opens PRs
5. **A custom IaC Security Agent** that scans Terraform, Bicep, K8s, Helm, and Dockerfiles
6. **An agentic workflow** (markdown, not YAML) that generates a weekly security report

## Files You'll Create

| Module | File                                           | Purpose                                                                             |
| ------ | ---------------------------------------------- | ----------------------------------------------------------------------------------- |
| 2      | `.github/copilot-instructions.md`              | Coding standards + security requirements — shapes all Copilot behavior              |
| 2      | `.github/dependabot.yml`                       | Daily dep scanning (npm, pip, gomod, cargo, actions)                                |
| 2      | `.github/workflows/codeql.yml`                 | CodeQL SAST for JS, Python, Go — feeds Copilot Autofix                              |
| 4      | `.github/workflows/security-audit-autofix.yml` | Automated pipeline: audit deps + find test gaps → create issues → assign to Copilot |
| 6      | `.github/agents/iac-security-agent.md`         | Custom Copilot agent for IaC security scanning                                      |
| 7      | `.github/workflows/security-report.md`         | Agentic workflow: weekly IaC security report in markdown                            |

## Topic Coverage

| Workshop Topic                  |   Module 1    |    Module 2     |   Module 3   |        Module 4        |     Module 5      |     Module 6     |      Module 7      |
| ------------------------------- | :-----------: | :-------------: | :----------: | :--------------------: | :---------------: | :--------------: | :----------------: |
| **github.com agentic features** |   ✅ Agents   |                 |              |                        |  ✅ Code Review   | ✅ Custom Agents |                    |
| **Agentic loops**               |               |                 |              | ✅ scan→issue→agent→CI | ✅ commit history |                  |                    |
| **Agentic workflows**           |               |  ✅ Dependabot  |              | ✅ Automated pipeline  |                   |                  |      ✅ gh-aw      |
| **Custom/agentic SDLC**         |               | ✅ Instructions | ✅ DevSecOps |                        |   ✅ Full loop    |   ✅ IaC agent   |                    |
| **DevSecOps showcase**          | ✅ Assessment |    ✅ CodeQL    |  ✅ Autofix  |  ✅ Auto-remediation   |                   | ✅ IaC scanning  | ✅ Security report |

---

## Setup (Do This First)

### Fork the Repository

1. Go to the source repo on GitHub
2. Click **Fork** → ensure **"Copy all branches"** is checked → **Create fork**

### Enable GitHub Features

Go to `https://github.com/<your-username>/aks-store-demo/settings`:

1. **Settings → Actions → General** →
   - Select "Allow all actions and reusable workflows" → Save
2. **Settings → General → Features** →
   - Enable **Issues** ✅
3. **Settings → Advanced Security** → Enable:
   - Dependabot alerts ✅
   - Secret Protection ✅
4. **Settings → Code security → Code scanning** → Set CodeQL to **default setup**

### Create a PAT_TOKEN Secret

This is required for assigning Copilot to issues via workflows.

1. Go to <https://github.com/settings/tokens> → **Fine-grained tokens** → **Generate new token**
2. **Token name**: `workshop-copilot-assign`
3. **Repository access**: select your fork
4. **Permissions** (all Read & Write): **Actions**, **Contents**, **Issues**, **Pull requests**
5. Click **Generate token** → copy the token
6. Go to your fork → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**
7. Add two secrets with the same token value:
   - Name: `PAT_TOKEN` → paste token
   - Name: `COPILOT_GITHUB_TOKEN` → paste the same token (needed for Module 7)

### Verify Copilot Coding Agent

Go to any issue on your fork → click the **Assignees** dropdown → **"Copilot"** should appear as an option.

> **If Copilot doesn't appear**: Verify your plan is Copilot Pro, Pro+, Business, or Enterprise and that Coding Agent is enabled in your account/org settings.

---

## Module 1: Copilot Agents — Security Assessment (10 min)

**What you'll learn**: How to use the Agents tab on github.com to have Copilot analyze your repository.

**Format**: Demo 3 min → Hands-on 7 min

### Steps

1. Go to your fork on github.com
2. Open the **Agents** tab (or Copilot Chat on the repo page)
3. Ask Copilot to audit the repo:

   ```
   Analyze this repository for security gaps, missing tests, and hardcoded credentials.
   Provide a summary of findings organized by severity.
   ```

4. Review the assessment — note the findings across the polyglot codebase

### What to observe

- Copilot can read and understand code across all 8 microservices (Go, Node.js, Python, Rust, Vue.js)
- It identifies concrete security issues, not just generic advice
- This is the baseline you'll improve throughout the workshop

---

## Module 2: Foundation — Custom Instructions + DevSecOps (12 min)

**What you'll learn**: Set up the three foundational files that enable Copilot to understand your project and automate security scanning.

**Format**: Demo 4 min → Hands-on 8 min

**Catch-up**: If you fall behind, create a PR from `checkpoint/module-2` → `main` and merge it.

### Step 2a: Create `.github/copilot-instructions.md`

On your fork → **"Add file"** → **"Create new file"** → set the path to `.github/copilot-instructions.md` → paste:

```markdown
# Copilot Instructions for aks-store-demo

## Project Context

Polyglot microservices demo for Azure Kubernetes Service with 8 services:

- **order-service** — JavaScript / Node.js (Fastify)
- **makeline-service** — Go (Gin)
- **product-service** — Rust (Actix-web)
- **ai-service** — Python (FastAPI)
- **store-front** — Vue.js 3 (TypeScript / Vite)
- **store-admin** — Vue.js 3 (TypeScript / Vite)
- **virtual-customer** — Rust (load simulator)
- **virtual-worker** — Rust (load simulator)

Data flow: customer → store-front → order-service → RabbitMQ → makeline-service → MongoDB/CosmosDB

## Coding Standards

- **Go**: standard project layout, table-driven tests using `testing` package, `golint` compliant
- **Node.js**: Fastify patterns, TAP test framework, Fastify schema-based request validation
- **Python**: pytest with fixtures, type hints required, input validation on all endpoints
- **Rust**: idiomatic Rust, doc comments on public items, `cargo clippy` clean, `#[cfg(test)]` modules
- **Vue.js**: Composition API, TypeScript strict mode, Vitest for unit tests

## Security Requirements

- No hardcoded secrets — use environment variables or mounted secrets
- Validate all user input at API boundaries
- Pin dependency versions explicitly (no caret ranges for production dependencies)
- All PRs must include tests for new or changed code
- Docker images should use specific base image tags, not `latest`

## PR Standards

- PRs must include a clear description of what changed and why
- Link related issues using `Fixes #N` or `Closes #N`
- Include evidence that tests pass (CI green or manual test output)
```

Commit directly to `main`.

### Step 2b: Create `.github/dependabot.yml`

Create new file → path: `.github/dependabot.yml` → paste:

```yaml
version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/src/order-service"
    schedule:
      interval: "daily"
    labels:
      - "dependencies"
      - "security"
      - "automated"

  - package-ecosystem: "npm"
    directory: "/src/store-front"
    schedule:
      interval: "daily"
    labels:
      - "dependencies"
      - "automated"

  - package-ecosystem: "npm"
    directory: "/src/store-admin"
    schedule:
      interval: "daily"
    labels:
      - "dependencies"
      - "automated"

  - package-ecosystem: "pip"
    directory: "/src/ai-service"
    schedule:
      interval: "daily"
    labels:
      - "dependencies"
      - "automated"

  - package-ecosystem: "gomod"
    directory: "/src/makeline-service"
    schedule:
      interval: "daily"
    labels:
      - "dependencies"
      - "automated"

  - package-ecosystem: "cargo"
    directory: "/src/product-service"
    schedule:
      interval: "daily"
    labels:
      - "dependencies"
      - "automated"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "daily"
    labels:
      - "dependencies"
      - "automated"
```

Commit to `main`.

### Step 2c: Create `.github/workflows/codeql.yml`

Create new file → path: `.github/workflows/codeql.yml` → paste:

```yaml
name: "CodeQL"

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: "30 5 * * 1"

jobs:
  analyze:
    name: Analyze (${{ matrix.language }})
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      packages: read
      actions: read
      contents: read
    strategy:
      fail-fast: false
      matrix:
        include:
          - language: javascript-typescript
            build-mode: none
          - language: python
            build-mode: none
          - language: go
            build-mode: autobuild
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: ${{ matrix.language }}
          build-mode: ${{ matrix.build-mode }}
      - if: matrix.build-mode == 'autobuild'
        name: Autobuild
        uses: github/codeql-action/autobuild@v3
      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
        with:
          category: "/language:${{ matrix.language }}"
```

Commit to `main`.

### Verify

After committing all 3 files, confirm they appear on `main`:

- `.github/copilot-instructions.md`
- `.github/dependabot.yml`
- `.github/workflows/codeql.yml`

> **Note**: Dependabot alerts and CodeQL results take **15–30 minutes** to appear. Module 3 will showcase them.

---

## Module 3: DevSecOps Showcase (10 min)

**What you'll learn**: See the DevSecOps tooling you just configured in action.

**Format**: Facilitator demo (your results are still processing; the facilitator shows a pre-baked fork)

### What to look for

After 15–30 minutes from Module 2 setup (check back later if in a live workshop):

1. **Security tab → Dependabot alerts** — vulnerability findings across npm, pip, Go, Cargo
2. **Security tab → Code scanning** — CodeQL SAST results for JS, Python, Go
3. **Pull requests tab** — Dependabot may have created dependency bump PRs
4. Click a CodeQL alert → look for the **"Copilot Autofix"** button — this is AI-powered auto-remediation

### Key takeaway

With 3 files, you now have a continuous security pipeline: Dependabot scans dependencies, CodeQL scans code, and Copilot can auto-fix findings.

---

## Module 4: Agentic Workflows — Automated Issue Pipeline (20 min)

**What you'll learn**: Create a workflow that automatically scans for vulnerabilities, creates GitHub Issues, and assigns them to Copilot Coding Agent — fully autonomously.

**Format**: Demo 7 min → Hands-on 13 min

**Catch-up**: If you fall behind, create a PR from `checkpoint/module-4` → `main` and merge it.

### How it works

The workflow has 3 parallel jobs:

| Job                   | Scanner     | Target        | What it creates                      |
| --------------------- | ----------- | ------------- | ------------------------------------ |
| `audit-node`          | `npm audit` | order-service | Security issue → assigned to Copilot |
| `audit-python`        | `pip-audit` | ai-service    | Security issue → assigned to Copilot |
| `test-coverage-check` | file search | all services  | Quality issue → assigned to Copilot  |

The workflow uses a two-step pattern for Copilot assignment:

- **Step 1**: Create the issue via `actions/github-script` (uses `GITHUB_TOKEN`)
- **Step 2**: Assign Copilot via `gh api` (uses `PAT_TOKEN` — your personal token with Copilot entitlement)

> `GITHUB_TOKEN` _cannot_ assign Copilot because it lacks your user's Copilot entitlement. The `PAT_TOKEN` carries your identity.

### Steps

**Option A — Merge checkpoint** (recommended):

1. Go to your fork → **Pull requests** → **New pull request**
2. Base: `main` ← Compare: `checkpoint/module-4`
3. Create PR → Merge it

**Option B — Create manually**: View the file at `https://github.com/<your-username>/aks-store-demo/blob/checkpoint/module-4/.github/workflows/security-audit-autofix.yml` and create it via the github.com UI.

### Trigger the workflow

1. Go to your fork's **Actions** tab
2. Click **"Security Audit → Auto-Create Issues → Assign to Copilot"** in the left sidebar
3. Click **"Run workflow"** → select `main` branch → **"Run workflow"**
4. Wait ~2 minutes for the jobs to complete

### Verify

1. Go to **Issues** tab — you should see 2–3 new issues assigned to Copilot with **"Copilot is working"**
2. Go to **Pull Requests** tab — within ~5–8 minutes, Copilot will create PRs from those issues

---

## Module 5: Coding Agent + Code Review — Full Loop (15 min)

**What you'll learn**: Inspect the Copilot agent's work from Module 4, manually create an issue, and use Copilot Code Review.

**Format**: Demo 5 min → Hands-on 10 min

### Step 5a: Inspect Agent PRs from Module 4

1. Go to **Pull Requests** tab and open a Copilot-created PR
2. Check the **commit history** — multiple commits indicate the agentic loop (plan → code → CI → fix → iterate)
3. Review the code — did the agent follow the coding standards from your `copilot-instructions.md`?

### Step 5b: Create a Manual Issue for Copilot

Pick one of these issues (or create your own):

#### Option: Add Input Validation (Node.js — Easy)

1. Go to **Issues** → **New issue**
2. **Title**: `Add input validation to order-service POST /order endpoint`
3. **Body**:

```markdown
## Description

The order-service (`src/order-service/`) accepts orders via POST /order without validating the request body. Add Fastify schema-based request validation.

## Requirements

- Add a JSON schema to the POST /order route for request body validation
- Required fields: `storeId` (string), `customerOrderId` (string), `items` (array)
- `items` array must not be empty
- Each item needs `productId` (integer) and `quantity` (positive integer)
- Return 400 with a descriptive error for invalid payloads
- Add TAP tests for both valid and invalid payloads in `test/`

## Files to modify

- `src/order-service/routes/` — add schema to the order route
- `src/order-service/test/` — add validation tests
```

4. **Labels**: `enhancement`
5. **Assignees**: `copilot`
6. Submit → verify **"Copilot is working"** appears

#### Option: Add Health Endpoints (Go — Medium)

1. Go to **Issues** → **New issue**
2. **Title**: `Add /health and /ready endpoints to makeline-service`
3. **Body**:

```markdown
## Description

The makeline-service (`src/makeline-service/`) needs Kubernetes health check endpoints for liveness and readiness probes.

## Requirements

- `GET /health` — liveness probe: returns HTTP 200 with `{"status": "ok"}` if the process is running
- `GET /ready` — readiness probe: returns HTTP 200 only if the MongoDB/CosmosDB connection is established; returns 503 otherwise
- Add both endpoints to `main.go` using the Gin router
- Add table-driven unit tests in a new `health_test.go` file
- Update `src/makeline-service/README.md` with the new endpoints

## Files to modify

- `src/makeline-service/main.go`
- `src/makeline-service/health_test.go` (new file)
- `src/makeline-service/README.md`
```

4. **Labels**: `enhancement`
5. **Assignees**: `copilot`
6. Submit → verify **"Copilot is working"** appears

### Step 5c: Use Copilot Code Review

1. Open a Copilot-created PR (from Module 4 or 5)
2. Click **"Reviewers"** → type `copilot` → select it
3. Copilot will post review comments on the PR
4. **Reply to a review comment** — ask Copilot to refine something and watch it respond

### What to observe

- The full autonomous loop: issue → agent creates branch → writes code → CI runs → agent iterates → PR ready
- Copilot Code Review catches issues just like a human reviewer
- The agent respects the custom instructions from Module 2

---

## Module 6: IaC Security Agent — Custom Copilot Agent (10 min)

**What you'll learn**: Create a specialized Copilot agent that scans Infrastructure-as-Code files for security misconfigurations.

**Format**: Demo 3 min → Hands-on 7 min

### Step 6a: Create the Agent Definition

On your fork → **"Add file"** → **"Create new file"** → path: `.github/agents/iac-security-agent.md` → paste:

```markdown
---
name: IaCSecurityAgent
description: IaC & Cloud Configuration Guard - Scans Terraform, Bicep, ARM, Kubernetes manifests, and Helm charts for misconfigurations and insecure defaults
---

# IaC & Cloud Configuration Guard Agent

You are the IaC & Cloud Config Guard, an expert in infrastructure-as-code security specializing in Terraform, Bicep/ARM, Kubernetes manifests, and Helm charts. Your mission is to identify misconfigurations and insecure defaults, then propose actionable remediations aligned to cloud security baselines.

## Core Responsibilities

- Detect insecure defaults and misconfigurations in IaC
- Propose minimal, targeted fixes that maintain functionality
- Map findings to security frameworks and compliance controls
- Generate PR-ready remediation plans

## Supported IaC Technologies

| Technology | File Patterns               | Where in this repo               |
| ---------- | --------------------------- | -------------------------------- |
| Terraform  | `*.tf`, `*.tfvars`          | `infra/terraform/`               |
| Bicep      | `*.bicep`                   | `infra/bicep/`                   |
| Kubernetes | `*.yaml` (K8s)              | `kustomize/`, `aks-store-*.yaml` |
| Helm       | `Chart.yaml`, `values.yaml` | `charts/aks-store-demo/`         |
| Dockerfile | `Dockerfile`                | `src/*/Dockerfile`               |

## Security Categories

Organize findings into these security domains:

### 1. Identity & Access Management (IAM)

- Overly permissive RBAC roles
- Missing managed identity configuration
- Hardcoded credentials or secrets
- Wildcard permissions

### 2. Network Security

- Public endpoints without justification
- Missing network segmentation
- Overly permissive security groups (0.0.0.0/0)
- Missing private endpoints

### 3. Data Protection & Encryption

- Encryption at rest disabled
- TLS version below 1.2
- Secrets in plain text

### 4. Container & Workload Security

- Containers running as root
- Privileged containers
- Missing resource limits
- Unpinned image tags (`:latest`)
- Writable root filesystem

### 5. Logging & Monitoring

- Diagnostic settings not configured
- Audit logging disabled
- Missing alerting configuration

## Severity Classification

| Severity | Criteria                                               |
| -------- | ------------------------------------------------------ |
| CRITICAL | Immediate exploitation risk; data breach likely        |
| HIGH     | Significant security gap; elevated attack surface      |
| MEDIUM   | Security best practice violation; defense in depth gap |
| LOW      | Minor hardening opportunity                            |

## Output Format

Generate a structured security report:

### Summary Table

| Category | Critical | High | Medium | Low |
| -------- | -------- | ---- | ------ | --- |

### Detailed Findings

For each finding:

- **File** and **line number**
- **Resource** affected
- **Issue** description
- **Impact** assessment
- **Remediation** with code diff
- **Control mapping** (CIS Azure, NIST 800-53)

## Review Process

1. Discover IaC files in this repository
2. Categorize resources by security domain
3. Apply security checks against cloud security baselines
4. Prioritize findings by severity and blast radius
5. Generate remediations with minimal, targeted fixes
6. Map findings to compliance frameworks

Exit with a complete report. Do not wait for user input unless clarification is needed.
```

Commit to `main`.

### Step 6b: Invoke the Agent via an Issue

1. Go to **Issues** → **New issue**
2. **Title**: `[Security] IaC Security Scan`
3. **Body**:

```
@copilot Use the IaCSecurityAgent to scan this repository's infrastructure code and generate a security report.

Scan:
- `infra/terraform/` (Terraform)
- `infra/bicep/` (Bicep)
- `kustomize/` (Kubernetes manifests)
- `charts/` (Helm charts)
- `src/*/Dockerfile` (Dockerfiles)
- Root `aks-store-*.yaml` files (Kubernetes deployment manifests)
```

4. **Assignees**: `copilot`
5. Submit

### Verify

- Copilot uses the IaCSecurityAgent to generate a comprehensive security report
- Check the resulting PR — it should contain findings across Terraform, Bicep, K8s, Helm, and Dockerfiles organized by security domain

### Key takeaway

Custom Agents are specialized Copilot personalities defined in simple markdown. They can be invoked from issues, PRs, or agentic workflows and bring domain expertise to automated tasks.

---

## Module 7: GitHub Agentic Workflows — gh-aw (10 min)

**What you'll learn**: Define repository automation in **markdown** instead of YAML. An AI agent interprets and executes your instructions in a sandboxed environment.

**Format**: Demo 5 min → Hands-on 5 min

### Key Concepts

- **Markdown, not YAML** — define what you want in natural language
- **AI-powered execution** — a coding agent interprets and acts on your instructions
- **Safe outputs** — write operations (create issue, comment on PR) are pre-approved and sandboxed
- **Guardrails** — read-only by default, network isolation, tool allowlisting

### Option A: Create via github.com Copilot Chat (no CLI needed)

On your fork, open Copilot Chat and enter:

```
Create a workflow for GitHub Agentic Workflows using https://raw.githubusercontent.com/github/gh-aw/main/create.md

The purpose of the workflow is a weekly IaC security report that scans Terraform (infra/terraform/), Bicep (infra/bicep/), Kubernetes manifests (kustomize/, aks-store-*.yaml), Helm charts (charts/), and Dockerfiles (src/*/Dockerfile) for security misconfigurations. Generate findings grouped by IAM, Network Security, Data Protection, Container Security, and Logging. Deliver the report as a GitHub issue.
```

Copilot will create the workflow file and its lock file for you.

### Option B: Create manually + CLI

#### Create the workflow file

On your fork → **"Add file"** → **"Create new file"** → path: `.github/workflows/security-report.md` → paste:

```markdown
---
on:
  schedule: weekly
  workflow_dispatch:
permissions:
  contents: read
  issues: read
  pull-requests: read
safe-outputs:
  create-issue:
    title-prefix: "[security-report] "
    labels: [security, report, automated]
    close-older-issues: true
---

## Weekly IaC Security Report

Scan all infrastructure-as-code files in this repository and generate a comprehensive security report as a GitHub issue.

## What to scan

- Terraform files in `infra/terraform/`
- Bicep files in `infra/bicep/`
- Kubernetes manifests in `kustomize/` and root `aks-store-*.yaml` files
- Helm charts in `charts/`
- Dockerfiles in each `src/*/Dockerfile`

## What to report

- Misconfigurations and insecure defaults
- Findings grouped by: Identity & Access, Network Security, Data Protection, Container Security, Logging
- Severity classification: Critical, High, Medium, Low
- Specific file paths and line numbers for each finding
- Recommended fixes with code snippets
- Compliance mapping to CIS Azure and NIST 800-53 where applicable

## Report format

Use a structured markdown report with:

- Summary table (category × severity counts)
- Detailed findings with file, line, issue, impact, and remediation
- A "Quick Wins" section listing the top 5 easiest fixes with highest impact
```

#### Compile with CLI (if gh-aw is installed)

```bash
# Install the gh-aw extension (once)
gh extension install github/gh-aw

# Compile the markdown into a GitHub Actions workflow
gh aw compile
# This generates .github/workflows/security-report.lock.yml

# Commit both files
git add .github/workflows/security-report.md .github/workflows/security-report.lock.yml
git commit -m "add weekly IaC security report agentic workflow"
git push
```

#### Set up the engine secret

Ensure the `COPILOT_GITHUB_TOKEN` secret is set (from Setup — it uses the same PAT as `PAT_TOKEN`).

| Engine                    | Secret name            | Where to get it                |
| ------------------------- | ---------------------- | ------------------------------ |
| **Copilot** (recommended) | `COPILOT_GITHUB_TOKEN` | Same PAT you created in Setup  |
| Claude                    | `ANTHROPIC_API_KEY`    | https://console.anthropic.com/ |
| Codex                     | `OPENAI_API_KEY`       | https://platform.openai.com/   |

### Trigger the workflow

1. Go to **Actions** tab → **"Weekly IaC Security Report"** → **"Run workflow"**
2. Wait ~2–3 minutes → check **Issues** tab for a new `[security-report]` issue

### Key takeaway

Agentic workflows represent the next generation of CI/CD — define automation in plain language, let an AI agent execute it, with built-in safety guardrails. Combined with the custom agent from Module 6, you get specialized, recurring AI-driven tasks.

---

## Recap: What You Built

Here's the full autonomous pipeline now running on your fork:

```
Dependabot (daily)           CodeQL (on push/PR)         Security Audit Workflow (manual/scheduled)
    │                              │                              │
    ▼                              ▼                              ▼
Dependency bump PRs        Code scanning alerts         npm audit + pip-audit + test coverage
    │                         │                              │
    ▼                         ▼                              ▼
Auto-merge (optional)     Copilot Autofix              Auto-created issues → assigned to Copilot
                                                              │
                                                              ▼
                                                     Copilot Coding Agent
                                                         │         │
                                                         ▼         ▼
                                                    Creates PRs   Code Review
                                                              │
IaC Security Agent ←── Issue/PR trigger                       ▼
Weekly Security Report ←── Agentic Workflow (gh-aw)      Merge to main
```

Everything is automated. Everything runs on github.com. No IDE required.

---

## Take-Home Exercises (Post-Workshop)

### Exercise 1: Review Dependabot & CodeQL Results (Easy)

**When**: Next day (results need time to appear after Module 2 setup).

1. **Security** tab → **Dependabot alerts** — review vulnerability findings
2. **Security** tab → **Code scanning** — review CodeQL findings
3. Click a CodeQL finding → look for **"Copilot Autofix"** → click to see the AI-proposed fix

### Exercise 2: Container Image Scan Workflow (Medium)

Create a Trivy-based container scan workflow. Pre-built in `checkpoint/complete`.

**Quick**: Merge `checkpoint/complete` into `main`.

**Manual**: Create `.github/workflows/docker-scan.yml` with a Trivy scan → auto-create issues → assign to Copilot pattern (see `handout-take-home.md` for full YAML).

### Exercise 3: More Issues for Copilot (Easy–Medium)

Create any of these issues and assign to `copilot`:

| Title                                     | Description                                                                                          |
| ----------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| Add pytest tests for ai-service           | Create `src/ai-service/test_main.py` with pytest tests for health check and /generate endpoints      |
| Add Go unit tests for makeline-service    | Create `src/makeline-service/orders_test.go` with table-driven tests for order processing functions  |
| Create API documentation for all services | Create `docs/api/` with a markdown file per service documenting endpoints, schemas, example requests |
| Add OpenAPI spec for order-service        | Create `src/order-service/openapi.yaml` with a full OpenAPI 3.0 spec                                 |
| Create .github/CODEOWNERS                 | Map each `src/` subdirectory to appropriate reviewers/teams                                          |

### Exercise 4: Branch Protection Rules (Easy)

1. Go to **Settings** → **Branches** → **Add branch protection rule**
2. Branch name pattern: `main`
3. Enable: require PR, require 1 approval, require status checks, require Copilot code review
4. Now every PR — human or Copilot — must pass CI and get reviewed before merge

### Exercise 5: Extend the Pipeline (Advanced)

Add more scanners to `security-audit-autofix.yml`: `govulncheck` for Go, `cargo-audit` for Rust, SBOM generation (see `handout-take-home.md` for details).

---

## Troubleshooting

| What                             | Expected                    | If it fails                                                                       |
| -------------------------------- | --------------------------- | --------------------------------------------------------------------------------- |
| Dependabot alerts appear         | Within 15–30 min            | Check Settings → Advanced Security → Dependabot is enabled                        |
| CodeQL runs                      | Triggered by merge to main  | Check Actions tab for the CodeQL workflow run                                     |
| Security audit workflow triggers | Manual trigger works        | Check Actions → workflow → "Run workflow" button visible                          |
| Issues auto-created              | 2–3 issues in Issues tab    | Check workflow logs for errors (permissions, jq parsing)                          |
| `copilot` assignee works         | "Copilot is working" badge  | Needs PAT_TOKEN secret + Copilot Pro/Pro+/Business/Enterprise                     |
| Copilot creates PRs              | PRs within 5–8 min of issue | Check the issue page for agent status updates                                     |
| Copilot Code Review              | Review comments on PR       | Add `copilot` as reviewer manually via Reviewers dropdown                         |
| IaC Security Agent               | Structured security report  | Verify `.github/agents/iac-security-agent.md` is committed and agent name matches |
| gh-aw workflow runs              | Issue created with report   | Check engine secret (`COPILOT_GITHUB_TOKEN`), verify `.lock.yml` file exists      |

### The #1 thing to validate

**Assigning an issue to `copilot` triggers the Coding Agent.** If that doesn't work:

- Verify your Copilot plan is **Pro, Pro+, Business, or Enterprise**
- Verify the `PAT_TOKEN` secret is set (the workflow uses it to assign Copilot via the REST API)
- Try assigning Copilot manually from the issue UI — if it doesn't appear, Coding Agent isn't enabled for your account/org

---

## Resources

- [GitHub Agentic Workflows Docs](https://github.github.com/gh-aw/)
- [gh-aw Quick Start](https://github.github.com/gh-aw/setup/quick-start/)
- [gh-aw Example Gallery](https://github.github.com/gh-aw/#gallery)
