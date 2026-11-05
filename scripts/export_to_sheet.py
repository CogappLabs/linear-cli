"""Push a `linear-cli issue export` directory into a Google Sheet.

The sheet is moved into FOLDER_ID and media files are uploaded there, with the
sheet linking to them, since a cell can't hold a video.

Usage:
    uv run --with-editable ~/git/cogapp-sheets python scripts/export_to_sheet.py EXPORT_DIR SHEET_ID FOLDER_ID
"""

import json
import mimetypes
import re
import sys
from pathlib import Path

import google.auth
from cogapp_sheets import Client
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

SCOPES = [
    "https://www.googleapis.com/auth/spreadsheets",
    "https://www.googleapis.com/auth/drive",
]
PRIORITY = {0: "", 1: "Urgent", 2: "High", 3: "Medium", 4: "Low"}
HEADER = [
    "ID",
    "Title",
    "Status",
    "Assignee",
    "Priority",
    "Labels",
    "Description",
    "Comments",
    "Media",
]


# The team's story template, left unfilled in most tickets.
TEMPLATE_RE = re.compile(
    r"(?:\n---\s*)?\s*(?:As a \\?\[user\\?\][^\n]*\s*)?"
    r"### Acceptance criteria\s*- \[ \]\s*"
    r"### Requires\s*- \[ \] Spec\s*- \[ \] Test plan\s*"
    r"### Reviews\s*- \[ \] Producer review\s*- \[ \] Code review\s*$"
)
# Cogapp Slack files and Linear uploads both need a login the reader won't have;
# the uploads are copied to Drive and linked from the Media column instead.
SLACK_FILE_RE = re.compile(r"\[[^\]]*\]\(<?https://[\w-]+\.slack\.com/files/[^)>]*>?\)")
UPLOAD_RE = re.compile(r"!?\[([^\]]*)\]\(https://uploads\.linear\.app/[^)]*\)")


def link(url: str, text: str) -> str:
    return f'=HYPERLINK("{url}", "{text.replace(chr(34), chr(39))}")'


def clean(markdown: str) -> str:
    markdown = TEMPLATE_RE.sub("", markdown)
    markdown = SLACK_FILE_RE.sub("", markdown)
    markdown = UPLOAD_RE.sub(r"\1 (see Media column)", markdown)
    return markdown.strip()


def description_cell(iss: dict) -> str:
    text = clean(iss["description"])
    if not text and iss["media"]:
        return "No description. See the recording in the Media column."
    return text


def comments_cell(iss: dict) -> str:
    return "\n\n".join(
        f"{c['author']} ({c['createdAt'][:10]}):\n{clean(c['body'])}"
        for c in sorted(iss["comments"], key=lambda c: c["createdAt"])
    )


def main(export_dir: Path, sheet_id: str, folder: str) -> None:
    issues = json.loads((export_dir / "issues.json").read_text())
    sheets = Client(sheet_id, scopes=SCOPES)
    creds, _ = google.auth.default(scopes=SCOPES)
    drive = build("drive", "v3", credentials=creds, cache_discovery=False)

    print("Moving sheet into target folder…", flush=True)
    current = (
        drive.files()
        .get(fileId=sheet_id, fields="parents", supportsAllDrives=True)
        .execute()["parents"]
    )
    if folder not in current:
        drive.files().update(
            fileId=sheet_id,
            addParents=folder,
            removeParents=",".join(current),
            supportsAllDrives=True,
        ).execute()

    rows = [HEADER]
    for iss in issues:
        media_cell = ""
        links = []
        for m in iss["media"]:
            path = export_dir / m["path"]
            print(
                f"Uploading {iss['identifier']} {m['name']} ({path.stat().st_size // 1_000_000} MB)…",
                flush=True,
            )
            mime = mimetypes.guess_type(path.name)[0] or "application/octet-stream"
            f = (
                drive.files()
                .create(
                    body={
                        "name": f"{iss['identifier']} {m['name']}",
                        "parents": [folder],
                    },
                    media_body=MediaFileUpload(
                        str(path), mimetype=mime, resumable=True
                    ),
                    fields="webViewLink",
                    supportsAllDrives=True,
                )
                .execute()
            )
            links.append((f["webViewLink"], m["name"]))
        if len(links) == 1:
            media_cell = link(*links[0])
        elif links:
            media_cell = "\n".join(url for url, _ in links)

        rows.append(
            [
                link(iss["url"], iss["identifier"]),
                iss["title"],
                iss["state"],
                iss["assignee"],
                PRIORITY.get(iss["priority"], ""),
                ", ".join(iss["labels"]),
                description_cell(iss),
                comments_cell(iss),
                media_cell,
            ]
        )

    tab = sheets.meta()["sheets"][0]["properties"]
    print(f"Writing {len(rows) - 1} rows to '{tab['title']}'…", flush=True)
    sheets.write(f"'{tab['title']}'!A1", rows)
    # Only the link columns are parsed, so descriptions starting with = or - stay text.
    sheets.write(
        f"'{tab['title']}'!A1", [[r[0]] for r in rows], value_input="USER_ENTERED"
    )
    sheets.write(
        f"'{tab['title']}'!I1", [[r[8]] for r in rows], value_input="USER_ENTERED"
    )
    sheets.batch_update(
        [
            {
                "repeatCell": {
                    "range": {"sheetId": tab["sheetId"]},
                    "cell": {
                        "userEnteredFormat": {
                            "wrapStrategy": "WRAP",
                            "verticalAlignment": "TOP",
                        }
                    },
                    "fields": "userEnteredFormat(wrapStrategy,verticalAlignment)",
                }
            },
            {
                "addTable": {
                    "table": {
                        "name": "Issues",
                        "range": {
                            "sheetId": tab["sheetId"],
                            "startRowIndex": 0,
                            "endRowIndex": len(rows),
                            "startColumnIndex": 0,
                            "endColumnIndex": len(HEADER),
                        },
                    }
                }
            },
            {
                "updateSheetProperties": {
                    "properties": {
                        "sheetId": tab["sheetId"],
                        "gridProperties": {"frozenRowCount": 1},
                    },
                    "fields": "gridProperties.frozenRowCount",
                }
            },
            *[
                {
                    "updateDimensionProperties": {
                        "range": {
                            "sheetId": tab["sheetId"],
                            "dimension": "COLUMNS",
                            "startIndex": i,
                            "endIndex": i + 1,
                        },
                        "properties": {"pixelSize": px},
                        "fields": "pixelSize",
                    }
                }
                for i, px in enumerate([90, 260, 100, 160, 80, 120, 640, 640, 220])
            ],
        ]
    )
    print("Done.", flush=True)


if __name__ == "__main__":
    main(Path(sys.argv[1]), sys.argv[2], sys.argv[3])
