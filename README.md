# JPEG Compressor

Простое CLI-приложение для сжатия JPEG изображений.

[![Lint](https://github.com/dalbezh/jcompressor/actions/workflows/lint.yml/badge.svg)](https://github.com/dalbezh/jcompressor/actions/workflows/lint.yml)

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

## Параметры

По умолчанию изображения сжимаются с качеством **50** (из 100).
