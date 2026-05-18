#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: detect-docker-image-targets.sh [--github-output] <before-ref> <sha>
EOF
}

write_github_output=false

if [[ "${1:-}" == "--github-output" ]]; then
  write_github_output=true
  shift
fi

if [[ $# -ne 2 ]]; then
  usage
  exit 64
fi

before_ref="$1"
sha="$2"
repo_root="$(git rev-parse --show-toplevel)"

changed_files=()
if [[ "$before_ref" =~ ^0+$ ]]; then
  while IFS= read -r file; do
    changed_files+=("$file")
  done < <(git diff-tree --no-commit-id --name-only -r "$sha")
else
  while IFS= read -r file; do
    changed_files+=("$file")
  done < <(git diff --name-only "$before_ref" "$sha")
fi

bot=false
worker=false

local_dependency_dirs() {
  local package="$1"

  go list -deps -f '{{if and .Module (eq .Module.Path "github.com/fogo-sh/borik")}}{{.Dir}}{{end}}' "$package" |
    while IFS= read -r dir; do
      [[ -n "$dir" ]] || continue
      dir="$(cd "$dir" && pwd -P)"
      printf '%s\n' "${dir#"$repo_root"/}"
    done |
    sort -u
}

bot_dirs=()
while IFS= read -r dir; do
  bot_dirs+=("$dir")
done < <(local_dependency_dirs ./cmd/borik)

worker_dirs=()
while IFS= read -r dir; do
  worker_dirs+=("$dir")
done < <(local_dependency_dirs ./cmd/borik-worker)

file_is_in_dirs() {
  local file="$1"
  shift

  local dir
  for dir in "$@"; do
    if [[ "$file" == "$dir" || "$file" == "$dir"/* ]]; then
      return 0
    fi
  done

  return 1
}

if ((${#changed_files[@]})); then
for file in "${changed_files[@]}"; do
  case "$file" in
    .dockerignore|Dockerfile|go.mod|go.sum|scripts/detect-docker-image-targets.sh|.github/workflows/docker-image.yml)
      bot=true
      worker=true
      ;;
    scripts/install-imagemagick.sh)
      worker=true
  esac

  if file_is_in_dirs "$file" "${bot_dirs[@]}"; then
      bot=true
    fi

    if file_is_in_dirs "$file" "${worker_dirs[@]}"; then
      worker=true
    fi
  done
fi

if [[ "$write_github_output" == true ]]; then
  {
    echo "bot=$bot"
    echo "worker=$worker"
  } >> "$GITHUB_OUTPUT"
else
  printf 'bot=%s\nworker=%s\n' "$bot" "$worker"
fi
