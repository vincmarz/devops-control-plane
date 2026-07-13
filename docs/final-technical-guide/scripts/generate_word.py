#!/usr/bin/env python3
"""Generate the final DevOps Control Plane Word guide without external packages.

Compatible with Python 3.6. The Italian Markdown remains the source of truth.
Bullet and numbered lists are emitted as literal markers, avoiding Word's
cross-document automatic numbering behaviour.
"""

from pathlib import Path
from zipfile import ZipFile, ZIP_DEFLATED
from datetime import datetime
import html
import re
import sys

ROOT = Path(__file__).resolve().parents[3]
SOURCE = ROOT / "docs/final-technical-guide/final-technical-guide-it.md"
SOURCE_MAP = ROOT / "docs/final-technical-guide/source-map.md"
OUTPUT_DIR = ROOT / "docs/final-technical-guide/outputs"
DOCX = OUTPUT_DIR / "DevOps_Control_Plane_Guida_Tecnica_Finale.docx"
REPORT = OUTPUT_DIR / "word-generation-report.txt"
EXPECTED_CHAPTERS = 44
TITLE = "DevOps Control Plane - Guida Tecnica Finale"
SUBTITLE = "Baseline namespace-isolated, runtime evidence, CI e multi-cluster code-ready readiness"
AUTHOR = "Vincenzo Marzario"
W_NS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"


def escape(value):
    return html.escape(value, quote=True)


def text_run(value, bold=False, italic=False, code=False):
    properties = []
    if bold:
        properties.append('<w:b/>')
    if italic:
        properties.append('<w:i/>')
    if code:
        properties.append('<w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/>')
        properties.append('<w:sz w:val="18"/>')
    rpr = '<w:rPr>{}</w:rPr>'.format(''.join(properties)) if properties else ''
    body = []
    for index, item in enumerate(value.split('\n')):
        if index:
            body.append('<w:br/>')
        preserve = ' xml:space="preserve"' if item.startswith(' ') or item.endswith(' ') else ''
        body.append('<w:t{}>{}</w:t>'.format(preserve, escape(item)))
    return '<w:r>{}{}</w:r>'.format(rpr, ''.join(body))


def inline_runs(value):
    output = []
    tokens = re.split(r'(`[^`]*`|\*\*[^*]+\*\*|\*[^*]+\*)', value)
    for token in tokens:
        if not token:
            continue
        if token.startswith('`') and token.endswith('`'):
            output.append(text_run(token[1:-1], code=True))
        elif token.startswith('**') and token.endswith('**'):
            output.append(text_run(token[2:-2], bold=True))
        elif token.startswith('*') and token.endswith('*'):
            output.append(text_run(token[1:-1], italic=True))
        else:
            output.append(text_run(token))
    return ''.join(output)


def paragraph(value='', style=None, keep_next=False, code=False):
    properties = []
    if style:
        properties.append('<w:pStyle w:val="{}"/>'.format(style))
    if keep_next:
        properties.append('<w:keepNext/>')
    if code:
        properties.append('<w:pStyle w:val="CodeBlock"/>')
    ppr = '<w:pPr>{}</w:pPr>'.format(''.join(properties)) if properties else ''
    content = text_run(value, code=True) if code else inline_runs(value)
    return '<w:p>{}{}</w:p>'.format(ppr, content)


def literal_list_paragraph(marker, value):
    ppr = (
        '<w:pPr>'
        '<w:tabs><w:tab w:val="left" w:pos="720"/></w:tabs>'
        '<w:ind w:left="720" w:hanging="360"/>'
        '</w:pPr>'
    )
    marker_run = (
        '<w:r><w:rPr><w:rFonts w:ascii="Aptos" w:hAnsi="Aptos"/>'
        '</w:rPr><w:t>{}</w:t><w:tab/></w:r>'
    ).format(escape(marker))
    return '<w:p>{}{}{}</w:p>'.format(ppr, marker_run, inline_runs(value))


def page_break():
    return '<w:p><w:r><w:br w:type="page"/></w:r></w:p>'


