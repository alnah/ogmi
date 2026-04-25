# Ogmi

[![CI](https://github.com/alnah/ogmi/actions/workflows/ci.yml/badge.svg)](https://github.com/alnah/ogmi/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/alnah/ogmi)](https://github.com/alnah/ogmi/releases)
[![Codecov](https://codecov.io/gh/alnah/ogmi/graph/badge.svg)](https://codecov.io/gh/alnah/ogmi)
[![Go Reference](https://pkg.go.dev/badge/github.com/alnah/ogmi.svg)](https://pkg.go.dev/github.com/alnah/ogmi)

> Ogmi is a command-line interface (CLI) that queries structured language-learning descriptors and returns JSON data.

Teachers, schools, publishers, or any curriculum designers can use it to inspect descriptor data.
Coding agents can use its JSON output in lesson-planning and curriculum workflows.

## Status

Ogmi is a v0 software. The CLI, JSON structures may still change before a stable release.
Source code is licensed under Apache-2.0, but descriptor specs may have separate rights.
Some bundled specs are under rights review. See [NOTICE](NOTICE) and [specs/README.md](specs/README.md).

## What it does

- JSON output by default for agents and automation.
- Human-readable output with `--format text`.
- Embedded specs by default.
- External specs override with `--specs PATH` or `OGMI_SPECS`.
- Query descriptors by corpus, level, scale, path fields, ID, and text.
- Inspect schema, scales, examples, and coverage.

## Install

### Go install

```sh
go install github.com/alnah/ogmi/cmd/ogmi@latest
```

## Quick start

List available descriptor corpora:

```sh
ogmi descriptors corpora
```

List descriptors as JSON:

```sh
ogmi descriptors list --corpus themes --level a1 --limit 5
```

Print text output:

```sh
ogmi descriptors corpora --format text
```

Inspect fields and schema:

```sh
ogmi descriptors fields --corpus themes
ogmi descriptors schema --field level --corpus themes
```

Get one descriptor by ID:

```sh
ogmi descriptors get \
  --corpus themes \
  --id themes.descriptors.food.meals_food_preferences_and_basic_needs.a1
```

Build a coverage matrix:

```sh
ogmi descriptors coverage --corpus themes
```

Export bundled specs:

```sh
ogmi specs export --output ./specs
```

Use external specs:

```sh
ogmi --specs ./specs descriptors list --corpus themes --limit 3
```

## Output

Data commands return JSON by default.
JSON responses include a `kind` and `schemaVersion` field so agents can identify payload shape.

Example:

```json
{
  "kind": "descriptor_corpora",
  "schemaVersion": "v1",
  "corpora": [
    {
      "name": "themes",
      "pathFields": [],
      "defaultCoverageAxes": ["corpus", "level"],
      "files": ["specs/themes/descriptors.yml"]
    }
  ]
}
```

Use `--format text` with commands that support text output.

## Specs source precedence

Ogmi resolves specs in this order:

1. `--specs PATH`
2. `OGMI_SPECS`
3. embedded specs

External specs replace embedded specs.
They are useful for private, licensed, experimental, or institution-specific descriptor data.

## Commands

| Command | Purpose |
| --- | --- |
| `ogmi descriptors corpora` | List descriptor corpora. |
| `ogmi descriptors fields` | List descriptor fields. |
| `ogmi descriptors schema` | Describe descriptor schema and field values. |
| `ogmi descriptors list` | Query descriptors with filters and pagination. |
| `ogmi descriptors scales` | Query descriptor scales. |
| `ogmi descriptors get` | Get one descriptor by ID. |
| `ogmi descriptors compare-levels` | Compare level coverage for descriptors. |
| `ogmi descriptors coverage` | Build a coverage matrix. |
| `ogmi descriptors examples` | Show example descriptor queries. |
| `ogmi specs export` | Export embedded specs to a directory. |
| `ogmi version` | Print the Ogmi version. |

Run help for details:

```sh
ogmi --help
ogmi descriptors list --help
```

## Descriptor corpora

Current bundled corpora are:

- `cefr`
- `french`
- `texts`
- `themes`

See [specs/README.md](specs/README.md) before reusing bundled descriptor data.
Some specs are under rights review and may not be covered by the software license.

## Development

Use the Makefile for local checks:

```sh
make help
make quick
make check
```

Common targets:

```sh
make fmt              # format Go files and imports
make lint             # run golangci-lint
make test             # run tests
make test-race        # run race tests
make cover            # write coverage.out and print coverage summary
make govulncheck      # check reachable Go vulnerabilities
make goreleaser-check # validate GoReleaser config
make snapshot         # build local snapshot release artifacts
make clean            # remove dist/ and coverage.out
```

Before pushing changes:

```sh
make check
```

## License

Ogmi software source code is licensed under the Apache License, Version 2.0.
See [LICENSE](LICENSE).

Descriptor specs are data/content and may have separate rights.
Third-party material remains subject to its original rights and is not relicensed by Ogmi.
See [NOTICE](NOTICE) and [specs/README.md](specs/README.md).

Services around Ogmi, such as hosting, integration, training, and support, may use separate terms.
