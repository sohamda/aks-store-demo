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

## Required safe-output behavior

- You must emit at least one safe output in every run.
- If you produce a report, call the `create-issue` safe output (`create_issue` tool) exactly once with the full report content.
- If no issue is needed (for example, no findings), call the `noop` safe output exactly once with a short completion message.
- If required capabilities are unavailable, call `missing_tool` with a concise reason.
- If required scan data is unavailable, call `missing_data` with a concise reason.
