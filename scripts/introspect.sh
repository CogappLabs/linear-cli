#!/usr/bin/env bash
# Introspect selected Linear GraphQL types and their fields.
# Usage: LINEAR_API_KEY=... ./scripts/introspect.sh [TypeName ...]
# Defaults to the core types the CLI touches.
set -euo pipefail

: "${LINEAR_API_KEY:?set LINEAR_API_KEY}"
ENDPOINT="https://api.linear.app/graphql"

types=("$@")
if [ ${#types[@]} -eq 0 ]; then
  types=(Issue Comment Team User WorkflowState Project ProjectMilestone Cycle IssueLabel Attachment Document IssueCreateInput IssueUpdateInput IssueFilter Query Mutation)
fi

for t in "${types[@]}"; do
  echo ">> introspecting $t" >&2
  q=$(printf '{"query":"query($n:String!){__type(name:$n){name kind description fields{name type{name kind ofType{name kind ofType{name kind}}}} inputFields{name type{name kind ofType{name kind ofType{name kind}}}}}}","variables":{"n":"%s"}}' "$t")
  curl -s "$ENDPOINT" -H "Authorization: $LINEAR_API_KEY" -H "Content-Type: application/json" -d "$q"
  echo
done
