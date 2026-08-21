"""
Manga source API client for https://komikcast.app.
"""

import html
import json
import re
import time
from typing import Optional, Any
import requests


class MangaSourceAPI:
    BASE_URL = "https://komikcast.app"

    def __init__(self, base_url: Optional[str] = None, timeout: int = 60, retries: int = 3):
        if base_url:
            self.BASE_URL = base_url.rstrip("/")
        self.session = requests.Session()
        self.session.headers.update({
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
                          "(KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
            "Accept": "application/json, text/plain, */*",
        })
        self.timeout = timeout
        self.retries = retries
        self._inertia_version: Optional[str] = None
        self._slug_cache: dict[int, str] = {}
        self._genres_cache: Optional[list] = None

    def _ensure_version(self) -> str:
        if self._inertia_version:
            return self._inertia_version
        try:
            resp = self.session.get(f"{self.BASE_URL}/", timeout=self.timeout)
            match = re.search(r'data-page="([^"]+)"', resp.text)
            if match:
                data = json.loads(html.unescape(match.group(1)))
                self._inertia_version = data.get("version")
                return self._inertia_version or ""
        except Exception:
            pass
        return ""

    def _get_page(self, path: str, params: Optional[dict] = None) -> dict:
        url = f"{self.BASE_URL}{path}"
        last_error: Exception | None = None

        for attempt in range(1, self.retries + 1):
            version = self._ensure_version()
            headers = {
                "X-Inertia": "true",
                "X-Inertia-Version": version,
                "X-Requested-With": "XMLHttpRequest",
            }
            try:
                resp = self.session.get(url, params=params, headers=headers, timeout=self.timeout)

                # If Inertia returns 409, the version changed -> refresh version and retry
                if resp.status_code == 409:
                    self._inertia_version = None
                    self._ensure_version()
                    continue

                if resp.status_code not in {408, 429, 500, 502, 503, 504}:
                    resp.raise_for_status()
                    try:
                        return resp.json()
                    except Exception:
                        # Fallback parse HTML data-page
                        match = re.search(r'data-page="([^"]+)"', resp.text)
                        if match:
                            data = json.loads(html.unescape(match.group(1)))
                            if data.get("version"):
                                self._inertia_version = data.get("version")
                            return data
                        raise RuntimeError(f"Failed to parse Inertia JSON from {url}")

                last_error = RuntimeError(f"source API HTTP {resp.status_code}: {resp.text[:300]}")
            except (requests.Timeout, requests.ConnectionError) as exc:
                last_error = exc

            if attempt < self.retries:
                sleep_for = min(2 ** attempt, 10)
                print(f"source API request failed, retrying in {sleep_for}s ({attempt}/{self.retries}): {last_error}")
                time.sleep(sleep_for)

        if last_error:
            raise last_error
        raise RuntimeError("source API request failed")

    def _format_series_item(self, item: dict) -> dict:
        item_id = item.get("id")
        slug = item.get("slug") or ""
        if item_id and slug:
            self._slug_cache[item_id] = slug

        last_ch = item.get("last_chapter")
        if isinstance(last_ch, dict):
            total_ch = last_ch.get("chapter_number", 0)
        else:
            total_ch = len(item.get("chapters", []))

        genres = []
        for g in item.get("genres", []) or []:
            if isinstance(g, dict):
                genres.append({
                    "id": g.get("id"),
                    "data": {
                        "id": g.get("id"),
                        "name": g.get("name") or "",
                        "slug": g.get("slug") or "",
                    }
                })

        data_payload = {
            "id": item_id,
            "slug": slug,
            "title": item.get("title") or slug,
            "nativeTitle": item.get("title") or "",
            "author": item.get("author") or "",
            "artist": item.get("artist") or "",
            "status": item.get("status") or "",
            "type": item.get("type") or "",
            "format": item.get("type") or "",
            "rating": str(item.get("rating") or ""),
            "totalChapters": total_ch,
            "synopsis": item.get("synopsis") or "",
            "coverImage": item.get("poster") or "",
            "genres": genres,
            "releaseYear": item.get("release_year"),
            "viewsCount": item.get("views_count", 0),
        }

        return {
            "id": item_id,
            "data": data_payload,
        }

    # ── Series / Manga ────────────────────────────────────────────────

    def list_series(self, page: int = 1) -> dict:
        """Get paginated list of all manga series."""
        raw = self._get_page("/manga", {"page": page})
        props = raw.get("props", {})
        if "genres" in props and not self._genres_cache:
            self._genres_cache = props.get("genres")

        mangas = props.get("mangas", {})
        data_items = mangas.get("data", [])
        return {
            "data": [self._format_series_item(it) for it in data_items],
            "meta": {
                "page": mangas.get("current_page", page),
                "lastPage": mangas.get("last_page", 1),
                "total": mangas.get("total", len(data_items)),
                "perPage": mangas.get("per_page", 24),
            }
        }

    def _resolve_slug(self, series_id_or_slug: int | str) -> str:
        if isinstance(series_id_or_slug, str) and not str(series_id_or_slug).isdigit():
            return series_id_or_slug
        sid = int(series_id_or_slug)
        if sid in self._slug_cache:
            return self._slug_cache[sid]
        # Search or list to find slug
        res = self.list_series(1)
        if sid in self._slug_cache:
            return self._slug_cache[sid]
        raise RuntimeError(f"Cannot resolve slug for series ID: {series_id_or_slug}")

    def get_series(self, series_id_or_slug: int | str) -> dict:
        """Get a single series by its slug or numeric ID."""
        slug = self._resolve_slug(series_id_or_slug)
        raw = self._get_page(f"/manga/{slug}")
        props = raw.get("props", {})
        manga = props.get("manga")
        if not manga:
            raise RuntimeError(f"Manga not found for slug: {slug}")

        formatted = self._format_series_item(manga)
        return {
            "id": formatted["id"],
            "data": {
                "id": formatted["id"],
                "data": {
                    **formatted["data"],
                    "chapters": manga.get("chapters", []),
                }
            }
        }

    def search_series(self, query: str) -> dict:
        """Search for a series by its slug or name."""
        clean_query = query.strip()
        results: list[dict] = []
        seen_ids: set[int] = set()

        # 1. If it looks like a direct slug, try fetching it directly
        slug_candidate = re.sub(r"[^a-zA-Z0-9_-]", "", clean_query).lower()
        if slug_candidate:
            try:
                direct = self._get_page(f"/manga/{slug_candidate}")
                manga = direct.get("props", {}).get("manga")
                if manga and manga.get("id"):
                    seen_ids.add(manga["id"])
                    results.append(self._format_series_item(manga))
            except Exception:
                pass

        # 2. Query the search endpoint with original query and spaced version
        search_terms = [clean_query]
        if "-" in clean_query or "_" in clean_query:
            spaced = clean_query.replace("-", " ").replace("_", " ").strip()
            if spaced not in search_terms:
                search_terms.append(spaced)

        for term in search_terms:
            try:
                raw = self._get_page("/manga", {"search": term})
                props = raw.get("props", {})
                mangas = props.get("mangas", {})
                data_items = mangas.get("data", [])
                for it in data_items:
                    it_id = it.get("id")
                    if it_id and it_id not in seen_ids:
                        seen_ids.add(it_id)
                        results.append(self._format_series_item(it))
            except Exception:
                pass

        return {
            "data": results,
            "meta": {
                "page": 1,
                "lastPage": 1,
                "total": len(results),
                "perPage": len(results) or 24,
            }
        }

    # ── Chapters ──────────────────────────────────────────────────────

    def list_chapters(self, series_id_or_slug: int | str, page: int = 1) -> dict:
        """List chapters for a series by its slug or ID."""
        series = self.get_series(series_id_or_slug)
        chapters = series["data"]["data"].get("chapters", [])

        formatted_chapters = []
        for ch in chapters:
            ch_num = ch.get("chapter_number")
            formatted_chapters.append({
                "id": ch.get("id"),
                "data": {
                    "id": ch.get("id"),
                    "index": ch_num,
                    "slug": ch.get("slug") or f"chapter-{ch_num}",
                    "title": ch.get("title") or f"Chapter {ch_num}",
                },
                "views": {
                    "total": ch.get("views_count", 0),
                }
            })

        return {
            "data": formatted_chapters,
            "meta": {
                "page": 1,
                "lastPage": 1,
                "total": len(formatted_chapters),
            }
        }

    def get_chapter(self, series_slug: str, chapter_index: int | str) -> dict:
        """
        Get a single chapter with all page images.

        Parameters
        ----------
        series_slug : str
            The series slug (e.g. 'ao-ashi').
        chapter_index : int or str
            The chapter number / index (e.g. 1, '4.2', 'chapter-1').
        """
        ch_str = str(chapter_index).strip()
        if ch_str.startswith("chapter-"):
            ch_str = ch_str.replace("chapter-", "")
        if "." in ch_str:
            ch_str = ch_str.rstrip("0").rstrip(".")

        raw = self._get_page(f"/manga/{series_slug}/chapter/{ch_str}")
        props = raw.get("props", {})
        chapter = props.get("chapter")
        if not chapter:
            raise RuntimeError(f"Chapter '{chapter_index}' not found for '{series_slug}'")

        images_raw = chapter.get("images", [])
        images = []
        for img in images_raw:
            if isinstance(img, dict) and "image_path" in img:
                images.append(img["image_path"])
            elif isinstance(img, str):
                images.append(img)

        return {
            "data": {
                "data": {
                    "id": chapter.get("id"),
                    "title": chapter.get("title") or f"Chapter {ch_str}",
                    "slug": chapter.get("slug") or f"chapter-{ch_str}",
                    "index": chapter.get("chapter_number", ch_str),
                    "images": images,
                }
            }
        }

    def get_chapter_by_slug(self, series_slug: str, chapter_slug: str) -> dict:
        """Get a single chapter by its slug."""
        return self.get_chapter(series_slug, chapter_slug)

    # ── Genres ────────────────────────────────────────────────────────

    def list_genres(self) -> dict:
        """Get all available genres."""
        if not self._genres_cache:
            self.list_series(1)
        genres = self._genres_cache or []
        return {
            "data": [
                {
                    "id": g.get("id"),
                    "data": {
                        "id": g.get("id"),
                        "name": g.get("name"),
                        "slug": g.get("slug"),
                    }
                }
                for g in genres
            ]
        }

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

    def all_chapters(self, series_id_or_slug: int | str):
        """
        Generator that yields every chapter for a series.

        Parameters
        ----------
        series_id_or_slug : int or str
            Series ID or slug.
        """
        first = self.list_chapters(series_id_or_slug, 1)
        for ch in first.get("data", []):
            yield ch
