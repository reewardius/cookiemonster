while read target; do
  echo "[*] Checking: $target"
  findings=$(echo "$target" | nuclei -t cookie-extractor.yaml -silent)
  if [ -n "$findings" ]; then
    echo "$findings" | cut -d "=" -f 2 | cut -d ";" -f 1 | while read cookie; do
      # Проверяем, что cookie не пустая
      if [ -n "$cookie" ]; then
        output=$(cookiemonster -cookie "$cookie" 2>&1)
        if echo "$output" | grep -q "Success" && ! echo "$output" | grep -q "Sorry"; then
          echo "[+] Target: $target | Cookie: $cookie"
          echo "$output"
        fi
      fi
    done
  fi
done < alive_http_services.txt
