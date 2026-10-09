# linear-cli

A small, fast Go CLI for [Linear](https://linear.app), wrapping the Linear
GraphQL API. Built for terminal and agent use: every read command supports
`--json`, and single-issue commands default to the issue on your current git
branch.

Modelled on the `colflow-cli` scaffold (urfave/cli/v3 + a thin GraphQL client +
one file per command).

## Install

```bash
brew install CogappLabs/tap/linear-cli
```

Use the tap-qualified name. A bare `brew install linear-cli` resolves to
unrelated core formulae (`linode-cli`, the `linear` cask).

Or build from source:

```bash
go build -o linear-cli ./cmd/linear
# then move `linear-cli` onto your PATH
```

## Authentication

Create a personal API key at `linear.app/settings/account/security`.

For a Homebrew install (the binary runs from any directory), store the key once:

```bash
linear-cli auth login    # prompts, verifies, saves to the config file
```

The key is saved to `~/Library/Application Support/linear-cli/config.json` on
macOS (`~/.config/linear-cli/config.json` on Linux) and used from anywhere.

Resolution order, highest first:
1. `LINEAR_API_KEY` already in the environment (set it in your shell profile for
   a global key).
2. `LINEAR_API_KEY` from a `.env` / `.env.local` in the current directory or any
   parent (a real shell variable takes precedence over the file). Per-project
   convenience: it only applies when you run the CLI inside that tree, so it
   does not replace `auth login` for a Homebrew install.
3. The config file written by `auth login`.

The key is sent as the raw `Authorization` header (Linear does not use a
`Bearer`/`Basic` prefix for personal keys).

## Commands (implemented)

```bash
linear-cli auth login|status
linear-cli whoami [--json]

linear-cli issue list [-t KEY] [--mine] [-s STATE_TYPE_OR_NAME ...] [--not-done] [-n N] [--all] [--json]
linear-cli issue pull -t KEY [-s STATE_TYPE_OR_NAME ...] [--not-done] [-o out.parquet | --json]
linear-cli issue view [ID] [--comments] [--json]
linear-cli issue create -T TITLE [-t KEY] [-d DESC] [-a EMAIL|me] [-p 0-4] [-l LABEL ...] [--json]
linear-cli issue update [ID] [-T TITLE] [-d DESC] [-s STATE_NAME] [-a EMAIL|me] [-p 0-4] [-l LABEL ...] [--remove-label LABEL ...] [--json]
linear-cli issue start [ID] [--json]
linear-cli issue comment [ID] -b BODY
linear-cli issue comment-edit COMMENT_ID -b BODY
linear-cli issue comment-delete COMMENT_ID
linear-cli issue url [ID]
linear-cli issue branch [ID]
linear-cli issue relate ID OTHER-ID [--type related|blocks|blocked-by|duplicate]
linear-cli issue unrelate ID OTHER-ID

linear-cli team list [--json]
linear-cli team view KEY [--json]
linear-cli user list [--all] [--json]
linear-cli user view EMAIL [--json]

linear-cli project list [-t KEY] [--json]
linear-cli project view ID [--json]
linear-cli project save [ID] --name NAME [-t KEY] [-d DESC] [--target-date YYYY-MM-DD] [--json]
linear-cli milestone list -p PROJECT_ID [--json]
linear-cli milestone view ID [--json]
linear-cli milestone save [ID] --name NAME [-p PROJECT_ID] [-d DESC] [--target-date YYYY-MM-DD] [--json]
linear-cli cycle list -t KEY [--json]
linear-cli label list [-t KEY] [--json]
linear-cli label create -t KEY --name NAME [--color HEX] [--description D] [--json]
linear-cli document list [--json]
linear-cli document view ID [--json]
linear-cli document save [ID] -T TITLE [-c CONTENT] [-t KEY | -p PROJECT_ID] [--json]

linear-cli status list -t KEY [--json]
linear-cli status view -t KEY NAME [--json]
linear-cli attachment list ISSUE [--json]
linear-cli attachment add ISSUE URL [--title T] [--json]
linear-cli attachment upload ISSUE FILE [--title T] [--json]
linear-cli attachment extract-images ISSUE [--json]
linear-cli attachment delete ATTACHMENT_ID
linear-cli release list [--json]
linear-cli release view ID [--json]
linear-cli release notes RELEASE_ID [--json]
linear-cli status-update list -p PROJECT_ID [--json]
linear-cli status-update create -p PROJECT_ID -b BODY [--health onTrack|atRisk|offTrack] [--json]
linear-cli status-update delete ID
linear-cli agent-skill list [--json]
linear-cli agent-skill view ID [--json]

linear-cli config set-team KEY
linear-cli config show
```

`[ID]` is optional on single-issue commands; when omitted the CLI reads an
`ABC-123` identifier from the current git branch name.

### Bulk pull + Parquet

`issue pull` fetches a team's entire matching set (paging through all results,
not one page) and can write it to a Parquet file:

