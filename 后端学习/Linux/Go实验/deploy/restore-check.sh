#!/usr/bin/env bash
set -Eeuo pipefail

BACKUP_FILE="${1:-}"
DEFAULTS_FILE="${DEFAULTS_FILE:-/etc/shortlink/mysql-backup.cnf}"
SOURCE_DB="${SOURCE_DB:-shortlink}"
RESTORE_DB="${RESTORE_DB:-shortlink_restore_check}"

[[ -n "$BACKUP_FILE" ]] || {
	echo "usage: $0 /path/to/backup.sql.gz" >&2
	exit 1
}
[[ -r "$BACKUP_FILE" ]] || {
	echo "backup is not readable: $BACKUP_FILE" >&2
	exit 1
}
[[ -r "$DEFAULTS_FILE" ]] || {
	echo "defaults file is not readable: $DEFAULTS_FILE" >&2
	exit 1
}
[[ "$RESTORE_DB" =~ ^[0-9A-Za-z_]+$ ]] || {
	echo "RESTORE_DB contains unsafe characters" >&2
	exit 1
}
[[ "$RESTORE_DB" != "$SOURCE_DB" ]] || {
	echo "RESTORE_DB must differ from SOURCE_DB" >&2
	exit 1
}

gzip -t "$BACKUP_FILE"

mysql --defaults-extra-file="$DEFAULTS_FILE" \
	-e "DROP DATABASE IF EXISTS \`${RESTORE_DB}\`; CREATE DATABASE \`${RESTORE_DB}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;"

gunzip -c "$BACKUP_FILE" | mysql --defaults-extra-file="$DEFAULTS_FILE" "$RESTORE_DB"

table_count="$(mysql --defaults-extra-file="$DEFAULTS_FILE" --batch --skip-column-names \
	-e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='${RESTORE_DB}';")"

printf 'restore check succeeded: database=%s tables=%s\n' "$RESTORE_DB" "$table_count"
printf 'the verification database is intentionally kept for inspection; remove it manually after checking\n'
