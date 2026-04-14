# Build from source
FROM maven:3.9.8-eclipse-temurin-21 AS builder

WORKDIR /app

# Copy pom.xml and download dependencies
COPY pom.xml .
RUN mvn dependency:go-offline -B

# Copy source code
COPY src ./src

# Build application
RUN mvn clean package -DskipTests=true

# Runtime stage
FROM eclipse-temurin:21-jre-alpine

WORKDIR /app

# Copy assembled JAR from builder
COPY --from=builder /app/target/ride-service-*.jar ride-service.jar

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8082/rides/health || exit 1

# Expose port
EXPOSE 8082

# Run application
ENTRYPOINT ["java", "-XX:+UseG1GC", "-XX:MaxRAMPercentage=75.0", "-jar", "ride-service.jar"]
