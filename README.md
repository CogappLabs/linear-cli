# linear-cli

A small, fast Go CLI for [Linear](https://linear.app), wrapping the Linear
GraphQL API. Built for terminal and agent use: every read command supports
`--json`, and single-issue commands default to the issue on your current git
branch.

Modelled on the `colflow-cli` scaffold (urfave/cli/v3 + a thin GraphQL client +
one file per command).

## Install

```bash
go build -o linear ./cmd/linear
# then move `linear` onto your PATH
```

Homebrew tap + GoReleaser are planned (see Release below), not yet wired.

## Authentication

Create a personal API key at
`linear.app/settings/account/security`, then either:

```bash
linear auth login        # prompts, verifies, saves to ~/.config/linear-cli/config.json
```

or set `LINEAR_API_KEY` in your environment. A project `.env` / `.env.local`
containing `LINEAR_API_KEY` is loaded automatically (the CLI walks up from the
current directory). OS environment variables take precedence over the config
file.

The key is sent as the raw `Authorization` header (Linear does not use a
`Bearer`/`Basic` prefix for personal keys).

## Commands (implemented)

```bash
linear auth login|status
linear whoami [--json]

linear issue list [-t KEY] [--mine] [-s STATE_TYPE ...] [--not-done] [-n N] [--all] [--json]
linear issue pull -t KEY [-s STATE_TYPE ...] [--not-done] [-o out.parquet | --json]
linear issue view [ID] [--comments] [--json]
linear issue create -T TITLE [-t KEY] [-d DESC] [-a EMAIL|me] [-p 0-4] [--json]
linear issue update [ID] [-T TITLE] [-d DESC] [-s STATE_NAME] [-a EMAIL|me] [-p 0-4] [--json]
linear issue start [ID] [--json]
linear issue comment [ID] -b BODY
linear issue comment-edit COMMENT_ID -b BODY
linear issue comment-delete COMMENT_ID
linear issue url [ID]
linear issue branch [ID]

linear team list [--json]
linear team view KEY [--json]
linear user list [--all] [--json]
linear user view EMAIL [--json]

linear project list [-t KEY] [--json]
linear project view ID [--json]
linear project save [ID] --name NAME [-t KEY] [-d DESC] [--target-date YYYY-MM-DD] [--json]
linear milestone list -p PROJECT_ID [--json]
linear milestone view ID [--json]
linear milestone save [ID] --name NAME [-p PROJECT_ID] [-d DESC] [--target-date YYYY-MM-DD] [--json]
linear cycle list -t KEY [--json]
linear label list [-t KEY] [--json]
linear label create -t KEY --name NAME [--color HEX] [--description D] [--json]
linear document list [--json]
linear document view ID [--json]
linear document save [ID] -T TITLE [-c CONTENT] [-t KEY | -p PROJECT_ID] [--json]

linear status list -t KEY [--json]
linear status view -t KEY NAME [--json]
linear attachment list ISSUE [--json]
linear attachment add ISSUE URL [--title T] [--json]
linear attachment delete ATTACHMENT_ID
linear release list [--json]
linear release view ID [--json]
linear release notes RELEASE_ID [--json]
linear status-update list -p PROJECT_ID [--json]
linear status-update create -p PROJECT_ID -b BODY [--health onTrack|atRisk|offTrack] [--json]
linear status-update delete ID
linear agent-skill list [--json]
linear agent-skill view ID [--json]

linear config set-team KEY
linear config show
```

`[ID]` is optional on single-issue commands; when omitted the CLI reads an
`ABC-123` identifier from the current git branch name.

### Bulk pull + Parquet

`issue pull` fetches a team's entire matching set (paging through all results,
not one page) and can write it to a Parquet file:

```bash
linear issue pull -t ENG --not-done -o eng-open.parquet   # everything not completed/canceled
linear issue pull -t ENG -s started -s unstarted --json   # multiple statuses, JSON to stdout
linear issue pull -t ENG                                    # table to stdout
```

Parquet is written with the pure-Go `parquet-go` writer (no CGO, no duckdb
shellout). Columns: identifier, title, state, state_type, assignee, team,
priority, labels, url, created_at, updated_at, description. Query it with
duckdb: `duckdb -c "SELECT state_type, count(*) FROM 'eng-open.parquet' GROUP BY 1"`.

Flag conventions:
- `-t/--team` = team **key** (e.g. `ENG`). `linear config set-team ENG` sets a default.
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

Every Linear MCP tool, its planned CLI command, and current status. Linear's MCP
uses a `save_*` upsert pattern (create when no id, update when id present); the
CLI mirrors that with a `save` subcommand where it makes sense, or splits into
`create`/`update`.

Status: ✅ built · 🟡 next wave · ⚪ later.

### Issues
| MCP tool | CLI | Status |
|---|---|---|
| `get_issue` | `issue view [ID]` | ✅ |
| `list_issues` | `issue list` | ✅ |
| `save_issue` | `issue create` / `issue update` / `issue start` | ✅ |

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

### Issue statuses (workflow states)
| MCP tool | CLI | Status |
|---|---|---|
| `list_issue_statuses` | `status list -t KEY` | 🟡 |
| `get_issue_status` | `status view -t KEY NAME` | 🟡 |

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
| `create_attachment_from_upload` | `attachment upload ISSUE FILE` | ⚪ (file-upload flow) |
| `prepare_attachment_upload` | (internal to upload) | ⚪ |
| `extract_images` | `attachment extract-images ISSUE` | ⚪ |

### Releases
| MCP tool | CLI | Status |
|---|---|---|
| `list_releases` | `release list` | ✅ |
| `get_release` | `release view ID` | ✅ |
| `list_release_notes` | `release notes RELEASE_ID` | ✅ |
| `get_release_note` / `save_release*` / `list_release_pipelines` | — | ⚪ |

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
go build -o linear ./cmd/linear
go vet ./...
gofmt -l .          # should print nothing

# Introspect the GraphQL schema for a type (needs LINEAR_API_KEY):
./scripts/introspect.sh Issue Project Cycle
```

Architecture and GraphQL notes are in `CLAUDE.md`.

## Release (planned)

GoReleaser on a `v*` tag → GitHub release (darwin/linux × amd64/arm64) →
updates `CogappLabs/homebrew-tap`, matching `colflow-cli`.
