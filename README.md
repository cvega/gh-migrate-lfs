# gh-migrate-lfs

`gh-migrate-lfs` is a [GitHub CLI](https://cli.github.com) extension to assist in the migration of Git LFS files between GitHub organizations. While [GitHub Enterprise Importer](https://github.com/github/gh-gei) handles many aspects of organization migration, there can be challenges with large Git LFS files. This extension helps ensure all your LFS content is properly migrated. Whether you're consolidating organizations, setting up new environments, or need to replicate repositories with LFS content, this extension can help.

## Install

```bash
gh extension install mona-actions/gh-migrate-lfs
```

## Usage: Export

Export a list of repositories containing Git LFS files to a CSV file.

```bash
Usage:
  migrate-lfs export [flags]

Flags:
  -f, --file-prefix string    Output filenames prefix 
  -h, --help                  help for export
  -n, --hostname string       GitHub Enterprise Server hostname (optional) Ex. github.example.com
  -o, --organization string   Organization to export (required)
  -t, --token string          GitHub token (required)
  -s, --search-depth string   Search depth for .gitattributes file (default "1")
```

### Example Export Command

```bash
gh migrate-lfs export \
  --organization my-org \
  --token ghp_xxxxxxxxxxxx \
  --file-prefix my-lfs \
  --depth 2
```

This will create a file named `my-lfs_lfs.csv` containing all repositories with LFS files. The export process provides detailed feedback:

```
📊 Export Summary:
Total repositories found: 25
✅ Successfully processed: 23 repositories
❌ Failed to process: 2 repositories
🔍 Repositories with LFS: 8
📁 Output file: my-lfs_lfs.csv
🔍 Maximum search depth: 2
```

## Usage: Pull

Clone repositories and download their LFS content. 

```bash
Usage:
  migrate-lfs pull [flags]

Flags:
  -f, --file string          Exported LFS repos file path, csv format (required)
  -d, --work-dir string      Working directory for cloned repositories (required)
  -h, --help                 help for pull
  -n, --hostname string      GitHub Enterprise Server hostname URL (optional)
  -t, --token string         GitHub token with repo scope (required)
  -w, --workers string       Concurrency git workers (default "1")
```

### Example Pull Command

```bash
gh migrate-lfs pull \
  --file my-lfs_lfs.csv \
  --work-dir ./repos \
  --source-org source-org \
  --token ghp_xxxxxxxxxxxx \
  --workers 4
```

The pull process provides detailed feedback:

```
📊 Pull Summary:
✅ Successfully processed: 8 repositories
❌ Failed: 0 repositories
```

## Usage: Sync

Push LFS content to repositories in the target organization.

```bash
Usage:
  migrate-lfs sync [flags]

Flags:
  -f, --file string         Exported LFS repos file path, csv format (required)
  -h, --help                help for sync
  -o, --target-org string   Target organization name (required)
  -t, --token string        GitHub token with repo scope (required)
  -w, --work-dir string     Working directory with cloned repositories (required)
```

### Example Sync Command

```bash
gh migrate-lfs sync \
  --input my-lfs_lfs.csv \
  --new-org target-org \
  --token ghp_xxxxxxxxxxxx \
  --work-dir ./repos
```

The sync process provides detailed feedback:

```
📊 Sync Summary:
✅ Successfully processed: 8 repositories
❌ Failed: 0 repositories
```

### CSV Format

The tool exports and imports repository information using the following CSV format:

```csv
Repository,GitAttributesPath
example-repo,.gitattributes
another-repo,src/.gitattributes
```

- `Repository`: The name of the repository
- `GitAttributesPath`: Path to .gitattributes file containing LFS configurations

## Required Permissions

### For Export, Pull and Sync

- repository contents: `repo`
- clone: `repo`
- git lfs pull: `repo`
- git lfs push: `repo`

## Proxy Support

The tool supports proxy configuration through both command-line flags and environment variables:

### Command-line flags:
```bash
Global Flags:
      --http-proxy string    HTTP proxy (can also use HTTP_PROXY env var)
      --https-proxy string   HTTPS proxy (can also use HTTPS_PROXY env var)
      --no-proxy string      No proxy list (can also use NO_PROXY env var)
```

```bash
# Example usage with proxy:
gh migrate-lfs export \
  --organization my-org \
  --token ghp_xxxxxxxxxxxx \
  --file-prefix my-lfs \
  --https-proxy https://proxy.example.com:8080
```

### Environment variables:
- `HTTP_PROXY`: Proxy for HTTP requests
- `HTTPS_PROXY`: Proxy for HTTPS requests
- `NO_PROXY`: Comma-separated list of hosts to exclude from proxy

Example with environment variables:
```bash
export HTTPS_PROXY=https://proxy.example.com:8080
export NO_PROXY=github.internal.com
```
```bash
gh migrate-lfs export \
  --organization my-org \
  --token ghp_xxxxxxxxxxxx \
  --file-prefix my-lfs
```

## Limitations

- Target repositories must exist in the destination organization before syncing
- Large LFS files may take significant time to download and upload
- Network bandwidth and storage space should be considered when migrating large LFS repositories
- The tool will retry failed operations but may still encounter persistent access or network issues
- Deep directory structures may require adjusting the search depth parameter

## License

- [MIT](./license) (c) [Mona-Actions](https://github.com/mona-actions)
- [Contributing](./contributing.md)