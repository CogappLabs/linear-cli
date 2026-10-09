# linear-cli

Go CLI for Linear. Wraps the Linear GraphQL API for issues, teams, and users from the terminal. Modelled on the colflow-cli scaffold (same stack: urfave/cli/v3 + a thin GraphQL client + one file per command).

## Architecture

- `cmd/linear/main.go`: root urfave/cli/v3 command, registers subcommands. Calls `config.LoadDotEnv()` at startup.
- `internal/client/`: Linear GraphQL client. `client.go` is transport (`Query()`), `queries.go` per-operation wrappers, `types.go` response shapes. All API access goes through `Query()`.
- `internal/config/`: API key + default team persisted to `~/.config/linear-cli/config.json` (0600). `ResolveAPIKey` prefers `LINEAR_API_KEY` env over the file. `dotenv.go` walks up from cwd loading `.env`/`.env.local` so a project key is picked up (OS env wins).
- `internal/format/`: terminal output helpers (colour, `ColorState`, `Priority`, `TimeAgo`, ANSI-aware `PadRight`).
- `internal/gitutil/`: extracts an `ABC-123` identifier from the current git branch (`IssueIDFromBranch`), for commands that default to "the issue I'm working on".
- `internal/commands/`: one file per command group, each returning `*cli.Command`. `common.go` has `requireAuth()` (loads the key into the client), `printJSON`, `die`.

## Conventions

- urfave/cli/v3. Every command that hits the API calls `requireAuth()` first thing in its `Action`.
- Every read command supports `--json` (raw struct dump). Human output is the default.
- Commands that act on a single issue (`view`, `update`, `start`, `comment`, `url`, `branch`) take an optional `[ID]` arg; when omitted they fall back to `gitutil.IssueIDFromBranch()`.
- Issue identifiers (`ABC-123`) are passed straight to the API. `issue(id:)` and `issueUpdate(id:)` accept them. Only `commentCreate` needs the issue UUID, so `issue comment` resolves the identifier to a UUID via `GetIssue` first.
- State names, label names and assignee emails are resolved to ids client-side (`resolveStateID`, `resolveLabelIDs`, `resolveUserID`) because the mutations want ids. Labels match the team's own plus workspace labels; `issue update -l` adds via `addedLabelIds` rather than replacing. `--assignee me` and state-by-name are the ergonomic path.

## Adding a new command

1. Create `internal/commands/<name>.go` returning `*cli.Command` (or add a subcommand to an existing group).
2. Call `requireAuth()` at the start of the `Action`.
3. Add a `--json` flag and honour it for read commands.
4. Add any new GraphQL wrapper to `internal/client/queries.go`, not inline in the command.
5. Register in `cmd/linear/main.go` (top-level) or the group's `Commands` slice.
6. Build: `go build -o linear-cli ./cmd/linear`.
7. Update `README.md` and this file if the command surface or file layout changed.

## Authentication

- Personal API key from linear.app/settings/account/security.
- Header is the raw key: `Authorization: <API_KEY>` (no `Bearer`/`Basic` prefix).
- Resolution order: `LINEAR_API_KEY` env (incl. loaded `.env`) → `~/.config/linear-cli/config.json`.
- `LINEAR_CLI_CONFIG_DIR` overrides the config directory (tests set it to a temp dir so they never touch the real config; `os.UserConfigDir()` ignores `XDG_CONFIG_HOME` on macOS, so an explicit override is the reliable way to isolate).
- `linear-cli auth login` verifies the key with a `viewer` query before saving.

## GraphQL schema notes

Endpoint: `https://api.linear.app/graphql`. Introspect with `scripts/introspect.sh` (reads `LINEAR_API_KEY`; pass type names, defaults to the core set).

- `issue(id: String!)` accepts either the identifier (`ABC-123`) or the UUID.
- `issues(filter: IssueFilter, first: Int, orderBy: updatedAt)`: filter is nested comparators, e.g. `{ team: { key: { eq: "ENG" } }, assignee: { id: { eq: "<uuid>" } }, state: { type: { eq: "started" } } }`.
- Connections wrap lists as `{ nodes { … }, pageInfo { hasNextPage endCursor } }`. Current commands take the first page only; add cursor paging when a list needs it.
- `IssueCreateInput` requires `teamId`; useful optional fields: `title`, `description` (markdown), `assigneeId`, `stateId`, `priority` (0-4), `labelIds`, `projectId`, `projectMilestoneId`, `parentId`, `dueDate` (TimelessDate), `cycleId`.
- `IssueUpdateInput` mirrors those (no `teamId` needed; the issue already has a team).
- `IssueRelation` rows read `issue <type> relatedIssue` (types: blocks, duplicate, related, similar). `issue.relations` are outgoing, `issue.inverseRelations` incoming; "blocked by" is a `blocks` row with the issues swapped. `issueRelationDelete` takes the relation id.
- `issue list/pull -s` values that are not a state type are matched as workflow state names (`eqIgnoreCase`), see `buildStateFilter`.
- `priority` is `0 none, 1 urgent, 2 high, 3 medium, 4 low` (`format.Priority`).
- `WorkflowState.type` is one of `triage, backlog, unstarted, started, completed, canceled`. `issue start` picks the first `started` state for the team.
- `commentCreate(input: CommentCreateInput)` needs `issueId` as the UUID plus `body` (markdown).
- Mutations return `{ success, <entity> }`; wrappers error when `success` is false.

