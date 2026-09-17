# devlogo

CLI for building a static engineering log from Markdown, with entry pages, an index, and tag pages.

## Installation

Requires Go 1.27.1 or later.

```sh
go install github.com/korotkovfedor/devlogo/cmd/devlogo@latest
```

The installation directory (usually `~/go/bin`, or `GOBIN` if set) must be in your `PATH`.

## Quick start

A configuration file and HTML templates are required. A ready-to-use example is included in `example/`:

```sh
git clone https://github.com/korotkovfedor/devlogo.git
cd devlogo/example
devlogo serve
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080). Press `Ctrl+C` to stop the server.

| Command | Description |
| --- | --- |
| `devlogo build` | Build the site in `output_dir`. |
| `devlogo serve` | Build the site and start a local server. |

Restart `serve` after editing entries or templates: automatic rebuilds are not supported.

## Configuration

`config.yaml` is read from the current working directory. Directory paths are relative to it, and template filenames are relative to `template_dir`.

```yaml
title: "My App"
content_dir: "./entries"
template_dir: "./templates"
page_template: "page.html"
index_template: "index.html"
tags_template: "tags.html"
output_dir: "./dist"
server_port: 8080
```

Each build deletes and recreates `output_dir`. You can deploy the generated directory to any static hosting service.

## Entries

Add `.md` files directly to `content_dir`. Start each file with YAML front matter:

```markdown
---
title: "Added a local server"
date: 2026-09-13
type: feature
status: done
tags: [go, cli]
---

## Change

The log can now be previewed locally with `devlogo serve`.
```

`title`, `date`, `type`, and `status` are required. `tags` is optional.

## License

[MIT](LICENSE).
