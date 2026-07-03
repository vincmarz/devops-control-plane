#!/usr/bin/env python3

from collections import defaultdict
from pathlib import Path
import re

REPO_ROOT = Path.cwd()
DOCS_DIR = REPO_ROOT / "docs"
OUTPUT = DOCS_DIR / "italian-documentation-inventory.md"

TERMS = [
    "Questo", "Questa", "questo", "questa", "Obiettivo", "obiettivo",
    "Regola", "regola", "ambiente", "ambienti", "produzione", "collaudo",
    "manutenzione", "ripristino", "evidenza", "evidenze", "documentazione",
    "architettura", "integrazione", "validazione", "approvazione", "richiesta",
    "modifica", "stato", "flusso", "operativo", "operativa", "procedura",
    "sicurezza", "segreti", "creazione", "aggiornamento",
    "responsabile", "utente", "utenti", "sviluppatore", "sviluppatori",
    "promozione",
]

TERM_RE = re.compile(r"(" + "|".join(re.escape(t) for t in sorted(set(TERMS), key=len, reverse=True)) + r")")


class Match(object):
    def __init__(self, file_path, line_no, term, line):
        self.file = file_path
        self.line_no = line_no
        self.term = term
        self.line = line


def iter_markdown_files(root):
    if not root.exists():
        return []
    return sorted([p for p in root.rglob("*.md") if p.is_file()])


def scan_file(path):
    matches = []
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except UnicodeDecodeError:
        lines = path.read_text(errors="replace").splitlines()

    for idx, line in enumerate(lines, start=1):
        found = TERM_RE.search(line)
        if not found:
            continue
        matches.append(Match(path, idx, found.group(1), line.strip()))
    return matches


def priority_for(path):
    rel = path.as_posix()
    if rel.endswith("docs/00-vision.md"):
        return "P1"
    if rel.endswith("docs/01-scope-mvp.md"):
        return "P1"
    if rel.endswith("docs/05-architecture.md"):
        return "P1"
    if "/adr/" in rel or rel.endswith("docs/adr/README.md"):
        return "P1"
    if "/runbooks/" in rel:
        return "P2"
    if rel.endswith("README.md"):
        return "P2"
    return "P3"


def migration_recommendation(path):
    rel = path.as_posix()
    if rel.endswith("docs/00-vision.md"):
        return "Translate and align early vision content to English before final technical documentation."
    if rel.endswith("docs/01-scope-mvp.md"):
        return "Translate and normalize MVP scope terminology to English."
    if rel.endswith("docs/05-architecture.md"):
        return "Translate architecture narrative and align terminology with ADRs and runbooks."
    if "/adr/" in rel or rel.endswith("docs/adr/README.md"):
        return "ADR content and index must remain English; fix any residual mixed-language wording."
    if "/runbooks/" in rel:
        return "Review runbook for residual Italian terms; most recent runbooks should already be English."
    return "Review and translate if content is still relevant; otherwise mark as historical or superseded."


def md_escape(text):
    return text.replace("|", "\\|")