## Bulk pull + Parquet

`issue pull` (`issue_pull.go`) fetches a whole team via `ListIssues` with
`All: true`, which pages the connection (250/page) until exhausted. It flattens
each issue to `issueRow` and writes Parquet with `parquet.NewGenericWriter`
(pure Go, no CGO). Nested objects are flattened to scalar columns; labels are
comma-joined. `--parquet/-o` writes the file, else `--json` or a table.

Multi-status filtering lives in `buildIssueFilter`: `--state` (repeatable) →
`state.type.in`; `--not-done` → `state.type.nin ["completed","canceled"]`
(overrides `--state`). Shared by `issue list` and `issue pull`.

## Linear MCP parity

Full MCP-tool → CLI mapping with status is in `README.md`. Built: issues (incl.
bulk pull), comments (+ edit/delete), teams, users, projects (+ save),
milestones (+ save), cycles, labels (+ create), documents (+ save), statuses,
attachments (list/add/delete), releases (list/view/notes), status updates
(project updates), agent skills.

MCP-only (no public-schema backing, cannot build): code diffs & review,
docs search. Also unbuilt: attachment file uploads, release-note/pipeline
writes. Field names came from the Linear SDK schema at
`~/git/linear/packages/sdk/src/schema.graphql` (more reliable than live
introspection, which the sandbox blocks via `.env`).

## Tests

`go test ./...`. Unit tests cover pure logic only, no network: `buildIssueFilter`
(filter shaping), `identifierFromBranch` (branch parsing), `format` helpers,
the markdown image regex, and config save/load + key resolution. Command
`Action`s and GraphQL wrappers are exercised by hand against the live API (writes
go to the `LUKE` team). CI runs gofmt, vet, build, and test on Linux.

## Release

GoReleaser on a `v*` tag builds darwin/linux amd64+arm64, publishes a GitHub release, and updates `CogappLabs/homebrew-tap` (needs the `TAP_GITHUB_TOKEN` secret). Config in `.goreleaser.yaml`, workflow in `.github/workflows/release.yml`. Mirrors colflow.

## Command file layout

Each top-level group is one file returning a `*cli.Command`. Wave-2 read
commands live in their own files, and their GraphQL wrappers live in
`internal/client/queries_<domain>.go` (not the shared `queries.go`) so they can
be added independently:

- `issue.go` + `issue_mutate.go` + `issue_comments.go` + `issue_pull.go`
- `team.go`, `user.go` (list + view)
- `project.go` (+ `queries_project.go`, list/view/save)
- `milestone.go` (+ `queries_milestone.go`, list/view/save)
- `cycle.go` (+ `queries_cycle.go`)
- `label.go` (+ `queries_label.go`, list + create)
- `document.go` (+ `queries_document.go`, list/view/save)
- `status.go` (+ `queries_status.go`, workflow states)
- `attachment.go` (+ `queries_attachment.go`, list/add/delete)
- `release.go` (+ `queries_release.go`, list/view/notes)
- `statusupdate.go` (+ `queries_statusupdate.go`, project updates)
- `agentskill.go` (+ `queries_agentskill.go`)
- `config.go`

`save` subcommands (project/milestone/document) and comment edit/delete are in
the existing `<domain>.go` / `issue_mutate.go` files; their client wrappers were
added to the matching `queries_*.go`.

Shared `queries.go` holds issues, teams, users, states, and the `ListUsers`
pagination helper. `ListUsers` pages through the full connection (250/page)
because the workspace has more than 50 members, so a single page would drop
the rest. Anything resolving a user by email (`GetUser`, `resolveUserID`)
depends on that.

## Available commands

`auth login|status`, `whoami`, `issue list|pull|view|create|update|start|comment|comments|comment-edit|comment-delete|url|branch`, `team list|view`, `user list|view`, `project list|view|save`, `milestone list|view|save`, `cycle list`, `label list|create`, `document list|view|save`, `status list|view`, `attachment list|add|delete`, `release list|view|notes`, `status-update list|create|delete`, `agent-skill list|view`, `config set-team|show`.
