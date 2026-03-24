#!/bin/bash

set -e

PROTO_DIR="proto"
GEN_DIR="proto/gen"

# Очистка предыдущей генерации
rm -rf ${GEN_DIR}
mkdir -p ${GEN_DIR}

# Генерация кода для каждого .proto файла
for proto in ${PROTO_DIR}/*.proto; do
    # Получаем имя файла без расширения
    filename=$(basename "$proto" .proto)
    echo "Generating ${proto}..."
    
    # Создаём подпапку для каждого пакета
    mkdir -p ${GEN_DIR}/${filename}
    
    protoc \
        --proto_path=${PROTO_DIR} \
        --go_out=${GEN_DIR}/${filename} \
        --go_opt=paths=source_relative \
        --go-grpc_out=${GEN_DIR}/${filename} \
        --go-grpc_opt=paths=source_relative \
        ${proto}
done

echo "Proto generation completed"