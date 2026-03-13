import json
import logging
import re
from urllib.parse import urljoin, urlparse

import trafilatura
from lxml import etree
from lxml import html as lxml_html

from python_parser_api.models.content_block import ContentBlock
from python_parser_api.models.content_block_type import ContentBlockType
from python_parser_api.models.link import Link
from python_parser_api.models.link_type import LinkType
from python_parser_api.models.scrapped_page import ScrappedPage

logger = logging.getLogger("app.parser")

_HEAD_REND_RE = re.compile(r"h(\d)")
_TEI_NS = {"tei": "http://www.tei-c.org/ns/1.0"}


def _extract_metadata(html: str, url: str) -> dict:
    raw = trafilatura.extract(
        html,
        url=url,
        output_format="json",
        with_metadata=True,
        include_comments=False,
    )
    if not raw:
        return {}
    try:
        data = json.loads(raw)
    except json.JSONDecodeError:
        return {}
    return {
        "title": data.get("title", ""),
        "description": data.get("description"),
        "language": data.get("language"),
    }


def _extract_blocks(html: str, url: str) -> list[ContentBlock]:
    tei_xml = trafilatura.extract(
        html,
        url=url,
        output_format="xmltei",
        include_comments=False,
        include_tables=True,
    )
    if not tei_xml:
        return []

    try:
        root = etree.fromstring(tei_xml.encode("utf-8"))
    except etree.XMLSyntaxError:
        logger.warning("failed to parse TEI XML")
        return []

    body = root.find(".//tei:body", _TEI_NS)
    if body is None:
        return []

    blocks: list[ContentBlock] = []
    idx = 0

    for elem in body.iter():
        tag = etree.QName(elem.tag).localname if isinstance(elem.tag, str) else None
        if tag is None:
            continue

        text = "".join(elem.itertext()).strip()
        if not text:
            continue

        if tag == "head":
            rend = elem.get("rend", "h2")
            m = _HEAD_REND_RE.search(rend)
            level = int(m.group(1)) if m else 2
            blocks.append(ContentBlock(
                html_id=f"blk-{idx}",
                type=ContentBlockType.HEADING,
                heading_level=level,
                text=text,
            ))
            idx += 1
        elif tag == "p":
            blocks.append(ContentBlock(
                html_id=f"blk-{idx}",
                type=ContentBlockType.PARAGRAPH,
                text=text,
            ))
            idx += 1
        elif tag == "item":
            blocks.append(ContentBlock(
                html_id=f"blk-{idx}",
                type=ContentBlockType.LISTITEM,
                text=text,
            ))
            idx += 1

    return blocks


def _extract_links(html: str, url: str) -> list[Link]:
    try:
        tree = lxml_html.fromstring(html)
    except etree.Error:
        return []

    source_domain = urlparse(url).netloc
    links: list[Link] = []
    seen: set[str] = set()

    for a in tree.iter("a"):
        href = a.get("href")
        if not href or href.startswith("#"):
            continue

        abs_href = urljoin(url, href)
        parsed = urlparse(abs_href)
        if parsed.scheme not in ("http", "https"):
            continue

        clean = parsed._replace(fragment="").geturl()
        if clean in seen:
            continue
        seen.add(clean)

        links.append(Link(
            href=abs_href,
            type=LinkType.INTERNAL if parsed.netloc == source_domain else LinkType.EXTERNAL,
            anchor_text=a.text_content().strip() or None,
            position=0,
        ))

    return links


def parse(html: str, url: str) -> ScrappedPage:
    metadata = _extract_metadata(html, url)
    blocks = _extract_blocks(html, url)
    links = _extract_links(html, url)

    return ScrappedPage(
        url=url,
        title=metadata.get("title", ""),
        description=metadata.get("description"),
        language=metadata.get("language"),
        blocks=blocks,
        links=links if links else None,
    )