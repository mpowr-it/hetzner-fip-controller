#!/bin/bash
set -euo pipefail

# Lokale Variablen (statisch für Test)
GITHUB_REPOSITORY="mpowr-it/cu20240002-paligo-marketplace"
GITHUB_WORKFLOW="build-and-sign"
GITHUB_SHA="abcdef1234567890abcdef1234567890abcdef12"
VERSION="v0.1"

# Vorbereitung
mkdir -p merged provenance signed-provenance

# Input-JSON
echo "📦 Lade Images-Informationen..."
cat merged/all-images.json

# Image-Loop
echo "🔄 Generiere Provenance-Daten..."
COUNT=$(jq length merged/all-images.json)

for i in $(seq 0 $((COUNT - 1))); do
  IMAGE=$(jq -r ".[$i].image" merged/all-images.json)
  TAG=$(jq -r ".[$i].tag" merged/all-images.json)
  DIGEST=$(jq -r ".[$i].digest" merged/all-images.json)
  FULL_DIGEST=$(jq -r ".[$i].full_digest" merged/all-images.json)

  SHA_ONLY=$(echo "$DIGEST" | cut -d':' -f2)
  OUTPUT="provenance/provenance-${TAG}.json"

  echo "🔵 Verarbeite $IMAGE:$TAG ($FULL_DIGEST)"

  # Erzeuge Provenance JSON
  jq -n \
    --arg image "$IMAGE" \
    --arg sha256 "$SHA_ONLY" \
    --arg repo "https://github.com/${GITHUB_REPOSITORY}" \
    --arg commit "$GITHUB_SHA" \
    --arg workflow "$GITHUB_WORKFLOW" \
    --arg version "$VERSION" \
    --arg buildStartedOn "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
    --arg buildFinishedOn "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
    '{
      "_type": "https://in-toto.io/Statement/v0.1",
      "subject": [{
        "name": $image,
        "digest": { "sha256": $sha256 }
      }],
      "predicateType": "https://slsa.dev/provenance/v0.2",
      "predicate": {
        "builder": {
          "id": $repo,
          "version": $version
        },
        "buildType": "https://github.com/Attestations/GitHubActions@v1",
        "invocation": {
          "configSource": {
            "uri": ("git+" + $repo + ".git"),
            "digest": { "sha1": $commit },
            "entryPoint": $workflow
          }
        },
        "buildConfig": {},
        "metadata": {
          "buildStartedOn": $buildStartedOn,
          "buildFinishedOn": $buildFinishedOn
        },
        "materials": []
      }
    }' > "$OUTPUT"

  echo "✅ Provenance erstellt: $OUTPUT"

  # Cosign Test
  if command -v cosign &> /dev/null; then
    echo "🔏 Cosign Attestierung (lokal) für $FULL_DIGEST"
    cosign attest --yes --predicate "$OUTPUT" --type slsaprovenance "$FULL_DIGEST" || echo "⚠️ Cosign attest Fehler erwartet (ohne OIDC oder Keyless Setup)"
  else
    echo "⚠️ Cosign nicht installiert – Skipping cosign attest."
  fi

done

echo "🏁 Alle Provenance-Daten generiert und getestet."