def main():
    if not DOCS_DIR.exists():
        raise SystemExit("docs directory not found. Run this script from the repository root.")

    all_files = list(iter_markdown_files(DOCS_DIR))
    all_matches = []
    by_file = defaultdict(list)

    for file_path in all_files:
        matches = scan_file(file_path)
        if matches:
            by_file[file_path].extend(matches)
            all_matches.extend(matches)

    files_with_matches = sorted(by_file.keys(), key=lambda p: (priority_for(p), p.as_posix()))

    lines = []
    lines.append("# DevOps Control Plane — Italian Documentation Inventory")
    lines.append("")
    lines.append("## Document metadata")
    lines.append("")
    lines.append("- **Project:** DevOps Control Plane")
    lines.append("- **Phase:** 0.2 — Italian documentation inventory")
    lines.append("- **Status:** Generated inventory baseline for documentation migration")
    lines.append("- **Scope:** Markdown files under `docs/`")
    lines.append("- **Policy reference:** `docs/documentation-language-policy.md`")
    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("## 1. Purpose")
    lines.append("")
    lines.append("This inventory identifies documentation files that may contain Italian wording or mixed-language terminology.")
    lines.append("The inventory is heuristic and must be reviewed by maintainers before migration work starts.")
    lines.append("")
    lines.append("The official repository language is English. Existing Italian documentation should be migrated incrementally according to the repository language policy.")
    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("## 2. Summary")
    lines.append("")
    lines.append("```text")
    lines.append("Markdown files scanned: {0}".format(len(all_files)))
    lines.append("Files with potential Italian content: {0}".format(len(files_with_matches)))
    lines.append("Potential Italian term matches: {0}".format(len(all_matches)))
    lines.append("```")
    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("## 3. Migration priority")
    lines.append("")
    lines.append("Priority levels:")
    lines.append("")
    lines.append("```text")
    lines.append("P1 = Core documentation or ADR content; migrate first")
    lines.append("P2 = Operational or index documentation; review after P1")
    lines.append("P3 = Supporting or historical documentation; migrate or supersede later")
    lines.append("```")
    lines.append("")
    lines.append("| Priority | File | Matches | Recommendation |")
    lines.append("|---|---|---:|---|")
    for file_path in files_with_matches:
        rel = file_path.relative_to(REPO_ROOT).as_posix()
        match_count = len(by_file[file_path])
        lines.append("| {0} | `{1}` | {2} | {3} |".format(priority_for(file_path), rel, match_count, md_escape(migration_recommendation(file_path))))
    if not files_with_matches:
        lines.append("| N/A | No files detected | 0 | No migration targets detected by heuristic scan. |")

    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("## 4. Detailed findings")
    lines.append("")
    for file_path in files_with_matches:
        rel = file_path.relative_to(REPO_ROOT).as_posix()
        lines.append("### `{0}`".format(rel))
        lines.append("")
        lines.append("- **Priority:** {0}".format(priority_for(file_path)))
        lines.append("- **Potential matches:** {0}".format(len(by_file[file_path])))
        lines.append("")
        lines.append("| Line | Term | Snippet |")
        lines.append("|---:|---|---|")
        for match in by_file[file_path][:40]:
            snippet = match.line
            if len(snippet) > 180:
                snippet = snippet[:177] + "..."
            lines.append("| {0} | `{1}` | {2} |".format(match.line_no, md_escape(match.term), md_escape(snippet)))
        if len(by_file[file_path]) > 40:
            lines.append("| ... | ... | Output truncated after 40 matches for this file. Total matches: {0}. |".format(len(by_file[file_path])))
        lines.append("")

    lines.append("---")
    lines.append("")
    lines.append("## 5. Recommended migration plan")
    lines.append("")
    lines.append("Recommended order:")
    lines.append("")
    lines.append("```text")
    lines.append("1. Complete and commit the repository language policy if not already committed.")
    lines.append("2. Translate and align docs/00-vision.md.")
    lines.append("3. Translate and align docs/01-scope-mvp.md.")
    lines.append("4. Translate and align docs/05-architecture.md.")
    lines.append("5. Review ADR index and ADR files for residual mixed-language content.")
    lines.append("6. Review remaining documents by priority.")
    lines.append("7. Keep all new documentation in English.")
    lines.append("```")
    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("## 6. Acceptance criteria for Phase 0.2")
    lines.append("")
    lines.append("Phase 0.2 is complete when:")
    lines.append("")
    lines.append("```text")
    lines.append("Italian documentation inventory has been generated.")
    lines.append("Potential migration targets are listed.")
    lines.append("Migration priority is documented.")
    lines.append("Repository language policy is referenced.")
    lines.append("No automatic translation has been applied without review.")
    lines.append("```")
    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("## 7. Revision history")
    lines.append("")
    lines.append("| Date | Phase | Description |")
    lines.append("|---|---:|---|")
    lines.append("| 2026-07-03 | 0.2 | Initial generated Italian documentation inventory baseline. |")
    lines.append("")

    OUTPUT.write_text("\n".join(lines), encoding="utf-8")
    print("Generated {0}".format(OUTPUT))
    print("Markdown files scanned: {0}".format(len(all_files)))
    print("Files with potential Italian content: {0}".format(len(files_with_matches)))
    print("Potential Italian term matches: {0}".format(len(all_matches)))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
