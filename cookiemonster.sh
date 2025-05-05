#!/bin/bash

# bash cookiemonster.sh
# bash cookiemonster.sh -debug

DEBUG=false

# Обработка аргументов
if [ "$1" == "-debug" ]; then
  DEBUG=true
fi

# Основной цикл по целям
while IFS= read -r target; do
  echo "[*] Checking: $target"

  findings=$(echo "$target" | nuclei -t cookie-monster-extended.yaml -silent)

  if [ -n "$findings" ]; then
    echo "$findings" \
    | grep -oP '[a-zA-Z0-9_-]+=[^;]+' \
    | grep -viE '^(Path|Domain|Expires|HttpOnly|Secure|SameSite|Max-Age)=' \
    | while IFS= read -r pair; do
        key=$(echo "$pair" | cut -d '=' -f1)
        value=$(echo "$pair" | cut -d '=' -f2-)

        # Проверка, что значение не пустое
        if [ -n "$value" ]; then
          $DEBUG && echo "[DEBUG] Testing cookie: $pair"
          output=$(cookiemonster -cookie "$pair" 2>&1)

          if echo "$output" | grep -q "Success" && ! echo "$output" | grep -q "Sorry"; then
            echo "[+] Target: $target | Cookie: $pair"
            echo "$output"
          fi
        fi
      done
  fi
done < targets.txt
