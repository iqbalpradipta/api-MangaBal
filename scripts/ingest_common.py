import argparse
import os
import sys
import tempfile
import time
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

import requests

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from manga_source.api import MangaSourceAPI  # noqa: E402
from manga_source.downloader import sanitize_filename  # noqa: E402
from scripts.balstorage_client import BalStorageClient, file_id_from_upload, file_size_from_upload, mime_type_from_upload  # noqa: E402


IMAGE_HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
                  "(KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
    "Referer": "https://v2.komikcast.fit/",
    "Accept": "image/avif,image/webp,image/apng,image/*,*/*;q=0.8",
}


def build_parser(description: str) -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=description)
    parser.add_argument("--job-id", required=True)
    parser.add_argument("--api-base", required=True)
    parser.add_argument("--internal-token", required=True)
    parser.add_argument("--balstorage-base", required=True)
    parser.add_argument("--balstorage-email", required=True)
    parser.add_argument("--balstorage-password", required=True)
    parser.add_argument("--balstorage-root", default="Manga")
    parser.add_argument("--max-series", type=int)
    parser.add_argument("--max-chapters", type=int)
    parser.add_argument("--force", action="store_true", help="Re-upload and overwrite existing file metadata")
    parser.add_argument("--missing-only", action="store_true", help="Only fill missing manga/chapter/page data")
    return parser