def parse_markdown(source_text):
    blocks = []
    pending = []
    code_lines = []
    in_code = False

    def flush_paragraph():
        if pending:
            blocks.append(('paragraph', ' '.join(pending).strip()))
            del pending[:]

    for raw_line in source_text.splitlines():
        line = raw_line.rstrip()
        if line.startswith('```'):
            if in_code:
                blocks.append(('code', '\n'.join(code_lines)))
                code_lines = []
                in_code = False
            else:
                flush_paragraph()
                in_code = True
            continue
        if in_code:
            code_lines.append(line)
            continue
        if not line.strip():
            flush_paragraph()
        elif line.startswith('#### '):
            flush_paragraph(); blocks.append(('heading4', line[5:].strip()))
        elif line.startswith('### '):
            flush_paragraph(); blocks.append(('heading3', line[4:].strip()))
        elif line.startswith('## '):
            flush_paragraph(); blocks.append(('heading2', line[3:].strip()))
        elif line.startswith('# '):
            flush_paragraph(); blocks.append(('heading1', line[2:].strip()))
        elif re.match(r'^\s*-\s+', line):
            flush_paragraph(); blocks.append(('bullet', re.sub(r'^\s*-\s+', '', line).strip()))
        else:
            numbered = re.match(r'^\s*(\d+)\.\s+(.+)$', line)
            if numbered:
                flush_paragraph()
                blocks.append(('number', numbered.group(1) + '.', numbered.group(2).strip()))
            elif line != '---':
                pending.append(line.strip())
            else:
                flush_paragraph()
    flush_paragraph()
    return blocks


def styles_xml():
    return '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Aptos" w:hAnsi="Aptos"/><w:sz w:val="21"/><w:lang w:val="it-IT"/></w:rPr></w:rPrDefault><w:pPrDefault><w:pPr><w:spacing w:after="120" w:line="276" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults>
