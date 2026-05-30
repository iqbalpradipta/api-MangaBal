from ingest_common import MangaIngestor, build_parser


def main() -> int:
    parser = build_parser("Ingest all manga series into Manga API and BalStorage")
    args = parser.parse_args()

    ingestor = MangaIngestor(args)
    try:
        ingestor.setup()
        series_iter = ingestor.api.all_series()
        total = None
        for idx, series in enumerate(series_iter, 1):
            if args.max_series and idx > args.max_series:
                break
            if total is None:
                total = 0
            title = series.get("data", {}).get("title", "unknown")
            ingestor.progress(f"ingesting {title}", total_manga=total)
            try:
                ingestor.ingest_series(series)
            except Exception:
                ingestor.failed_items += 1
                if not args.max_series:
                    raise
        ingestor.finish("all manga ingest finished")
        return 0
    except Exception as exc:
        ingestor.fail(exc)
        raise


if __name__ == "__main__":
    raise SystemExit(main())
