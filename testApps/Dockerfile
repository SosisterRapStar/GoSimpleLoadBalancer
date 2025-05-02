FROM python:3.10-slim

WORKDIR /app

RUN pip install poetry

COPY pyproject.toml poetry.lock ./
RUN poetry config virtualenvs.create false \
    && poetry install --no-root

COPY app/ ./app

# Создаём директорию для логов внутри контейнера
RUN mkdir /logs

