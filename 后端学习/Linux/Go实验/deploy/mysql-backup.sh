#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

DB_NAME="${DB_NAME:-shortlink}"
DEFAULTS_FILE="${DEFAULTS_FILE:-/etc/shortlink/mysql-backup.cnf}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/shortlink/mysql}"
KEEP_DAYS="${KEEP_DAYS:-14}"

[[ -r "$DEFAULTS_FILE" ]] || {
	echo "cannot read MySQL defaults file: $DEFAULTS_FILE" >&2
	exit 1
}
[[ "$KEEP_DAYS" =~ ^[0-9]+$ ]] || {
	echo "KEEP_DAYS must be an integer" >&2
	exit 1
}

install -d -m 0700 "$BACKUP_DIR"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
tmp_file="$(mktemp "${BACKUP_DIR}/.${DB_NAME}.${timestamp}.XXXXXX.sql.gz")"
final_file="${BACKUP_DIR}/${DB_NAME}.${timestamp}.sql.gz"

cleanup() {
	rm -f -- "$tmp_file"
}
trap cleanup EXIT

mysqldump \
	--defaults-extra-file="$DEFAULTS_FILE" \
	--single-transaction \
	--quick \
	--routines \
	--triggers \
	--events \
	--set-gtid-purged=OFF \
	"$DB_NAME" | gzip -9 >"$tmp_file"

gzip -t "$tmp_file"
mv -- "$tmp_file" "$final_file"
sha256sum "$final_file" >"${final_file}.sha256"
trap - EXIT

find "$BACKUP_DIR" -maxdepth 1 -type f \
	\( -name "${DB_NAME}.*.sql.gz" -o -name "${DB_NAME}.*.sql.gz.sha256" \) \
	-mtime "+$KEEP_DAYS" -delete

printf 'backup created: %s\n' "$final_file"
