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

# Run application with optimized JVM args for low-resource environment (3GB RAM Node)
# -Xms256m: Start with 256MB heap
# -Xmx512m: Cap heap at 512MB to leave room for OS and other pods
# -XX:+ExitOnOutOfMemoryError: Ensure Pod restarts if it hits OOM instead of hanging the node
ENTRYPOINT ["java", "-Xms256m", "-Xmx512m", "-XX:+UseG1GC", "-XX:+ExitOnOutOfMemoryError", "-jar", "ride-service.jar"]
