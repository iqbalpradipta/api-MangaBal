"""
Command-line interface for the manga source scraper.
"""

import argparse
import json
import os
import sys
import time
from pathlib import Path

from .api import MangaSourceAPI
from .downloader import ChapterDownloader, sanitize_filename


def cmd_list(args):
    api = MangaSourceAPI()
    result = api.list_series(args.page)
    print(f"Page {result['meta']['page']} / {result['meta']['lastPage']}")
    print(f"Total series: {result['meta']['total']}")
    print("-" * 60)
    for s in result["data"]:
        d = s["data"]
        print(f"  [{s['id']}] {d['title']}")
        print(f"       Slug: {d['slug']}  |  Type: {d['type']}  |  "
              f"Format: {d['format']}  |  Rating: {d['rating']}  |  "
              f"Ch: {d['totalChapters']}")


def cmd_detail(args):
    api = MangaSourceAPI()

    if args.slug:
        result = api.search_series(args.slug)
        if result["data"]:
            series_id = result["data"][0]["id"]
        else:
            print(f"No series found with slug '{args.slug}'")
            return
    else:
        series_id = args.id

    series = api.get_series(series_id)
    d = series["data"]["data"]

    print(f"Title       : {d['title']}")
    print(f"Native      : {d.get('nativeTitle', '-')}")
    print(f"Author      : {d.get('author', '-')}")
    print(f"Status      : {d.get('status', '-')}")
    print(f"Format      : {d.get('format', '-')}")
    print(f"Rating      : {d.get('rating', '-')}")
    print(f"Chapters    : {d.get('totalChapters', '-')}")
    print(f"Synopsis    : {d.get('synopsis', '-')}")
    genres = d.get("genres", [])
    if genres:
        print(f"Genres      : {', '.join(g['data']['name'] for g in genres)}")
    else:
        genre_ids = d.get("genreIds", [])
        if genre_ids:
            print(f"Genre IDs   : {genre_ids}")


def cmd_chapters(args):
    api = MangaSourceAPI()
    result = api.list_chapters(args.id, args.page)

    if not result.get("data"):
        print("No chapters found. Try listing with --page 1")
        return

    print(f"Series ID: {args.id}")
    print(f"Page: {result.get('meta', {}).get('page', '?')}")
    print("-" * 50)
    for ch in result["data"]:
        d = ch["data"]
        title = d.get("title") or "-"
        print(f"  Ch {d['index']:4d}  |  {title}  |  "
              f"Views: {ch.get('views', {}).get('total', 0)}")


def cmd_download(args):
    api = MangaSourceAPI()
    downloader = ChapterDownloader(concurrency=args.concurrency)

    # Resolve series info
    if args.slug:
        result = api.search_series(args.slug)
        if not result["data"]:
            print(f"Series '{args.slug}' not found.")
            return
        series = result["data"][0]
    else:
        result = api.get_series(args.id)
        series = result if "data" in result else result

    d = series["data"] if "data" in series else series["data"]
    series_id = series["id"] if "id" in series else series.get("id", args.id)
    series_slug = d["slug"]
    series_title = d["title"]
    total = int(d.get("totalChapters", 0))

    print(f"Series : {series_title}")
    print(f"Slug   : {series_slug}")
    print(f"Chapters to download: {total}")
    print()

    if args.chapter:
        # Download a single chapter
        print(f"Fetching chapter {args.chapter}...")
        ch_data = api.get_chapter(series_slug, args.chapter)
        images = ch_data["data"]["data"]["images"]
        print(f"Found {len(images)} images")
        downloader.download_sync(
            images, args.output, series_title, args.chapter,
        )
    else:
        # Download all chapters
        for ch_meta in api.all_chapters(series_id):
            idx = ch_meta["data"]["index"]
            print(f"Chapter {idx}...")
            ch_data = api.get_chapter(series_slug, idx)
            images = ch_data["data"]["data"]["images"]
            print(f"  {len(images)} pages")
            downloader.download_sync(
                images, args.output, series_title, idx,
            )

    print("\nDone!")


