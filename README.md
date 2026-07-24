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

linear issue list [-t KEY] [--mine] [-s STATE_TYPE] [-n N] [--json]
linear issue view [ID] [--comments] [--json]
linear issue create -T TITLE [-t KEY] [-d DESC] [-a EMAIL|me] [-p 0-4] [--json]
linear issue update [ID] [-T TITLE] [-d DESC] [-s STATE_NAME] [-a EMAIL|me] [-p 0-4] [--json]
linear issue start [ID] [--json]
linear issue comment [ID] -b BODY
linear issue url [ID]
linear issue branch [ID]

linear team list [--json]
linear team view KEY [--json]
linear user list [--all] [--json]
linear user view EMAIL [--json]

linear project list [-t KEY] [--json]
linear project view ID [--json]
linear milestone list -p PROJECT_ID [--json]
linear milestone view ID [--json]
linear cycle list -t KEY [--json]
linear label list [-t KEY] [--json]
linear label create -t KEY --name NAME [--color HEX] [--description D] [--json]
linear document list [--json]
linear document view ID [--json]

linear config set-team KEY
linear config show
```

`[ID]` is optional on single-issue commands; when omitted the CLI reads an
`ABC-123` identifier from the current git branch name.

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
| `save_comment` | `issue comment [ID] -b` (create); `--edit ID` (update) | ✅ create · 🟡 edit |
| `delete_comment` | `issue comment-delete COMMENT_ID` | 🟡 |

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
| `save_project` | `project save [ID] --name … --state …` | ⚪ |

### Milestones
| MCP tool | CLI | Status |
|---|---|---|
| `list_milestones` | `milestone list -p PROJECT_ID` | ✅ |
| `get_milestone` | `milestone view ID` | ✅ |
| `save_milestone` | `milestone save [ID] --project … --name … --target-date …` | ⚪ |

### Cycles
| MCP tool | CLI | Status |
|---|---|---|
| `list_cycles` | `cycle list -t KEY` | ✅ |

### Documents
| MCP tool | CLI | Status |
|---|---|---|
| `list_documents` | `document list` | ✅ |
| `get_document` | `document view ID` | ✅ |
| `save_document` | `document save [ID] --title … --content …` | ⚪ |

### Attachments
| MCP tool | CLI | Status |
|---|---|---|
| `get_attachment` | `attachment view ID` | ⚪ |
| `create_attachment` | `attachment add ISSUE --url … --title …` | ⚪ |
| `create_attachment_from_upload` | `attachment upload ISSUE FILE` | ⚪ |
| `prepare_attachment_upload` | (internal to `attachment upload`) | ⚪ |
| `delete_attachment` | `attachment delete ID` | ⚪ |
| `extract_images` | `attachment extract-images ISSUE` | ⚪ |

### Code diffs & review
| MCP tool | CLI | Status |
|---|---|---|
| `list_diffs` | `diff list ISSUE` | ⚪ |
| `get_diff` | `diff view ID` | ⚪ |
| `get_diff_threads` | `diff threads ID` | ⚪ |
| `save_diff_comment` | `diff comment ID -b` | ⚪ |
| `delete_diff_comment` | `diff comment-delete ID` | ⚪ |
| `resolve_diff_thread` | `diff resolve THREAD_ID` | ⚪ |
| `submit_diff_review` | `diff review ID --state …` | ⚪ |
| `merge_diff` | `diff merge ID` | ⚪ |

### Releases
| MCP tool | CLI | Status |
|---|---|---|
| `list_releases` | `release list` | ⚪ |
| `get_release` | `release view ID` | ⚪ |
| `save_release` | `release save [ID] …` | ⚪ |
| `list_release_notes` | `release notes ID` | ⚪ |
| `get_release_note` | `release note-view ID` | ⚪ |
| `save_release_note` | `release note-save [ID] …` | ⚪ |
| `list_release_pipelines` | `release pipelines` | ⚪ |

### Status updates
| MCP tool | CLI | Status |
|---|---|---|
| `get_status_updates` | `status-update list --project ID` | ⚪ |
| `save_status_update` | `status-update save [ID] --project … --body …` | ⚪ |
| `delete_status_update` | `status-update delete ID` | ⚪ |

### Agent skills & docs
| MCP tool | CLI | Status |
|---|---|---|
| `list_agent_skills` | `agent-skill list` | ⚪ |
| `get_agent_skill` | `agent-skill view ID` | ⚪ |
| `search_documentation` | `docs search QUERY` | ⚪ |

## Build order

1. ✅ v1: auth, whoami, issues, teams, users, config.
2. ✅ Read parity: comments list, team/user view, projects, milestones, cycles,
   labels (+ create), documents. `list`/`view` for every core domain.
3. 🟡 Remaining reads: issue statuses (`status list/view`).
4. ⚪ Later (writes + specialised): project/milestone/document `save`,
   attachments, diffs & review, releases, status updates, agent skills, docs
   search.

Each wave is a set of atomic commits (one domain per commit where practical).

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
