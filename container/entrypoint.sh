#!/usr/bin/env sh

set -eu

echo "Entrypoint script"

: "${THEME_LOCATION_URL:?Error: THEME_LOCATION_URL is not set or empty}"
: "${WRANGLER_CONFIG:?Error: WRANGLER_CONFIG is not set or empty}"
: "${USER_CONTENT:?Error: USER_CONTENT is not set or empty}"

if [ -n "$USER_CONTENT" ]; then
	keys=$(echo "$USER_CONTENT" | jq -r 'keys[]')
fi

aws s3 cp $THEME_LOCATION_URL .
unzip -q theme.zip

# npm i -D wrangler@latest

npm ci --cache .npm --prefer-offline

# LOCATIONS=$(cat manifest.json | jq -r ".schemasLocation | map ({ (.name): .loader_location }) | add")
# TODO: Fix LOCATIONS
LOCATIONS=$(cat manifest.json | jq -r ".loaderLocations | map ({ (.id): .location }) | add")
for key in $keys; do
	location=$(echo $LOCATIONS | jq -r ".$key")
	content=$(echo $USER_CONTENT | jq -r ".$key")

	mkdir -p $(dirname $location)
	echo $content > $location
done


npm run build

# Zip the dist (optional)
# cd dist
# zip -rq ../dist.zip ./* 
# cd ..

echo $WRANGLER_CONFIG > "wrangler.jsonc"
wrangler deploy

# wrangler pages project create <TODO> --production-branch=main
# wrangler pages deploy ./dist --project-name=<TODO> --branch=main

# Uncommet the line below to keep the container running
# tail -f /dev/null

echo "Entrypoint ended"
