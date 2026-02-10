SELECT 'CREATE DATABASE celery_results'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'celery_results')\gexec