def cmd_download_all(args):
    """
    Download every chapter of every manga series from the configured source.

    Resumable: tracks progress in a JSON file so interrupted downloads
    can continue from where they left off.
    """
    api = MangaSourceAPI()
    downloader = ChapterDownloader(concurrency=args.concurrency)
    output = Path(args.output)
    output.mkdir(parents=True, exist_ok=True)

    # ── Load / create progress file ──────────────────────────────
    progress_file = output / ".download_progress.json"
    if progress_file.exists() and not args.reset:
        with open(progress_file, "r", encoding="utf-8") as f:
            progress = json.load(f)
        print(f"Resuming from progress file ({len(progress)} series already done)")
    else:
        progress = {}

    def save_progress():
        with open(progress_file, "w", encoding="utf-8") as f:
            json.dump(progress, f, ensure_ascii=False)

    # ── Stats ────────────────────────────────────────────────────
    stats = {"downloaded_chapters": 0, "downloaded_images": 0, "skipped_series": 0, "errors": 0}

    # ── Iterate all series ───────────────────────────────────────
    total_series = None
    page = 1
    series_index = 0
    start_time = time.time()

    while True:
        result = api.list_series(page)
        if total_series is None:
            total_series = result["meta"]["total"]
            print(f"Total series from source: {total_series}")
            print(f"Output directory: {output.resolve()}")
            print(f"Concurrency: {args.concurrency}")
            print()

        for s in result["data"]:
            series_index += 1
            sid = str(s["id"])
            d = s["data"]
            slug = d["slug"]
            title = d["title"]
            total_ch = int(d.get("totalChapters", 0))

            # Skip if already fully downloaded
            if sid in progress and progress[sid].get("status") == "done":
                stats["skipped_series"] += 1
                continue

            # Skip if --max-chapters exceeded for resumed series
            done_chapters = progress.get(sid, {}).get("chapters", [])
            if total_ch == 0 or (args.max_chapters and len(done_chapters) >= args.max_chapters):
                progress[sid] = {"slug": slug, "title": title, "chapters": done_chapters, "status": "done"}
                continue

            elapsed = time.time() - start_time
            eta = ""
            if series_index > 0:
                rate = series_index / elapsed
                remaining = (total_series - series_index) / rate
                eta = f" | ETA: {remaining/3600:.1f}h"

            print(f"\n{'='*60}")
            print(f"[{series_index}/{total_series}] {title}")
            print(f"    Slug: {slug}  |  Chapters: {total_ch}  |  Downloaded: {len(done_chapters)}{eta}")
            print(f"{'='*60}")

            try:
                for ch_meta in api.all_chapters(s["id"]):
                    ch_idx = ch_meta["data"]["index"]

                    if ch_idx in done_chapters:
                        continue
                    if args.max_chapters and len(done_chapters) >= args.max_chapters:
                        break

                    try:
                        ch_data = api.get_chapter(slug, ch_idx)
                        images = ch_data["data"]["data"]["images"]
                        downloader.download_sync(images, output, title, ch_idx)
                        done_chapters.append(ch_idx)
                        stats["downloaded_chapters"] += 1
                        stats["downloaded_images"] += len(images)
                        save_progress()
                    except Exception as e:
                        print(f"  ERROR ch {ch_idx}: {e}")
                        stats["errors"] += 1
                        if args.stop_on_error:
                            raise

                progress[sid] = {"slug": slug, "title": title, "chapters": done_chapters, "status": "done"}
                save_progress()

            except KeyboardInterrupt:
                print(f"\n\nInterrupted. Progress saved. Run again to resume.")
                print_summary(stats, start_time)
                return
            except Exception as e:
                print(f"  SERIES ERROR: {e}")
                progress[sid] = {"slug": slug, "title": title, "chapters": done_chapters, "status": "error"}
                save_progress()
                stats["errors"] += 1
                if args.stop_on_error:
                    raise

        meta = result["meta"]
        if page >= meta["lastPage"]:
            break
        page += 1

    print(f"\n{'='*60}")
    print(f"ALL DONE!")
    print_summary(stats, start_time)


