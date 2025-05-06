#!/bin/bash

DEBUG=false
EXPRESS=false

# Функция для вывода справки
show_help() {
  echo "Usage: bash cookie3.sh [OPTIONS]"
  echo
  echo "Options:"
  echo "  -debug           Enable debug mode"
  echo "  -express         Enable express mode"
  echo "  -help            Show this help message"
  echo
}

# Обработка аргументов
for arg in "$@"; do
  case "$arg" in
    -debug) DEBUG=true ;;
    -express) EXPRESS=true ;;
    -help) show_help; exit 0 ;;
    *) echo "Unknown option: $arg"; show_help; exit 1 ;;
  esac
done

# Основной цикл по целям
while IFS= read -r target; do
  echo "[*] Checking: $target"

  # Очистка переменных перед каждой итерацией
  unset token
  unset cookie_map

  # Извлечение cookies с помощью Nuclei
  findings=$(echo "$target" | nuclei -t cookie-extractor.yaml -silent)

  if [ -n "$findings" ]; then
    # Разбираем cookies в пары ключ=значение
    mapfile -t pairs < <(echo "$findings" | grep -oP '[a-zA-Z0-9._-]+=[^;]+' | grep -viE '^(Path|Domain|Expires|HttpOnly|Secure|SameSite|Max-Age)=')

    declare -A cookie_map

    # Заполнение карты cookies
    for pair in "${pairs[@]}"; do
      key=$(echo "$pair" | cut -d '=' -f1)
      value=$(echo "$pair" | cut -d '=' -f2-)
      cookie_map["$key"]="$value"
    done

    # Обработка cookies
    for key in "${!cookie_map[@]}"; do
      value="${cookie_map[$key]}"

      # Проверка, если используется EXPRESS режим, то комбинируем с подписью
      if $EXPRESS && [[ -n "${cookie_map[$key.sig]}" ]]; then
        value="$value^${cookie_map[$key.sig]}"
        unset cookie_map[$key.sig]
      fi

      # Если значение cookie существует, тестируем его
      if [ -n "$value" ]; then
        if $DEBUG; then
          echo "[DEBUG] Testing cookie: $key=$value"
        fi

        # В зависимости от режима, выполняем тестирование
        if $EXPRESS; then
          output=$(cookiemonster -cookie "$key=$value")
        else
          output=$(cookiemonster -cookie "$value")
        fi

        # Если тестирование прошло успешно, выводим результат
        if echo "$output" | grep -q "Success" && ! echo "$output" | grep -q "Sorry"; then
          echo "[+] Target: $target | Cookie: $key=$value"
          echo "$output"
        fi
      fi
    done
  fi
done < targets.txt
