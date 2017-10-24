FROM mysql

# Add a database schema creation script.
# When the container instance starts,
# it will execute any .sql scripts that we put into
# /docker-entrypoint-initdb.d/ directory.
COPY schema.sql /docker-entrypoint-initdb.d/schema.sql