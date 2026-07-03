# Documentation language inventory script

This script supports Phase 0.2 — Italian documentation inventory.

It scans Markdown files under `docs/` for a curated set of Italian terms and generates:

```text
docs/italian-documentation-inventory.md
```

## Usage

From the repository root:

```bash
python3 scripts/docs/inventory-italian-content.py
```

Then validate:

```bash
ls -l docs/italian-documentation-inventory.md

git diff -- docs/italian-documentation-inventory.md scripts/docs/inventory-italian-content.py scripts/docs/README.md

git diff --check
```

## Notes

The scan is heuristic. The generated report must be reviewed by maintainers before translation work starts.

The script is compatible with older Python 3 environments and does not require `from __future__ import annotations` or `dataclasses`.