def print_summary(stats, start_time):
    elapsed = time.time() - start_time
    print(f"  Series processed: {stats.get('skipped_series', 0)} skipped")
    print(f"  Chapters downloaded: {stats['downloaded_chapters']}")
    print(f"  Images downloaded: {stats['downloaded_images']}")
    print(f"  Errors: {stats['errors']}")
    print(f"  Time elapsed: {elapsed/3600:.1f} hours ({elapsed/60:.0f} minutes)")


def cmd_genres(args):
    api = MangaSourceAPI()
    result = api.list_genres()
    for g in result["data"]:
        d = g["data"]
        print(f"  [{g['id']:2d}] {d['name']}")


def cmd_search(args):
    api = MangaSourceAPI()
    result = api.search_series(args.query)
    if not result["data"]:
        print(f"No results for '{args.query}'")
        return
    for s in result["data"]:
        d = s["data"]
        print(f"  [{s['id']}] {d['title']}  (slug: {d['slug']})")


def main():
    parser = argparse.ArgumentParser(
        description="Manga Source Scraper",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  py main.py list --page 1
  py main.py detail --id 10069
  py main.py detail --slug mumumu
  py main.py search "Solo Leveling"
  py main.py chapters --id 10069
  py main.py download --slug mumumu --chapter 1
  py main.py download --slug mumumu --all --output ./manga
  py main.py genres
        """,
    )

    sub = parser.add_subparsers(dest="command")

    # list
    p = sub.add_parser("list", help="List manga series")
    p.add_argument("--page", type=int, default=1)
    p.set_defaults(func=cmd_list)

    # detail
    p = sub.add_parser("detail", help="Show series details")
    g = p.add_mutually_exclusive_group(required=True)
    g.add_argument("--id", type=int)
    g.add_argument("--slug")
    p.set_defaults(func=cmd_detail)

    # search
    p = sub.add_parser("search", help="Search series by slug/name")
    p.add_argument("query")
    p.set_defaults(func=cmd_search)

    # chapters
    p = sub.add_parser("chapters", help="List chapters of a series")
    p.add_argument("--id", type=int, required=True)
    p.add_argument("--page", type=int, default=1)
    p.set_defaults(func=cmd_chapters)

    # download
    p = sub.add_parser("download", help="Download manga chapters")
    g = p.add_mutually_exclusive_group(required=True)
    g.add_argument("--id", type=int)
    g.add_argument("--slug")
    p.add_argument("--chapter", type=int, help="Download specific chapter only")
    p.add_argument("--all", action="store_true", help="Download all chapters")
    p.add_argument("--output", "-o", default="./downloads", help="Output directory")
    p.add_argument("--concurrency", "-c", type=int, default=10,
                   help="Parallel downloads (default: 10)")
    p.set_defaults(func=cmd_download)

    # download-all
    p = sub.add_parser("download-all", help="Download ALL manga series (batch)")
    p.add_argument("--output", "-o", default="./downloads", help="Output directory")
    p.add_argument("--concurrency", "-c", type=int, default=10,
                   help="Parallel image downloads (default: 10)")
    p.add_argument("--max-chapters", type=int, default=None,
                   help="Max chapters per series (default: unlimited)")
    p.add_argument("--reset", action="store_true",
                   help="Ignore saved progress and start fresh")
    p.add_argument("--stop-on-error", action="store_true",
                   help="Stop immediately on any error")
    p.set_defaults(func=cmd_download_all)

    # genres
    p = sub.add_parser("genres", help="List all genres")
    p.set_defaults(func=cmd_genres)

    args = parser.parse_args()
    if args.command is None:
        parser.print_help()
        return

    args.func(args)


if __name__ == "__main__":
    main()
