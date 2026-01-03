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

# Try npm ci first if lockfile exists and is valid, but always fall back to npm install on any error
if [ -f package-lock.json ] || [ -f npm-shrinkwrap.json ]; then
	npm ci --cache .npm --prefer-offline
else
	npm install --cache .npm --prefer-offline
fi

# Extract locations from contentTypes array: map name -> loaderLocationPath
LOCATIONS=$(cat manifest.json | jq -r ".contentTypes | map({(.name): .loaderLocationPath}) | add")
if [ "$LOCATIONS" = "{}" ] || [ -z "$LOCATIONS" ]; then
	echo "ERROR: Failed to extract locations from manifest.json contentTypes"
	cat manifest.json
	exit 1
fi

for key in $keys; do
	location=$(echo "$LOCATIONS" | jq -r ".[\"$key\"] // empty")
	if [ -z "$location" ] || [ "$location" = "null" ]; then
		echo "WARNING: No location found for key: $key, skipping"
		echo "Available keys in manifest: $(echo "$LOCATIONS" | jq -r 'keys[]' | tr '\n' ' ')"
		continue
	fi
	
	content=$(echo "$USER_CONTENT" | jq -c "[.$key[] | 
		.data as \$data | 
		.resolved as \$resolved |
		(\$data // {}) as \$base |
		if \$resolved and (\$resolved | type) == \"object\" then
			reduce (\$resolved | keys[]) as \$relKey (\$base;
				(\$data[\$relKey] // null) as \$originalValue |
				if \$originalValue == null or \$originalValue == {} or \$originalValue == [] or 
				   ((\$originalValue | type) == \"object\" and (\$originalValue | to_entries | map(.value) | all(. == null or . == \"\" or . == [] or . == {}))) then
					if (\$originalValue | type) == \"object\" and (\$resolved[\$relKey] | type) == \"array\" and (\$resolved[\$relKey] | length) > 0 then
						.[\$relKey] = \$resolved[\$relKey][0]
					elif (\$originalValue | type) == \"array\" and (\$resolved[\$relKey] | type) == \"array\" then
						.[\$relKey] = \$resolved[\$relKey]
					else
						.[\$relKey] = \$resolved[\$relKey]
					end
				else . end
			)
		else \$base end
	]")
	if [ -z "$content" ] || [ "$content" = "null" ] || [ "$content" = "[]" ]; then
		echo "WARNING: No content found for key: $key, skipping"
		continue
	fi
	echo "$content"
	
	mkdir -p "$(dirname "$location")"
	echo "$content" > "$location"
	
	echo "INFO: Wrote content for collection '$key' to loader location '$location'"
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
