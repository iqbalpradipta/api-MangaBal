"""
Manga source API client.
"""

import requests
from typing import Optional


class MangaSourceAPI:
    BASE_URL = "https://be.komikcast.cc"

    def __init__(self, timeout: int = 30):
        self.session = requests.Session()
        self.session.headers.update({
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
                          "(KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
            "Accept": "application/json",
        })
        self.timeout = timeout

    def _get(self, path: str, params: Optional[dict] = None) -> dict:
        url = f"{self.BASE_URL}{path}"
        resp = self.session.get(url, params=params, timeout=self.timeout)
        resp.raise_for_status()
        return resp.json()

    # ── Series / Manga ────────────────────────────────────────────────

    def list_series(self, page: int = 1) -> dict:
        """Get paginated list of all manga series."""
        return self._get("/series", {"page": page})

    def get_series(self, series_id: int) -> dict:
        """Get a single series by its numeric ID."""
        return self._get(f"/series/{series_id}")

    def search_series(self, slug: str) -> dict:
        """Search for a series by its slug."""
        return self._get("/series", {"slug": slug})

    # ── Chapters ──────────────────────────────────────────────────────

    def list_chapters(self, series_id: int, page: int = 1) -> dict:
        """List chapters for a series by its numeric ID."""
        return self._get(f"/series/{series_id}/chapters", {"page": page})

    def get_chapter(self, series_slug: str, chapter_index: int) -> dict:
        """
        Get a single chapter with all page images.

        Parameters
        ----------
        series_slug : str
            The series slug (e.g. 'mumumu').
        chapter_index : int
            The chapter number / index (e.g. 1).
        """
        return self._get(f"/series/{series_slug}/chapters/{chapter_index}")

    def get_chapter_by_slug(self, series_slug: str, chapter_slug: str) -> dict:
        """Get a single chapter by its slug."""
        return self._get(f"/series/{series_slug}/chapters/{chapter_slug}")

    # ── Genres ────────────────────────────────────────────────────────

    def list_genres(self) -> dict:
        """Get all available genres."""
        return self._get("/genres")

    # ── Helpers ───────────────────────────────────────────────────────

    def all_series(self, max_pages: Optional[int] = None):
        """
        Generator that yields every series across all pages.

        Parameters
        ----------
        max_pages : int, optional
            Limit the number of pages to fetch.
        """
        first = self.list_series(1)
        total_pages = first["meta"]["lastPage"]
        if max_pages:
            total_pages = min(total_pages, max_pages)

        for item in first["data"]:
            yield item

        for p in range(2, total_pages + 1):
            page = self.list_series(p)
            for item in page["data"]:
                yield item

    def all_chapters(self, series_id: int):
        """
        Generator that yields every chapter for a series.

        Parameters
        ----------
        series_id : int
            Numeric series ID.
        """
        first = self.list_chapters(series_id, 1)
        if not first.get("data"):
            return

        for ch in first["data"]:
            yield ch

        meta = first.get("meta", {})
        total_pages = meta.get("lastPage", 1)
        for p in range(2, total_pages + 1):
            page = self.list_chapters(series_id, p)
            for ch in page.get("data", []):
                yield ch
