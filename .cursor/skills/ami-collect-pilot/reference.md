# Collect pilot reference

## Config fields (`collect_source.config_json`)

| Field | Purpose |
|-------|---------|
| `list_url` | First list page (tag/category home) |
| `article_base` | Base URL for resolving relative links |
| `link_extract.must_contain` | Substring required in article URLs |
| `link_extract.exclude_contains` | Substrings that disqualify a link |
| `link_extract.prefer_path_regex` | Optional Go regex on article path |
| `pagination.type` | `none`, `path`, `wordpress_ajax`, `scroll` (scroll unsupported) |
| `pagination.path_template` | For `path`: `{page}` placeholder |
| `pagination.start_page` | For `path`: first generated page number (often 2) |
| `pagination.max_pages` | Extra pages after first (not last page index) |
| `pagination.stop_on_empty_list` | Stop when a list page yields zero links |
| `max_items_per_run` | Cap per `run` |
| `delay_ms` | Pause between HTTP calls |
| `timeout_sec` | HTTP timeout |

## Detecting WordPress AJAX pagination

On the list HTML, confirm:

- `var blog_loadmore_params = {...}`
- `#loadMoreblog` with `data-load-posts`
- `action: loadmore` in theme JS

Then set `pagination.type` to `wordpress_ajax`. `max_pages` can mirror `max_page` from JSON (often 5 on Unit42 tag pages).

## Quality gate (article pilot)

Checks include HTML size, pipeline `succeeded`, raw on disk, images, uploads, remote `img` src, knowledge (informational).

Pilot `passed` requires `quality.passed` true on the article test response.

## API routes

- `GET /api/v1/settings/collect-sources`
- `POST /api/v1/settings/collect-sources`
- `PATCH /api/v1/settings/collect-sources/{id}`
- `POST /api/v1/settings/collect-sources/{id}/test/list`
- `POST /api/v1/settings/collect-sources/{id}/test/article`
- `POST /api/v1/settings/collect-sources/{id}/unlock-batch`
- `POST /api/v1/settings/collect-sources/{id}/run`
