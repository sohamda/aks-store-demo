FROM rust:1.80 AS builder

RUN USER=root cargo new --bin virtual-worker

# Create a new directory for our application
WORKDIR /virtual-worker

# Copy the Cargo.toml and Cargo.lock files to the container
COPY Cargo.toml Cargo.lock ./

# Build the dependencies and cache them
RUN cargo build --release

# Remove the .rs files created by cargo
RUN rm src/*.rs

# Copy the source files to the container
ADD . ./

# Remove the dummy target directory created by cargo
RUN rm ./target/release/deps/virtual_worker*

# Build the application
RUN cargo build --release

# Create a new stage and copy the binary from the builder stage
FROM debian:bookworm-slim AS runner
WORKDIR /app

# Set the build argument for the app version number
ARG APP_VERSION=0.1.0

# Install OpenSSL
RUN apt-get update && apt-get install -y libssl-dev && rm -rf /var/lib/apt/lists/*

# Create a non-root user for running the application
RUN useradd --system --no-create-home --shell /bin/false appuser

# Copy the binary from the builder stage
COPY --from=builder --chown=appuser:appuser /virtual-worker/target/release/virtual-worker /app

# Switch to non-root user
USER appuser

# Set the environment variable for the app version number
ENV APP_VERSION=$APP_VERSION

# Run the application
CMD ["./virtual-worker"]