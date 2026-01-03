#!/usr/bin/env sh

set -eu

echo "Unpublish entrypoint script"

: "${SITE_NAME:?Error: SITE_NAME is not set or empty}"
# : "${WRANGLER_CONFIG:?Error: WRANGLER_CONFIG is not set or empty}"

cat > wrangler.jsonc <<EOF
{
	"name": "$SITE_NAME"
}
EOF

echo "Deleting worker: ${SITE_NAME}"
wrangler delete --name "${SITE_NAME}" --force

echo "Unpublish entrypoint ended"
