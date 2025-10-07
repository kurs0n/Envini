# Use the official PostgreSQL image from Docker Hub
FROM postgres:16-alpine

# Set environment variables for the default user, password, and database
ENV POSTGRES_USER=envini
ENV POSTGRES_PASSWORD=envini
ENV POSTGRES_DB=envini
ENV POSTGRES_PORT=5433

# Expose the PostgreSQL port
EXPOSE 5433

# Copy custom PostgreSQL configuration
COPY postgresql.conf /etc/postgresql/postgresql.conf

# Set the custom configuration file
CMD ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]