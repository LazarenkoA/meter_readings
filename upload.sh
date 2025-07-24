#!/bin/bash

# Проверка наличия переменной окружения
if [ -z "$REMOTE_IP" ]; then
  echo "Ошибка: переменная окружения REMOTE_IP не установлена."
  exit 1
fi

# Сборка бинарника
echo "Сборка Go-программы..."
/usr/local/go/bin/go build -o meter_bot ./main.go

# Загрузка по SFTP
sftp -i /mnt/d/ssh/mini/key artem@"$REMOTE_IP" <<EOF
put meter_bot /var/tmp/
mkdir /var/tmp/node_mos_ru
put node_mos_ru/client.js /var/tmp/node_mos_ru/client.js
put node_mos_ru/main.js /var/tmp/node_mos_ru/main.js
put node_mos_ru/package.json /var/tmp/node_mos_ru/package.json
put node_mos_ru/package-lock.json /var/tmp/node_mos_ru/package-lock.json
exit
EOF

echo "Готово. md5sum - $(md5sum meter_bot)"
