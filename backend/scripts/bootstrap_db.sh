#!/usr/bin/env bash

set -euo pipefail

MODE="${1:-fresh}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BASELINE_FILE="${BACKEND_DIR}/schema/20260320_000000_baseline.sql"
MIGRATIONS_DIR="${BACKEND_DIR}/migrations"
ENV_FILE="${BACKEND_DIR}/.env"

if [[ -z "${DATABASE_URL:-}" && -f "${ENV_FILE}" ]]; then
    set -a
    # shellcheck disable=SC1090
    . "${ENV_FILE}"
    set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
    echo "DATABASE_URL is required (export it or set it in backend/.env)" >&2
    exit 1
fi

run_sql_file() {
    local file_path="$1"
    echo "==> applying ${file_path#${BACKEND_DIR}/}"
    psql "${DATABASE_URL}" -X -v ON_ERROR_STOP=1 -f "${file_path}"
}

case "${MODE}" in
    fresh)
        run_sql_file "${BASELINE_FILE}"
        ;;
    upgrade)
        ;;
    *)
        echo "usage: $0 [fresh|upgrade]" >&2
        exit 1
        ;;
esac

shopt -s nullglob
for migration in "${MIGRATIONS_DIR}"/*.sql; do
    run_sql_file "${migration}"
done

echo "schema bootstrap completed in ${MODE} mode"