```bash
linear-cli issue pull -t ENG --not-done -o eng-open.parquet   # everything not completed/canceled
linear-cli issue pull -t ENG -s started -s unstarted --json   # multiple statuses, JSON to stdout
linear-cli issue pull -t ENG                                    # table to stdout
```

Parquet is written with the pure-Go `parquet-go` writer (no CGO, no duckdb
shellout). Columns: identifier, title, state, state_type, assignee, team,
priority, labels, url, created_at, updated_at, description. Query it with
duckdb: `duckdb -c "SELECT state_type, count(*) FROM 'eng-open.parquet' GROUP BY 1"`.

Flag conventions:
- `-t/--team` = team **key** (e.g. `ENG`). `linear-cli config set-team ENG` sets a default.
- `-s/--state` on `update` = state **name** (`"In Review"`); on `list` = state **type** (`started`, `unstarted`, `completed`, `canceled`, `backlog`, `triage`).
- `-a/--assignee` = `me` or an email.
- `-p/--priority` = `0 none, 1 urgent, 2 high, 3 medium, 4 low`.

## Why a CLI (vs the Linear MCP server)

The Linear MCP server exposes 53 tools; each tool's schema costs context
whenever the server is active. This CLI adds no persistent tool schema, and its
`--json` output is trimmed to the fields the commands select rather than the
full 100+-field Linear objects. For the common path (issue list/view/create/
update/start/comment) it is cheaper on context and quicker to read. For the long
tail, fall back to the MCP server (see the parity table below).

## MCP parity

Every Linear MCP tool, the matching CLI command, and its status. Linear's MCP
uses a `save_*` upsert pattern (create when no id, update when id present); the
CLI mirrors that with a `save` subcommand where it makes sense, or splits into
`create`/`update`.

Status: ✅ built · 🟡 not built (buildable) · ⚪ not planned.

### Issues
| MCP tool | CLI | Status |
|---|---|---|
| `get_issue` | `issue view [ID]` | ✅ |
| `list_issues` | `issue list` | ✅ |
| `save_issue` | `issue create` / `issue update` / `issue start` | ✅ |
| `save_issue` (relations) | `issue relate ID OTHER-ID --type ...` / `issue unrelate ID OTHER-ID`; relations shown in `issue view` | ✅ |

### Comments
| MCP tool | CLI | Status |
|---|---|---|
| `list_comments` | `issue comments [ID]` | ✅ |
| `save_comment` | `issue comment [ID] -b` (create); `issue comment-edit ID -b` (update) | ✅ |
| `delete_comment` | `issue comment-delete COMMENT_ID` | ✅ |

### Teams & users
| MCP tool | CLI | Status |
|---|---|---|
| `list_teams` | `team list` | ✅ |
| `get_team` | `team view KEY` | ✅ |
| `list_users` | `user list` | ✅ |
| `get_user` | `user view EMAIL` | ✅ |

### Labels
| MCP tool | CLI | Status |
|---|---|---|
| `list_issue_labels` | `label list [-t KEY]` | ✅ |
| `list_project_labels` | `label list --project ID` | 🟡 |
| `create_issue_label` | `label create -t KEY --name NAME [--color]` | ✅ |

### Projects
| MCP tool | CLI | Status |
|---|---|---|
| `list_projects` | `project list [-t KEY]` | ✅ |
| `get_project` | `project view ID` | ✅ |
| `save_project` | `project save [ID] --name … --team …` | ✅ |

