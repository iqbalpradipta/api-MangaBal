from ingest_common import MangaIngestor, build_parser


def main() -> int:
    parser = build_parser("Ingest one manga series or chapter into Manga API and BalStorage")
    parser.add_argument("--slug", required=True)
    parser.add_argument("--chapter")
    args = parser.parse_args()

    ingestor = MangaIngestor(args)
    try:
        ingestor.setup()
        result = ingestor.api.search_series(args.slug)
        if not result.get("data"):
            raise RuntimeError(f"series not found: {args.slug}")
        ingestor.ingest_series(result["data"][0], only_chapter=args.chapter)
        ingestor.finish("series ingest finished")
        return 0
    except Exception as exc:
        ingestor.fail(exc)
        raise


if __name__ == "__main__":
    raise SystemExit(main())
