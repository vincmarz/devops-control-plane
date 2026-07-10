# Final Technical Guide

This directory contains the source material for the final technical guide of the DevOps Control Plane project.

The guide is written in Italian. Commands, resource names, API names, file paths and technical identifiers remain in their original format.

## Files

- `outline.md` — final guide structure.
- `source-map.md` — mapping between guide chapters and source repository documents.
- `writing-plan.md` — writing strategy, quality rules and review criteria.
- `final-technical-guide.md` — main source document for the final guide.
- `outputs/` — generated deliverables such as Word or PDF files, if versioned.

## Source of truth

The Markdown file `final-technical-guide.md` is the primary source of truth.

Generated Word or PDF files must be treated as derived outputs.

<!-- FINAL_GUIDE_STATUS_START -->
## Stato finale della guida tecnica

La guida tecnica finale del DevOps Control Plane e stata completata come sorgente Markdown e generata anche in formato Word.

### Sorgente primaria

Il documento sorgente principale resta il Markdown versionato nel repository:

```text
docs/final-technical-guide/final-technical-guide.md
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

- 43 capitoli numerati;
- Appendice A - Glossario;
- Appendice B - Comandi operativi principali;
- Appendice C - Commit e tag rilevanti;
- Appendice D - Limitazioni note.

La numerazione dei capitoli e stata revisionata e riallineata con la source map.

### Controlli finali eseguiti

Sono stati completati i controlli strutturali finali:

```text
12.14.3 Final structural quality checks
Status: PASS
Errors: none
Warnings: none
```

Evidence directory del controllo:

```text
/tmp/dcp-final-guide-12-14-3-20260710-130040
```

Il controllo ha verificato:

- presenza dei file principali;
- working tree pulito all'avvio del check;
- numerazione guida sequenziale 1..43;
- presenza delle appendici A/B/C/D;
- source-map sequenziale 1..43;
- allineamento tra guida e source-map;
- assenza di pattern di secret reali;
- presenza delle frasi chiave sulla readiness multi-cluster e sulla validazione fisica deferred.

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

1. aggiornare `docs/final-technical-guide/final-technical-guide.md`;
2. aggiornare `source-map.md` se cambiano capitoli o fonti;
3. rieseguire i controlli strutturali;
4. rigenerare il Word in `outputs/`;
5. committare Markdown, output e report aggiornati.
<!-- FINAL_GUIDE_STATUS_END -->
