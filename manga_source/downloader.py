"""
Async chapter image downloader.
"""

import asyncio
import os
import re
from pathlib import Path

import aiofiles
import aiohttp
from tqdm.asyncio import tqdm_asyncio


def sanitize_filename(name: str) -> str:
    """Remove characters that are illegal in file/directory names."""
    return re.sub(r'[<>:"/\\|?*]', "_", name).strip()


class ChapterDownloader:
    """Downloads chapter images concurrently."""

    def __init__(self, concurrency: int = 10, timeout: int = 60):
        self.concurrency = concurrency
        self.timeout = timeout
        self.headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
            "AppleWebKit/537.36 (KHTML, like Gecko) "
            "Chrome/125.0.0.0 Safari/537.36",
            "Referer": "https://v2.komikcast.fit/",
            "Accept": "image/avif,image/webp,image/apng,image/*,*/*;q=0.8",
        }

    async def _download_one(
        self,
        session: aiohttp.ClientSession,
        url: str,
        dest: Path,
        sem: asyncio.Semaphore,
    ) -> bool:
        async with sem:
            try:
                async with session.get(url, timeout=self.timeout) as resp:
                    if resp.status == 200:
                        async with aiofiles.open(dest, "wb") as f:
                            await f.write(await resp.read())
                        return True
                    return False
            except Exception:
                return False

    async def download_chapter(
        self,
        image_urls: list[str],
        output_dir: str | Path,
        series_title: str,
        chapter_index: int,
        chapter_title: str | None = None,
    ) -> Path:
        """
        Download all images for a chapter.

        Parameters
        ----------
        image_urls : list[str]
            List of image URLs from the API.
        output_dir : str or Path
            Root directory for downloads.
        series_title : str
            Used to create the sub-directory: {output_dir}/{series_title}
        chapter_index : int
            Chapter number.
        chapter_title : str, optional
            Chapter title appended to folder name.

        Returns
        -------
        Path
            The directory where images were saved.
        """
        series_dir = sanitize_filename(series_title)
        folder = f"Chapter {chapter_index}"
        if chapter_title:
            folder += f" - {sanitize_filename(chapter_title)}"

        dest_dir = Path(output_dir) / series_dir / sanitize_filename(folder)
        dest_dir.mkdir(parents=True, exist_ok=True)

        sem = asyncio.Semaphore(self.concurrency)
        connector = aiohttp.TCPConnector(
            limit=self.concurrency, limit_per_host=self.concurrency
        )
        timeout = aiohttp.ClientTimeout(total=self.timeout)

        async with aiohttp.ClientSession(
            headers=self.headers, connector=connector, timeout=timeout
        ) as session:
            tasks = []
            for i, url in enumerate(image_urls, 1):
                ext = url.rsplit(".", 1)[-1].split("?")[0]
                dest = dest_dir / f"{i:03d}.{ext}"
                if dest.exists():
                    continue
                tasks.append(self._download_one(session, url, dest, sem))

            if tasks:
                results = await tqdm_asyncio.gather(
                    *tasks, desc=f"  Ch {chapter_index}"
                )
                ok = sum(1 for r in results if r)
                print(f"  Downloaded {ok}/{len(tasks)} images")
            else:
                print(f"  All images already downloaded")

        return dest_dir

    def download_sync(
        self,
        image_urls: list[str],
        output_dir: str | Path,
        series_title: str,
        chapter_index: int,
        chapter_title: str | None = None,
    ) -> Path:
        """Synchronous wrapper around download_chapter."""
        return asyncio.run(
            self.download_chapter(
                image_urls, output_dir, series_title, chapter_index, chapter_title
            )
        )
