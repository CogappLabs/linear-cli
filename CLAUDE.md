# linear-cli

Go CLI for Linear. Wraps the Linear GraphQL API for issues, teams, and users from the terminal. Modelled on the colflow-cli scaffold (same stack: urfave/cli/v3 + a thin GraphQL client + one file per command).

## Architecture

- `cmd/linear/main.go` — root urfave/cli/v3 command, registers subcommands. Calls `config.LoadDotEnv()` at startup.
- `internal/client/` — Linear GraphQL client. `client.go` is transport (`Query()`), `queries.go` per-operation wrappers, `types.go` response shapes. All API access goes through `Query()`.
- `internal/config/` — API key + default team persisted to `~/.config/linear-cli/config.json` (0600). `ResolveAPIKey` prefers `LINEAR_API_KEY` env over the file. `dotenv.go` walks up from cwd loading `.env`/`.env.local` so a project key is picked up (OS env wins).
- `internal/format/` — terminal output helpers (colour, `ColorState`, `Priority`, `TimeAgo`, ANSI-aware `PadRight`).
- `internal/gitutil/` — extracts an `ABC-123` identifier from the current git branch (`IssueIDFromBranch`), for commands that default to "the issue I'm working on".
- `internal/commands/` — one file per command group, each returning `*cli.Command`. `common.go` has `requireAuth()` (loads the key into the client), `printJSON`, `die`.

## Conventions

- urfave/cli/v3. Every command that hits the API calls `requireAuth()` first thing in its `Action`.
- Every read command supports `--json` (raw struct dump). Human output is the default.
- Commands that act on a single issue (`view`, `update`, `start`, `comment`, `url`, `branch`) take an optional `[ID]` arg; when omitted they fall back to `gitutil.IssueIDFromBranch()`.
- Issue identifiers (`ABC-123`) are passed straight to the API — `issue(id:)`, `issueUpdate(id:)` accept them. Only `commentCreate` needs the issue UUID, so `issue comment` resolves the identifier to a UUID via `GetIssue` first.
- State names and assignee emails are resolved to ids client-side (`resolveStateID`, `resolveUserID`) because the mutations want ids. `--assignee me` and state-by-name are the ergonomic path.

## Adding a new command

1. Create `internal/commands/<name>.go` returning `*cli.Command` (or add a subcommand to an existing group).
2. Call `requireAuth()` at the start of the `Action`.
3. Add a `--json` flag and honour it for read commands.
4. Add any new GraphQL wrapper to `internal/client/queries.go`, not inline in the command.
5. Register in `cmd/linear/main.go` (top-level) or the group's `Commands` slice.
6. Build: `go build -o linear ./cmd/linear`.
7. Update `README.md` and this file if the command surface or file layout changed.

## Authentication

- Personal API key from linear.app/settings/account/security.
- Header is the raw key: `Authorization: <API_KEY>` — no `Bearer`/`Basic` prefix.
- Resolution order: `LINEAR_API_KEY` env (incl. loaded `.env`) → `~/.config/linear-cli/config.json`.
- `linear auth login` verifies the key with a `viewer` query before saving.

## GraphQL schema notes

Endpoint: `https://api.linear.app/graphql`. Introspect with `scripts/introspect.sh` (reads `LINEAR_API_KEY`; pass type names, defaults to the core set).

- `issue(id: String!)` accepts either the identifier (`ABC-123`) or the UUID.
- `issues(filter: IssueFilter, first: Int, orderBy: updatedAt)` — filter is nested comparators, e.g. `{ team: { key: { eq: "ENG" } }, assignee: { id: { eq: "<uuid>" } }, state: { type: { eq: "started" } } }`.
- Connections wrap lists as `{ nodes { … }, pageInfo { hasNextPage endCursor } }`. Current commands take the first page only; add cursor paging when a list needs it.
- `IssueCreateInput` requires `teamId`; useful optional fields: `title`, `description` (markdown), `assigneeId`, `stateId`, `priority` (0-4), `labelIds`, `projectId`, `projectMilestoneId`, `parentId`, `dueDate` (TimelessDate), `cycleId`.
- `IssueUpdateInput` mirrors those (no `teamId` needed; the issue already has a team).
- `priority` is `0 none, 1 urgent, 2 high, 3 medium, 4 low` (`format.Priority`).
- `WorkflowState.type` is one of `triage, backlog, unstarted, started, completed, canceled`. `issue start` picks the first `started` state for the team.
- `commentCreate(input: CommentCreateInput)` needs `issueId` as the UUID plus `body` (markdown).
- Mutations return `{ success, <entity> }`; wrappers error when `success` is false.

## Linear MCP parity (roadmap)

The Linear MCP server exposes a wider surface we may want as commands. Mapping of MCP tool → planned CLI command lives in `README.md` under "Roadmap". Not yet built: projects, milestones, cycles, labels, documents, attachments, diffs/reviews, releases, status updates. The GraphQL types for these were introspected (see `scripts/introspect.sh`).

## Release (planned)

Follow colflow: GoReleaser on `v*` tag → GitHub release (darwin/linux × amd64/arm64) → updates `CogappLabs/homebrew-tap`. Not yet wired.

## Command file layout

Each top-level group is one file returning a `*cli.Command`. Wave-2 read
commands live in their own files, and their GraphQL wrappers live in
`internal/client/queries_<domain>.go` (not the shared `queries.go`) so they can
be added independently:

- `issue.go` + `issue_mutate.go` + `issue_comments.go`
- `team.go`, `user.go` (list + view)
- `project.go` (+ `queries_project.go`)
- `milestone.go` (+ `queries_milestone.go`)
- `cycle.go` (+ `queries_cycle.go`)
- `label.go` (+ `queries_label.go`, list + create)
- `document.go` (+ `queries_document.go`)
- `config.go`

Shared `queries.go` holds issues, teams, users, states, and the `ListUsers`
pagination helper. `ListUsers` pages through the full connection (250/page) —
the workspace has >50 members, so a single page silently truncates. Anything
resolving a user by email (`GetUser`, `resolveUserID`) depends on that.

## Available commands

`auth login|status`, `whoami`, `issue list|view|create|update|start|comment|comments|url|branch`, `team list|view`, `user list|view`, `project list|view`, `milestone list|view`, `cycle list`, `label list|create`, `document list|view`, `config set-team|show`.
