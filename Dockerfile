# Stage 1: Build the application
FROM maven:3.9.6-eclipse-temurin-21 AS builder

# Настройка зеркал для Maven
RUN mkdir -p /root/.m2 && \
    echo '<?xml version="1.0" encoding="UTF-8"?> \
    <settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"> \
        <mirrors> \
            <mirror> \
                <id>aliyun</id> \
                <url>https://maven.aliyun.com/repository/public</url> \
                <mirrorOf>central</mirrorOf> \
            </mirror> \
        </mirrors> \
    </settings>' > /root/.m2/settings.xml

WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline -B

COPY src ./src
RUN mvn clean package -DskipTests && \
    # Очищаем ненужные файлы в том же слое
    rm -rf /root/.m2/repository/*/sources

# Stage 2: Run the application - ИСПОЛЬЗУЕМ ОПТИМИЗИРОВАННЫЙ ОБРАЗ
FROM eclipse-temurin:21-jre-jammy

# Создаем пользователя и директории
RUN addgroup --system --gid 1001 appuser && \
    adduser --system --uid 1001 --gid 1001 appuser && \
    mkdir -p /app/logs

WORKDIR /app

# Копируем только jar файл
COPY --from=builder /app/target/*.jar app.jar

# Переключаемся на непривилегированного пользователя
USER appuser

EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]