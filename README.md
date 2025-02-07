# deduplicate

A CLI tool that recursively scans all files in a given directory, computes SHA256 hashes, and stores results in an SQLite database. Duplicate files are moved to a `duplicate_to_be_deleted` folder.

## Features
- **Recursive scanning** of all files in a directory
- **SHA256 hashing** for each file (parallelized using Go routines)
- **SQLite storage** for file metadata
- **Duplicate detection** based on hash and file size
- **Automatic moving** of duplicates to a separate folder

## Installation

You can download the latest prebuilt binaries from the [Releases](https://github.com/aschbacd/deduplicate/releases) page.

Alternatively, install using Go:

```bash
go install github.com/aschbacd/deduplicate@latest
```

## Usage

```bash
deduplicate -path /your/folder -db results.db
```

```text
$ deduplicate --help

Usage: deduplicate [OPTIONS]

Options:
  -db string
        Path to SQLite database (default "deduplicate.db")
  -path string
        Directory to scan (default ".")

Example:
  deduplicate -path /your/folder -db results.db
```

### CLI Flags

|Flag|Description|Default|
|-|-|-|
|`-path`|Directory to scan|`.`|
|`-db`|SQLite database file|`deduplicate.db`|

## Releasing a New Version

To create a new release, push a tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will automatically build and publish a new release.

## License

This project is licensed under the MIT License.
