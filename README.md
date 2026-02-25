# JPEG Compressor

Простое CLI-приложение для сжатия JPEG изображений.

[![Lint](https://github.com/dalbezh/jcompressor/actions/workflows/lint.yml/badge.svg)](https://github.com/dalbezh/jcompressor/actions/workflows/lint.yml)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/dalbezh/jcompressor/badges/quality-score.png?b=main)](https://scrutinizer-ci.com/g/dalbezh/jcompressor/?branch=main)

## Установка
## Используя `make`

Показать окружение сборки:
```sh
make env
```

Построить бинарник:
```sh
make build
# В результате появится ./build/jcompressor
```

Установить (по умолчанию в /usr/local/bin):
```sh
make install
# если нужно sudo — Makefile сам использует sudo при необходимости
```

Установить в кастомный префикс (пример для локальной установки в /opt):
```sh
make install PREFIX=/opt
```

Удалить установленный бинарник:
```sh
make uninstall
```

Полная очистка артефактов сборки:
```sh
make clean
```

## Используя `go`

```sh
go build -o ./build/jcompressor ./cmd/jcompressor
```

## Параметры запуска

Запустить собранный бинарник:
```sh
jcompressor --help
```

```
Usage: jcompressor [flags] <input.jpg> [output_dir]

Flags:
  -h	show help
  -help
    	show help
  -q int
    	JPEG quality (1-100) (default 50)
  -quality int
    	JPEG quality (1-100) (default 50)
  -w	also create WebP version
  -webp
    	also create WebP version

If output_dir is omitted, files will be saved to ./compressed
```
