FROM postgres:11.4-alpine
ADD db_script.sql /docker-entrypoint-initdb.d
