# Policy Schema Reference

## Overview

Policies define automation rules for Azure DevOps pull request creation. Each policy specifies a scope, schedule, conditions, actions, and constraints.

## Example Policy

```json
{
  "name": "nightly-dev-to-qa",
  "description": "Automatically create a PR from dev to qa every night at 01:00 SAST",
  "version": "1.0.0",
  "enabled": true,
  "scope": {
    "organization": "pollinate",
    "project": "cloudops",
    "repository": "infrastructure"
  },
  "schedule": {
    "cron": "0 1 * * *",
    "timezone": "Africa/Johannesburg",
    "enabled": true
  },
  "conditions": [
    {
      "type": "branch_exists",
      "operator": "exists",
      "value": "refs/heads/dev"
    }
  ],
  "actions": [
    {
      "type": "create_pr",
      "parameters": {
        "sourceRefName": "refs/heads/dev",
        "targetRefName": "refs/heads/qa",
        "title": "Auto PR: dev -> qa",
        "description": "Automated nightly promotion from dev to qa"
      }
    }
  ],
  "constraints": {
    "max_active_prs": 3,
    "auto_complete": false,
    "merge_strategy": "squash"
  },
  "tags": ["nightly", "promotion", "dev-to-qa"]
}
```

## Field Reference

### Top Level

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | yes | — | Policy name (max 200 chars) |
| `description` | string | no | `""` | Human-readable description |
| `version` | string | yes | — | Semver version (e.g. `1.0.0`) |
| `enabled` | boolean | no | `true` | Whether the policy is active |
| `scope` | object | yes | — | Target Azure DevOps resources |
| `schedule` | object | yes | — | When to evaluate the policy |
| `conditions` | array | no | `[]` | Preconditions for PR creation |
| `actions` | array | yes | — | What to do when conditions pass |
| `constraints` | object | no | `null` | Limits on PR creation |
| `tags` | array | no | `[]` | Labels for filtering |

### Scope

| Field | Type | Required | Description |
|---|---|---|---|
| `organization` | string | yes | Azure DevOps organization name |
| `project` | string | yes | Project name |
| `repository` | string | no | Repository name (omit for all repos in project) |

### Schedule

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `cron` | string | yes | — | Standard 5-field cron expression |
| `timezone` | string | no | `"UTC"` | IANA timezone name |
| `enabled` | boolean | no | `true` | Whether scheduling is active |

### Conditions

Each condition has `type`, `operator`, and `value`:

| Type | Operators | Description |
|---|---|---|
| `branch_exists` | `exists`, `not_exists` | Check if a branch exists in the repo |
| `file_changed` | `contains`, `matches` | Check if specific files changed (Phase 2) |
| `tag_matches` | `equals`, `matches`, `contains` | Check for tags matching a pattern |
| `date_range` | `in_range`, `not_in_range` | Only create PRs within a date window |

### Actions

| Type | Parameters | Description |
|---|---|---|
| `create_pr` | `sourceRefName`, `targetRefName`, `title`, `description`, `isDraft` | Create a pull request |
| `add_reviewer` | `reviewerId` | Add a reviewer to the PR |
| `label` | `labels` | Apply labels to the PR |
| `comment` | `body` | Add a comment to the PR |

### Constraints

| Field | Type | Description |
|---|---|---|
| `max_active_prs` | integer | Max concurrent open PRs for this target branch |
| `auto_complete` | boolean | Auto-complete when all policies pass |
| `merge_strategy` | string | `merge`, `squash`, or `rebase` |
| `require_min_reviewers` | integer | Minimum reviewers before merge |