### Milestones
| MCP tool | CLI | Status |
|---|---|---|
| `list_milestones` | `milestone list -p PROJECT_ID` | ✅ |
| `get_milestone` | `milestone view ID` | ✅ |
| `save_milestone` | `milestone save [ID] --project … --name … --target-date …` | ✅ |

### Cycles
| MCP tool | CLI | Status |
|---|---|---|
| `list_cycles` | `cycle list -t KEY` | ✅ |

### Documents
| MCP tool | CLI | Status |
|---|---|---|
| `list_documents` | `document list` | ✅ |
| `get_document` | `document view ID` | ✅ |
| `save_document` | `document save [ID] --title … --content … --team|--project …` | ✅ |

### Issue statuses (workflow states)
| MCP tool | CLI | Status |
|---|---|---|
| `list_issue_statuses` | `status list -t KEY` | ✅ |
| `get_issue_status` | `status view -t KEY NAME` | ✅ |

### Attachments
| MCP tool | CLI | Status |
|---|---|---|
| `get_attachment` / `list` | `attachment list ISSUE` | ✅ |
| `create_attachment` | `attachment add ISSUE URL --title …` (via `attachmentLinkURL`) | ✅ |
| `delete_attachment` | `attachment delete ID` | ✅ |
| `create_attachment_from_upload` | `attachment upload ISSUE FILE` | ✅ |
| `prepare_attachment_upload` | (internal to `attachment upload`) | ✅ |
| `extract_images` | `attachment extract-images ISSUE` | ✅ |

### Releases
| MCP tool | CLI | Status |
|---|---|---|
| `list_releases` | `release list` | ✅ |
| `get_release` | `release view ID` | ✅ |
| `list_release_notes` | `release notes RELEASE_ID` | ✅ |
| `get_release_note` / `save_release*` / `list_release_pipelines` | n/a | ⚪ |

### Status updates (project updates)
| MCP tool | CLI | Status |
|---|---|---|
| `get_status_updates` | `status-update list -p PROJECT_ID` | ✅ |
| `save_status_update` | `status-update create -p PROJECT_ID -b BODY` | ✅ |
| `delete_status_update` | `status-update delete ID` | ✅ |

### Agent skills
| MCP tool | CLI | Status |
|---|---|---|
| `list_agent_skills` | `agent-skill list` | ✅ |
| `get_agent_skill` | `agent-skill view ID` | ✅ |

### Not backed by the public GraphQL API (MCP-only)
These MCP tools are server-side abstractions with no matching public-schema
query/mutation, so the CLI cannot cover them. Use the MCP server:

- Code diffs & review: `list_diffs`, `get_diff`, `get_diff_threads`,
  `save_diff_comment`, `delete_diff_comment`, `resolve_diff_thread`,
  `submit_diff_review`, `merge_diff`.
- Docs search: `search_documentation`.

## Build order

1. ✅ v1: auth, whoami, issues, teams, users, config.
2. ✅ Read parity: comments, team/user view, projects, milestones, cycles,
   labels, documents, statuses.
3. ✅ MCP holes: attachments (list/add/delete), releases (list/view/notes),
   status updates (list/create/delete), agent skills, comment edit/delete, and
   `save` for projects / milestones / documents.
4. ✅ Bulk: `issue pull` (full team pull, multi-status, Parquet export).
5. ⚪ Not planned: diffs & review and docs search (MCP-only, see above);
   attachment file uploads; release-note/pipeline writes.

## Development

```bash
go build -o linear-cli ./cmd/linear
go test ./...
go vet ./...
gofmt -l .          # should print nothing

# Introspect the GraphQL schema for a type (needs LINEAR_API_KEY):
./scripts/introspect.sh Issue Project Cycle
```

Architecture and GraphQL notes are in `CLAUDE.md`.

## Release

Push a `v*` tag. GoReleaser builds darwin/linux amd64+arm64, publishes a GitHub
release, and updates the `CogappLabs/homebrew-tap` formula (via the
`TAP_GITHUB_TOKEN` repo secret). Config in `.goreleaser.yaml`, workflow in
`.github/workflows/release.yml`. Mirrors `colflow-cli`.