class MangaIngestor:
    def __init__(self, args: argparse.Namespace):
        self.args = args
        self.api = MangaSourceAPI()
        self.http = requests.Session()
        self.http.headers.update({"X-Internal-Token": args.internal_token})
        self.api_base = args.api_base.rstrip("/")
        self.storage = BalStorageClient(
            args.balstorage_base,
            args.balstorage_email,
            args.balstorage_password,
        )
        self.root_folder: dict[str, Any] | None = None
        self.processed_manga = 0
        self.processed_chapters = 0
        self.processed_pages = 0
        self.failed_items = 0

    def setup(self) -> None:
        self.storage.login()
        self.root_folder = self.storage.ensure_folder(self.args.balstorage_root)

    def post_internal(self, path: str, payload: dict[str, Any] | None = None) -> dict[str, Any]:
        resp = self.request_with_retry(
            lambda: self.http.post(f"{self.api_base}{path}", json=payload or {}, timeout=60),
            "internal callback",
        )
        resp.raise_for_status()
        return resp.json()

    def get_api_data(self, path: str) -> dict[str, Any] | None:
        resp = self.request_with_retry(
            lambda: self.http.get(f"{self.api_base}{path}", timeout=60),
            "manga api lookup",
        )
        if resp.status_code == 404:
            return None
        resp.raise_for_status()
        payload = resp.json()
        data = payload.get("data")
        return data if isinstance(data, dict) else None

    def progress(self, message: str, total_manga: int = 0, total_chapters: int = 0, total_pages: int = 0) -> None:
        self.post_internal(
            f"/internal/ingest/jobs/{self.args.job_id}/progress",
            {
                "total_manga": total_manga,
                "processed_manga": self.processed_manga,
                "total_chapters": total_chapters,
                "processed_chapters": self.processed_chapters,
                "total_pages": total_pages,
                "processed_pages": self.processed_pages,
                "failed_items": self.failed_items,
                "message": message,
            },
        )

    def finish(self, message: str = "ingest finished") -> None:
        self.post_internal(f"/internal/ingest/jobs/{self.args.job_id}/finish", {"message": message})

    def fail(self, error: Exception) -> None:
        try:
            self.post_internal(
                f"/internal/ingest/jobs/{self.args.job_id}/fail",
                {"error_message": str(error)},
            )
        except Exception:
            pass

    def ingest_series(self, series: dict[str, Any], only_chapter: str | None = None) -> None:
        data = series.get("data", {})
        series_id = series.get("id")
        slug = data.get("slug")
        title = data.get("title") or slug
        if not series_id or not slug or not title:
            raise RuntimeError(f"invalid series payload: {series}")

        manga_folder = self.storage.ensure_folder(
            sanitize_filename(title),
            self._folder_id(self.root_folder),
        )

        existing_manga = self.get_api_data(f"/manga/{slug}")
        if self.args.missing_only and existing_manga and not self.args.force:
            self.progress(f"checking missing chapters for {title}")

        details = self._series_details(series_id, fallback=data)
        genres = self._genre_names(details)
        cover = self.upload_cover(
            details.get("coverImage") or data.get("coverImage"),
            manga_folder,
            existing_manga,
        )

        self.post_internal(
            "/internal/ingest/manga",
            {
                "job_id": self.args.job_id,
                "upstream_id": int(series_id),
                "slug": slug,
                "title": details.get("title") or title,
                "native_title": details.get("nativeTitle") or details.get("native_title") or "",
                "author": details.get("author") or "",
                "status": details.get("status") or "",
                "type": details.get("type") or data.get("type") or "",
                "format": details.get("format") or data.get("format") or "",
                "rating": str(details.get("rating") or data.get("rating") or ""),
                "total_chapters": int(details.get("totalChapters") or data.get("totalChapters") or 0),
                "synopsis": details.get("synopsis") or "",
                "cover_file_id": cover.get("file_id", ""),
                "cover_preview_url": cover.get("preview_url", ""),
                "cover_thumbnail_url": cover.get("thumbnail_url", ""),
                "balstorage_folder_id": self._folder_id(manga_folder),
                "genres": genres,
            },
        )

        chapters = list(self.api.all_chapters(int(series_id)))
        if only_chapter is not None:
            chapters = [
                ch for ch in chapters
                if self.chapter_key(ch.get("data", {}).get("index")) == self.chapter_key(only_chapter)
            ]
        if self.args.max_chapters:
            chapters = chapters[: self.args.max_chapters]

        for chapter_meta in chapters:
            self.ingest_chapter(slug, title, manga_folder, chapter_meta)

        self.processed_manga += 1

    def ingest_chapter(
        self,
        manga_slug: str,
        manga_title: str,
        manga_folder: dict[str, Any],
        chapter_meta: dict[str, Any],
    ) -> None:
        ch_data = chapter_meta.get("data", {})
        chapter_index = self.chapter_key(ch_data.get("index"))
        chapter_title = ch_data.get("title") or ""
        chapter_folder = self.storage.ensure_folder(
            sanitize_filename(f"Chapter {chapter_index}"),
            self._folder_id(manga_folder),
        )

        chapter_payload = self.api.get_chapter(manga_slug, chapter_index)
        images = chapter_payload["data"]["data"].get("images", [])
        existing_chapter = self.get_api_data(f"/manga/{manga_slug}/chapters/{chapter_index}")
        existing_pages = self.existing_pages_by_number(existing_chapter)

        if self.args.missing_only and not self.args.force and self.chapter_is_complete(existing_chapter, len(images)):
            self.processed_chapters += 1
            self.progress(f"skipped complete {manga_title} chapter {chapter_index}", total_pages=self.processed_pages)
            return

        self.post_internal(
            "/internal/ingest/chapters",
            {
                "job_id": self.args.job_id,
                "manga_slug": manga_slug,
                "chapters": [
                    {
                        "index": chapter_index,
                        "slug": ch_data.get("slug") or "",
                        "title": chapter_title,
                        "views": int(chapter_meta.get("views", {}).get("total", 0)),
                        "total_pages": len(images),
                        "balstorage_folder_id": self._folder_id(chapter_folder),
                    }
                ],
            },
        )

        page_payload = []
        with tempfile.TemporaryDirectory(prefix="manga_ingest_") as tmp:
            tmp_dir = Path(tmp)
            for page_number, image_url in enumerate(images, 1):
                existing_page = existing_pages.get(page_number)
                if not self.args.force and self.page_is_uploaded(existing_page):
                    page_payload.append(existing_page)
                    self.processed_pages += 1
                    continue

                image_path = self.download_image(image_url, tmp_dir, page_number)
                upload = self.storage.upload_file(self._folder_id(chapter_folder), image_path)
                file_id = file_id_from_upload(upload)
                urls = self.storage.file_urls(file_id)
                page_payload.append(
                    {
                        "page_number": page_number,
                        "source_image_url": image_url,
                        "balstorage_file_id": file_id,
                        "balstorage_folder_id": self._folder_id(chapter_folder),
                        "preview_url": urls["preview_url"],
                        "download_url": urls["download_url"],
                        "thumbnail_url": urls["thumbnail_url"],
                        "mime_type": mime_type_from_upload(upload) or self.mime_for(image_path),
                        "size": file_size_from_upload(upload, image_path),
                    }
                )
                self.processed_pages += 1

        self.post_internal(
            "/internal/ingest/pages",
            {
                "job_id": self.args.job_id,
                "manga_slug": manga_slug,
                "chapter_index": chapter_index,
                "pages": page_payload,
            },
        )
        self.processed_chapters += 1
        self.progress(f"ingested {manga_title} chapter {chapter_index}", total_pages=self.processed_pages)

    def download_image(self, url: str, tmp_dir: Path, page_number: int) -> Path:
        ext = Path(urlparse(url).path).suffix
        if not ext:
            ext = ".jpg"
        target = tmp_dir / f"{page_number:03d}{ext}"
        resp = self.request_with_retry(
            lambda: requests.get(url, headers=IMAGE_HEADERS, timeout=60),
            "image download",
        )
        resp.raise_for_status()
        target.write_bytes(resp.content)
        return target

    def upload_cover(
        self,
        url: str | None,
        manga_folder: dict[str, Any],
        existing_manga: dict[str, Any] | None = None,
    ) -> dict[str, str]:
        if not self.args.force and existing_manga:
            file_id = existing_manga.get("cover_file_id")
            preview_url = existing_manga.get("cover_preview_url")
            thumbnail_url = existing_manga.get("cover_thumbnail_url")
            if file_id and preview_url:
                return {
                    "file_id": file_id,
                    "preview_url": preview_url,
                    "thumbnail_url": thumbnail_url or "",
                }

        if not url:
            return {}

        with tempfile.TemporaryDirectory(prefix="manga_cover_") as tmp:
            tmp_dir = Path(tmp)
            image_path = self.download_named_image(url, tmp_dir, "cover")
            upload = self.storage.upload_file(self._folder_id(manga_folder), image_path)
            file_id = file_id_from_upload(upload)
            urls = self.storage.file_urls(file_id)
            return {
                "file_id": file_id,
                "preview_url": urls["preview_url"],
                "thumbnail_url": urls["thumbnail_url"],
            }

    def download_named_image(self, url: str, tmp_dir: Path, name: str) -> Path:
        ext = Path(urlparse(url).path).suffix
        if not ext:
            ext = ".jpg"
        target = tmp_dir / f"{name}{ext}"
        resp = self.request_with_retry(
            lambda: requests.get(url, headers=IMAGE_HEADERS, timeout=60),
            "image download",
        )
        resp.raise_for_status()
        target.write_bytes(resp.content)
        return target

    @staticmethod
    def request_with_retry(
        fn,
        label: str,
        retries: int = 3,
        retry_statuses: set[int] | None = None,
    ) -> requests.Response:
        retry_statuses = retry_statuses or {408, 429, 500, 502, 503, 504}
        last_error: Exception | None = None
        for attempt in range(1, retries + 1):
            try:
                resp = fn()
                if resp.status_code not in retry_statuses:
                    return resp
                last_error = RuntimeError(f"{label} HTTP {resp.status_code}: {resp.text[:300]}")
            except (requests.Timeout, requests.ConnectionError) as exc:
                last_error = exc

            if attempt < retries:
                sleep_for = min(2 ** attempt, 10)
                print(f"{label} failed, retrying in {sleep_for}s ({attempt}/{retries}): {last_error}")
                time.sleep(sleep_for)

        if last_error:
            raise last_error
        raise RuntimeError(f"{label} failed")

    def _series_details(self, series_id: int, fallback: dict[str, Any]) -> dict[str, Any]:
        try:
            payload = self.api.get_series(series_id)
            return payload.get("data", {}).get("data", fallback)
        except Exception:
            return fallback

    @staticmethod
    def _genre_names(details: dict[str, Any]) -> list[str]:
        names = []
        for genre in details.get("genres", []) or []:
            name = genre.get("data", {}).get("name")
            if name:
                names.append(name)
        return names

    @staticmethod
    def _folder_id(folder: dict[str, Any] | None) -> str:
        if not folder:
            raise RuntimeError("folder is not initialized")
        folder_id = folder.get("id")
        if not folder_id:
            raise RuntimeError(f"folder payload has no id: {folder}")
        return str(folder_id)

    @staticmethod
    def existing_pages_by_number(chapter: dict[str, Any] | None) -> dict[int, dict[str, Any]]:
        if not chapter:
            return {}
        pages = chapter.get("pages")
        if not isinstance(pages, list):
            return {}

        result = {}
        for page in pages:
            try:
                page_number = int(page.get("page_number"))
            except Exception:
                continue
            result[page_number] = {
                "page_number": page_number,
                "source_image_url": page.get("source_image_url") or "",
                "balstorage_file_id": page.get("balstorage_file_id") or "",
                "balstorage_folder_id": page.get("balstorage_folder_id") or "",
                "preview_url": page.get("preview_url") or "",
                "download_url": page.get("download_url") or "",
                "thumbnail_url": page.get("thumbnail_url") or "",
                "mime_type": page.get("mime_type") or "",
                "size": int(page.get("size") or 0),
            }
        return result

    @staticmethod
    def page_is_uploaded(page: dict[str, Any] | None) -> bool:
        if not page:
            return False
        return bool(page.get("balstorage_file_id") and page.get("preview_url"))

    @classmethod
    def chapter_is_complete(cls, chapter: dict[str, Any] | None, expected_pages: int) -> bool:
        if not chapter or expected_pages < 1:
            return False
        pages = cls.existing_pages_by_number(chapter)
        uploaded = sum(1 for page in pages.values() if cls.page_is_uploaded(page))
        return uploaded >= expected_pages

    @staticmethod
    def mime_for(path: Path) -> str:
        ext = path.suffix.lower()
        if ext in {".jpg", ".jpeg"}:
            return "image/jpeg"
        if ext == ".png":
            return "image/png"
        if ext == ".webp":
            return "image/webp"
        return "application/octet-stream"

    @staticmethod
    def chapter_key(value: Any) -> str:
        text = str(value).strip()
        if "." in text:
            text = text.rstrip("0").rstrip(".")
        return text
