# JPEG Compressor

Простое CLI-приложение для сжатия JPEG изображений.

## Установка

```bash
go build -o jcompressor ./cmd/jcompressor
```

## Использование

```bash
# Сжать изображение (результат будет сохранён как input_compressed.jpg)
./jcompressor input.jpg

# Сжать изображение с указанием выходного файла
./jcompressor input.jpg output.jpg
```

## Структура проекта

```
.
├── cmd/
│   └── jcompressor/
│       └── main.go          # Точка входа приложения
├── internal/
│   └── compressor/
│       └── compressor.go    # Логика сжатия JPEG
├── go.mod
└── README.md
```

## Параметры

По умолчанию изображения сжимаются с качеством **50** (из 100).
