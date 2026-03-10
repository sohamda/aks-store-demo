# GitHub Copilot Agentic Workshop — Content Overview

## Workshop at a Glance

|                  |                                                                                 |
| ---------------- | ------------------------------------------------------------------------------- |
| **Duration**     | ~75 min core + take-home exercises                                              |
| **Platform**     | github.com only (no IDE)                                                        |
| **Repo**         | aks-store-demo (8 polyglot microservices, zero Copilot/DevSecOps config)        |
| **Prerequisite** | GitHub account with Copilot Pro/Pro+/Business/Enterprise + Coding Agent enabled |
| **Setup**        | Fork the repo, enable GitHub Actions, create `PAT_TOKEN` repo secret            |

## Core Workshop Modules

| #   | Module                                           | Duration | Format                       | What Attendees Do                                                                                                                                         | Topic Covered                        | Catch-Up Branch       |
| --- | ------------------------------------------------ | -------- | ---------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ | --------------------- |
| 0   | **Welcome + Orientation**                        | 8 min    | Presenter-led                | Listen; confirm fork + Actions ready                                                                                                                      | —                                    | —                     |
| 1   | **Copilot Chat — Security Assessment**           | 10 min   | Demo 3 min → Hands-on 7 min  | Use Copilot Chat on github.com to audit the repo for security gaps, missing tests, hardcoded creds                                                        | github.com agentic features          | —                     |
| 2   | **Foundation — Custom Instructions + DevSecOps** | 12 min   | Demo 4 min → Hands-on 8 min  | Create 3 files via github.com UI: `copilot-instructions.md`, `dependabot.yml` (assignees: copilot), `codeql.yml`                                          | Custom/agentic SDLC                  | `checkpoint/module-2` |
| 3   | **DevSecOps Showcase**                           | 10 min   | Facilitator demo             | Watch: Dependabot alerts, CodeQL findings, Copilot Autofix, Secret Scanning (facilitator's pre-baked fork)                                                | Agentic SDLC, DevSecOps              | —                     |
| 4   | **Agentic Workflows — Automated Issue Pipeline** | 20 min   | Demo 7 min → Hands-on 13 min | Create `security-audit-autofix.yml` → trigger → watch: npm/pip audit + test coverage check → auto-creates issues → assigns to Copilot → agent creates PRs | Agentic workflows, automated actions | `checkpoint/module-4` |
| 5   | **Coding Agent + Code Review — Full Loop**       | 15 min   | Demo 5 min → Hands-on 10 min | Inspect agent PRs, manually create 1 issue & assign to Copilot, add Copilot as PR reviewer, interact with review comments                                 | Agentic loops, Code Review           | —                     |
| —   | **Recap + Wrap-Up**                              | 5 min    | Presenter-led                | Review the full autonomous pipeline diagram; see what's now running daily on their fork                                                                   | All topics                           | —                     |

## Topic Coverage Matrix

| Workshop Topic                  |        Module 1        |       Module 2        |       Module 3       |            Module 4            |       Module 5       |
| ------------------------------- | :--------------------: | :-------------------: | :------------------: | :----------------------------: | :------------------: |
| **github.com agentic features** |        ✅ Chat         |                       |                      |                                |    ✅ Code Review    |
| **Agentic loops**               |                        |                       |                      | ✅ scan→issue→agent→CI→iterate | ✅ PR commit history |
| **Agentic workflows**           |                        | ✅ Dependabot+Copilot |                      |     ✅ Automated pipeline      |                      |
| **Custom/agentic SDLC**         |                        |    ✅ Instructions    |     ✅ DevSecOps     |                                |     ✅ Full loop     |
| **DevSecOps showcase**          | ✅ Security assessment | ✅ CodeQL, Dependabot | ✅ Autofix, Scanning |      ✅ Auto-remediation       |                      |

## Files Created During Workshop

| Module | File                                           | Purpose                                                                             |
| ------ | ---------------------------------------------- | ----------------------------------------------------------------------------------- |
| 2      | `.github/copilot-instructions.md`              | Coding standards + security requirements — shapes all Copilot behavior              |
| 2      | `.github/dependabot.yml`                       | Daily dep scanning (npm, pip, gomod, cargo, actions) — auto-assigns to Copilot      |
| 2      | `.github/workflows/codeql.yml`                 | CodeQL SAST for JS, Python, Go — feeds Copilot Autofix                              |
| 4      | `.github/workflows/security-audit-autofix.yml` | Automated pipeline: audit deps + find test gaps → create issues → assign to Copilot |

## Take-Home Exercises (async, post-workshop)

| #   | Exercise                         | Difficulty  | What to Do                                                                                                  |
| --- | -------------------------------- | ----------- | ----------------------------------------------------------------------------------------------------------- |
| 1   | Review Dependabot/CodeQL results | Easy        | Check alerts and Copilot Autofix on your fork next day                                                      |
| 2   | Container scan workflow          | Medium      | Create `docker-scan.yml` (Trivy → issues → Copilot) — available in `checkpoint/complete`                    |
| 3   | More Copilot issues              | Easy–Medium | Create issues for: pytest for ai-service, Go tests for makeline-service, API docs, OpenAPI spec, CODEOWNERS |
| 4   | Branch protection rules          | Easy        | Require Copilot Code Review + CI pass before merge                                                          |
| 5   | Extend the pipeline              | Advanced    | Add govulncheck for Go, cargo-audit for Rust, SBOM generation                                               |

## Pre-Workshop Checklist

| Who             | Task                                                                     |
| --------------- | ------------------------------------------------------------------------ |
| **Facilitator** | Prepare source repo with checkpoint branches                             |
| **Facilitator** | Create pre-baked fork with Dependabot/CodeQL results (30+ min lead time) |
| **Facilitator** | Prepare handout with YAML snippets + issue bodies (copy-paste ready)     |
| **Facilitator** | Verify Copilot + Coding Agent enabled at org level                       |
| **Facilitator** | Dry-run full pipeline end-to-end                                         |
| **Attendee**    | Fork the repo                                                            |
| **Attendee**    | Enable GitHub Actions (Settings → Actions → Allow all)                   |
| **Attendee**    | Verify Copilot Chat works on github.com                                  |
| **Attendee**    | Create `PAT_TOKEN` repo secret (Fine-grained PAT: Actions, Contents, Issues, Pull requests — all R&W) |
