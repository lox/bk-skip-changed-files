# Buildkite Skip Unchanged Files

A common need in a Buildkite pipeline is to skip steps in a pipeline.yml if specific files in the changeset haven't changed.

This implements a library and a command line tool to help with this by preprocessing the pipeline.yml file and adding support for a `skip_if_unchanged` property that takes a list of zglobs.

## Usage

```
bk-skip-unchanged-files pipeline.yml --upload # uploads the modified pipeline.yml to Buildkite
```

## How diffs are detected

To detect changes, we use Git's `merge-base` command to find the common ancestor between the current HEAD and the base branch of the pull request. Then, we compare the files in the current state with this common ancestor.

Here's the basic process:

1. Find the common ancestor:
   ```bash
   git merge-base $BUILDKITE_PULL_REQUEST_BASE_BRANCH HEAD
   ```

2. Use this commit SHA to compare with the current state:
   ```bash
   git diff --name-only <common-ancestor-sha> HEAD
   ```

This approach will give us a list of files that have changed since the branch point, which we can then use to determine if specific steps should be skipped based on the `skip_if_unchanged` property.

### Handling different types of changes

The `git diff --name-only` command handles various types of file changes:

- **Changed files**: These will be listed in the output.
- **Removed files**: These will also be listed in the output.
- **Renamed files**: By default, these are treated as a file deletion and a new file addition. Both the old and new filenames will be listed.

To improve handling of renamed files, we can add the `--find-renames` option:

```bash
git diff --name-only --find-renames <common-ancestor-sha> HEAD
```

This will detect renamed files and only list the new filename, treating it as a modification rather than a separate deletion and addition.

For even more detailed information about the type of change for each file, we can use:

```bash
git diff --name-status --find-renames <common-ancestor-sha> HEAD
```

This will prefix each filename with a status code:
- `M` for modified files
- `A` for added files
- `D` for deleted files
- `R` for renamed files

Depending on your specific needs, you may choose to process this more detailed output to handle different types of changes in custom ways.
