FROM eclipse-temurin:21-jre-alpine
LABEL authors="vikazeni"

WORKDIR /application

COPY /ask-core/target/ask-core-1.0.0.jar ./app.jar

ENV JAVA_OPTS='-XX:MaxRAMPercentage=60.0'

ENTRYPOINT ["java", "-jar", "./app.jar", "$JAVA_OPTS"]
