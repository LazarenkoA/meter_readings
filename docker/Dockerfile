# --- Build Go application ---
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем Go модули и код
COPY ../go.mod go.sum ./
RUN go mod download

COPY .. .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -o meter_bot .

# --- Runtime образ с Node.js и Puppeteer ---
FROM node:20-bullseye

# Установка зависимостей для Puppeteer
RUN echo "deb http://ftp.de.debian.org/debian bookworm main" > /etc/apt/sources.list && \
    echo "deb http://ftp.de.debian.org/debian bookworm-updates main" >> /etc/apt/sources.list && \
    echo "deb http://security.debian.org/debian-security bookworm-security main" >> /etc/apt/sources.list

RUN apt-get update && apt-get install -y \
    ca-certificates \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcairo2 \
    libcups2 \
    libdbus-1-3 \
    libexpat1 \
    libfontconfig1 \
    libgbm1 \
    libgcc1 \
    libglib2.0-0 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libpango-1.0-0 \
    libx11-6 \
    libx11-xcb1 \
    libxcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxi6 \
    libxrandr2 \
    libxrender1 \
    libxss1 \
    libxtst6 \
    lsb-release \
    wget \
    xdg-utils \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

# Копируем Node.js проект
WORKDIR /node_mos_ru
COPY ../node_mos_ru/package.json .
RUN npm install && node -e "console.log(require('puppeteer').executablePath())"

# Копируем остальной Node.js код
COPY ../node_mos_ru .

# Копируем Go бинарник
WORKDIR /app
COPY --from=builder /app/meter_bot .
COPY --from=builder /app/.env .env

# Устанавливаем переменные среды
ENV NODE_SCRIPT_ROOT=../node_mos_ru
ENV NODE_PATH=node
ENV PATH=$PATH:/app

# Настройка временной зоны (пример для Москвы)
RUN apt-get install -y tzdata
RUN ln -fs /usr/share/zoneinfo/Europe/Moscow /etc/localtime

# Запускаем Go-приложение
ENTRYPOINT ["./meter_bot"]
