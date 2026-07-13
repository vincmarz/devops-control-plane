# Final Technical Guide

This directory contains the source material for the final technical guide of the DevOps Control Plane project.

The guide is written in Italian. Commands, resource names, API names, file paths and technical identifiers remain in their original format.

## Files

- `outline.md` — final guide structure.
- `source-map.md` — mapping between guide chapters and source repository documents.
- `writing-plan.md` — writing strategy, quality rules and review criteria.
- `final-technical-guide-it.md` — main source document for the final guide (Italian).
- `outputs/` — generated deliverables such as Word or PDF files, if versioned.

## Source of truth

The Markdown file `final-technical-guide-it.md` is the primary source of truth.

Generated Word or PDF files must be treated as derived outputs.

<!-- FINAL_GUIDE_STATUS_START -->
## Stato finale della guida tecnica

La guida tecnica finale del DevOps Control Plane e stata completata come sorgente Markdown e generata anche in formato Word.

### Sorgente primaria

Il documento sorgente principale resta il Markdown versionato nel repository:

```text
docs/final-technical-guide/final-technical-guide-it.md
```

Il Markdown e la sorgente di verita per modifiche future, revisioni, rigenerazioni e aggiornamenti documentali.

### Output Word generato

Il documento Word finale e disponibile in:

```text
docs/final-technical-guide/outputs/DevOps_Control_Plane_Guida_Tecnica_Finale.docx
```

Il report di generazione e disponibile in:

```text
docs/final-technical-guide/outputs/word-generation-report.txt
```

Il Word e un output derivato dal Markdown. In caso di modifiche future, aggiornare prima il Markdown e poi rigenerare il Word.

### Struttura validata

La guida finale contiene:

- 44 capitoli numerati;
- Appendice A - Glossario;
- Appendice B - Comandi operativi principali;
- Appendice C - Commit e tag rilevanti;
- Appendice D - Limitazioni note.

La numerazione dei capitoli e stata revisionata e riallineata con la source map.

### Controlli finali eseguiti

Sono stati completati i controlli strutturali finali:

```text
Final structural quality checks - 44 chapter CI baseline
Status: PASS
Expected chapters: 44
Errors: none
Warnings: none
```

Evidence directory del controllo:

```text
/tmp/dcp-final-guide-44-quality-20260713-155304
```

Il controllo ha verificato:

- presenza dei file principali;
- working tree pulito all'avvio del check;
- numerazione guida sequenziale 1..44;
- presenza delle appendici A/B/C/D;
- source-map sequenziale 1..44;
- allineamento tra guida e source-map;
- assenza di pattern di secret reali;
- presenza delle frasi chiave sulla readiness multi-cluster e sulla validazione fisica deferred.

### Integrazione Continuous Integration

La guida tecnica finale include ora il capitolo:

```text
14. Continuous Integration e test automatizzati
```

La fonte Markdown autorevole è:

```text
docs/continuous-integration-and-automated-testing.md
```

L'integrazione documentale ha aggiornato:

- indice documentale generale;
- outline della guida finale;
- writing plan;
- source map;
- guida italiana, ora composta da 44 capitoli;
- evidenze delle pull request CI nell'Appendice C.

Il quality check aggiornato ha verificato la presenza dei contenuti fondamentali relativi a GitHub Actions, `go vet`, race detector, coverage, integration test PostgreSQL, test HTTP end-to-end, concorrenza lifecycle, `SELECT ... FOR UPDATE`, TLS secure-by-default e required status check `test`.

Commit e merge rilevanti:

```text
170bad2 Document continuous integration and automated testing baseline
58103a9 Merge pull request #8 from vincmarz/docs/ci-automated-testing
cc504e8 Merge pull request #9 from vincmarz/docs/index-ci-baseline
23c7468 Merge pull request #10 from vincmarz/docs/final-guide-ci-planning
bb91c6c Merge pull request #11 from vincmarz/docs/final-guide-ci-source-map
c4d46b5 Merge pull request #13 from vincmarz/docs/final-guide-ci-italian-chapter
```

### Pulizia linguistica

E stata eseguita una pulizia conservativa degli accenti italiani nel documento finale.

Commit rilevanti:

```text
7883771 Clean up Italian accents in final technical guide
9b19b97 Fix remaining Italian accents in final technical guide
```

### Commit finali rilevanti

Commit principali della chiusura documento:

```text
53ba5ca Review final technical guide structure and numbering
d724df3 Align source map entries for chapters 2 and 5
7883771 Clean up Italian accents in final technical guide
9b19b97 Fix remaining Italian accents in final technical guide
54963fa Generate final technical guide Word document
```

Commit di generazione Word:

```text
54963fa Generate final technical guide Word document
```

### Nota su multi-cluster readiness

La formulazione ufficiale dello stato multi-cluster resta:

```text
Physical cross-cluster runtime validation is deferred by infrastructure availability.
Multi-cluster code readiness is completed, tested, documented and fail-closed.
```

Questa distinzione deve essere mantenuta anche in eventuali revisioni future del documento.

### Regola per modifiche future

Per modificare la guida:

1. aggiornare `docs/final-technical-guide/final-technical-guide-it.md`;
2. aggiornare `source-map.md` se cambiano capitoli o fonti;
3. rieseguire i controlli strutturali;
4. rigenerare il Word in `outputs/`;
5. committare Markdown, output e report aggiornati.
<!-- FINAL_GUIDE_STATUS_END -->