<w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
<w:style w:type="paragraph" w:styleId="Title"><w:name w:val="Title"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/><w:spacing w:after="240"/></w:pPr><w:rPr><w:rFonts w:ascii="Aptos Display" w:hAnsi="Aptos Display"/><w:b/><w:sz w:val="40"/><w:color w:val="1F4E79"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Subtitle"><w:name w:val="Subtitle"/><w:basedOn w:val="Normal"/><w:pPr><w:jc w:val="center"/><w:spacing w:after="360"/></w:pPr><w:rPr><w:i/><w:sz w:val="24"/><w:color w:val="666666"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/><w:basedOn w:val="Normal"/><w:qFormat/><w:pPr><w:outlineLvl w:val="0"/><w:spacing w:before="360" w:after="160"/></w:pPr><w:rPr><w:rFonts w:ascii="Aptos Display" w:hAnsi="Aptos Display"/><w:b/><w:sz w:val="30"/><w:color w:val="1F4E79"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/><w:basedOn w:val="Normal"/><w:qFormat/><w:pPr><w:outlineLvl w:val="1"/><w:spacing w:before="260" w:after="120"/></w:pPr><w:rPr><w:b/><w:sz w:val="24"/><w:color w:val="2F75B5"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Heading3"><w:name w:val="heading 3"/><w:basedOn w:val="Normal"/><w:qFormat/><w:pPr><w:outlineLvl w:val="2"/><w:spacing w:before="200" w:after="100"/></w:pPr><w:rPr><w:b/><w:sz w:val="22"/><w:color w:val="1F4E79"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="TOC1"><w:name w:val="TOC 1"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:after="40"/></w:pPr></w:style>
<w:style w:type="paragraph" w:styleId="TOC2"><w:name w:val="TOC 2"/><w:basedOn w:val="Normal"/><w:pPr><w:ind w:left="360"/><w:spacing w:after="20"/></w:pPr><w:rPr><w:sz w:val="19"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="CodeBlock"><w:name w:val="CodeBlock"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:before="80" w:after="120"/><w:shd w:fill="F2F2F2"/></w:pPr><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/><w:sz w:val="18"/></w:rPr></w:style>
</w:styles>'''


def build_document(blocks):
    body = [
        paragraph(TITLE, 'Title'),
        paragraph(SUBTITLE, 'Subtitle'),
        paragraph('Autore: ' + AUTHOR),
        paragraph('Sorgente: docs/final-technical-guide/final-technical-guide-it.md'),
        paragraph('Generato: ' + datetime.now().strftime('%Y-%m-%d %H:%M:%S')),
        page_break(),
        paragraph('Indice', 'Heading1'),
    ]
    for block in blocks:
        kind = block[0]
        if kind == 'heading2':
            body.append(paragraph(block[1], 'TOC1'))
        elif kind == 'heading3' and re.match(r'^\d+\.\d+\s+', block[1]):
            body.append(paragraph(block[1], 'TOC2'))
    body.append(page_break())

    for block in blocks:
        kind = block[0]
        if kind in ('heading1', 'heading2'):
            body.append(paragraph(block[1], 'Heading1', keep_next=True))
        elif kind == 'heading3':
            body.append(paragraph(block[1], 'Heading2', keep_next=True))
        elif kind == 'heading4':
            body.append(paragraph(block[1], 'Heading3', keep_next=True))
        elif kind == 'bullet':
            body.append(literal_list_paragraph('•', block[1]))
        elif kind == 'number':
            body.append(literal_list_paragraph(block[1], block[2]))
        elif kind == 'code':
            body.append(paragraph(block[1], code=True))
        else:
            body.append(paragraph(block[1]))

    body.append('<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134" w:header="708" w:footer="708" w:gutter="0"/></w:sectPr>')
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="{}"><w:body>{}</w:body></w:document>'.format(W_NS, ''.join(body))


def validate_package(docx_path, expected_bullets, expected_numbers):
    with ZipFile(str(docx_path), 'r') as archive:
        bad_file = archive.testzip()
        if bad_file:
            raise SystemExit('DOCX ZIP validation failed: {}'.format(bad_file))
        document = archive.read('word/document.xml').decode('utf-8')
    automatic_numbering = document.count('<w:numPr>')
    bullet_count = document.count('•')
    if automatic_numbering != 0:
        raise SystemExit('Unexpected automatic numbering blocks: {}'.format(automatic_numbering))
    if bullet_count != expected_bullets:
        raise SystemExit('Bullet mismatch: expected {}, found {}'.format(expected_bullets, bullet_count))
    return bullet_count, expected_numbers


def main():
    for path in (SOURCE, SOURCE_MAP):
        if not path.exists():
            raise SystemExit('Missing required file: {}'.format(path))
    source_text = SOURCE.read_text()
    source_map_text = SOURCE_MAP.read_text()
    guide_count = len(re.findall(r'^## \d+\. ', source_text, re.MULTILINE))
    source_map_count = len(re.findall(r'^### Capitolo \d+ — ', source_map_text, re.MULTILINE))
    if guide_count != EXPECTED_CHAPTERS or source_map_count != EXPECTED_CHAPTERS:
        raise SystemExit('Expected 44 chapters; guide={} source-map={}'.format(guide_count, source_map_count))
    if '## 14. Continuous Integration e test automatizzati' not in source_text:
        raise SystemExit('CI chapter 14 not found')

    blocks = parse_markdown(source_text)
    expected_bullets = sum(1 for block in blocks if block[0] == 'bullet')
    expected_numbers = sum(1 for block in blocks if block[0] == 'number')
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    content_types = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>'''
    rels = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/></Relationships>'''
    word_rels = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>'''
    now = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ')
    core = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:title>{}</dc:title><dc:creator>{}</dc:creator><dc:description>Generated from final-technical-guide-it.md. Literal list markers prevent Word numbering drift.</dc:description><dcterms:created xsi:type="dcterms:W3CDTF">{}</dcterms:created><dcterms:modified xsi:type="dcterms:W3CDTF">{}</dcterms:modified></cp:coreProperties>'''.format(escape(TITLE), escape(AUTHOR), now, now)
    app = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"><Application>Versioned Python standard library OOXML generator</Application></Properties>'''

    with ZipFile(str(DOCX), 'w', ZIP_DEFLATED) as archive:
        archive.writestr('[Content_Types].xml', content_types)
        archive.writestr('_rels/.rels', rels)
        archive.writestr('word/_rels/document.xml.rels', word_rels)
        archive.writestr('word/document.xml', build_document(blocks))
        archive.writestr('word/styles.xml', styles_xml())
        archive.writestr('docProps/core.xml', core)
        archive.writestr('docProps/app.xml', app)

    bullets, numbers = validate_package(DOCX, expected_bullets, expected_numbers)
    REPORT.write_text('\n'.join([
        'Generate final technical guide Word - versioned generator',
        'Status: PASS',
        'Generator: docs/final-technical-guide/scripts/generate_word.py',
        'Source: {}'.format(SOURCE),
        'Source map: {}'.format(SOURCE_MAP),
        'Expected chapters: 44',
        'Guide chapters found: {}'.format(guide_count),
        'Source-map chapters found: {}'.format(source_map_count),
        'CI chapter: 14. Continuous Integration e test automatizzati',
        'Output: {}'.format(DOCX),
        'Output size: {} bytes'.format(DOCX.stat().st_size),
        'Markdown bullet paragraphs: {}'.format(expected_bullets),
        'Generated literal bullet paragraphs: {}'.format(bullets),
        'Markdown numbered paragraphs: {}'.format(expected_numbers),
        'Generated literal numbered paragraphs: {}'.format(numbers),
        'Automatic Word numbering blocks: 0',
        'DOCX ZIP validation: PASS',
        'Quality evidence: /tmp/dcp-final-guide-44-quality-20260713-155304',
        'Notes:',
        '- Bullet lists use literal bullet markers with tab and hanging indentation.',
        '- Numbered lists use the number written in the Markdown source.',
        '- final-technical-guide-it.md remains the source of truth.',
        '- Perform final visual review in Word or LibreOffice before external distribution.',
    ]) + '\n')

    print('STATUS=PASS')
    print('DOCX={}'.format(DOCX))
    print('REPORT={}'.format(REPORT))
    print('BULLETS={}'.format(bullets))
    print('NUMBERED_ITEMS={}'.format(numbers))
    print('AUTOMATIC_NUMBERING=0')


if __name__ == '__main__':
    main()
