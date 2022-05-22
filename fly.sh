#!/bin/bash

function setSkiverConfigSecret() {
  local FILE="${1}"
  if [ ! -f "$FILE" ]; then
      echo "$FILE does not exist ."
      return
  fi
  local b64="$(base64 ${FILE})"

  echo "Secret is $(wc -c <<< ${b64}) bytes (from file '${FILE}')"

  flyctl secrets set SKIVER_CONFIG_B64="$b64" || echo "Probably did not change??"
}

setSkiverConfigSecret "$1"
