# Security Policy

## Supported versions

Ogmi is pre-1.0 software. Users must upgrade to the latest release when a security fix is published.

## Reporting a vulnerability

Do not report security vulnerabilities through public GitHub issues, discussions, or pull requests.

Use GitHub private vulnerability reporting for this repository.

Include:

- affected version, tag, commit, or release asset;
- operating system and architecture;
- steps to reproduce;
- expected and actual behavior;
- impact;
- proof of concept, if available;
- whether the issue affects default embedded specs, external specs, release
  artifacts, or the build/release workflow.

I aim to acknowledge reports within 7 days.

## Scope

Security issues include, but are not limited to:

- arbitrary file overwrite or path traversal;
- unsafe handling of external specs;
- unexpected code execution;
- release artifact or checksum compromise;
- dependency vulnerabilities that are reachable through Ogmi;
- CI or release workflow issues that could compromise published artifacts.

## Out of scope

The following are not usually security vulnerabilities:

- incorrect, incomplete, or disputed descriptor content;
- rights, licensing, or attribution issues in descriptor data;
- model or agent output quality when another tool consumes Ogmi JSON;
- malicious external specs supplied by a user, unless Ogmi handles them in an
  unsafe way beyond parsing and reporting errors;
- vulnerabilities only present in unsupported Go versions or unsupported
  operating systems.

## Disclosure

If a vulnerability is confirmed, I will coordinate disclosure with the reporter,
prepare a fix, publish a release, and document the issue in a GitHub Security
Advisory when appropriate.
