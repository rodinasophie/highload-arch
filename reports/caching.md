# Кеширование  
Данная домашняя работа была направлена на изучение кеширования в высоконагруженных системах. В качестве кеша был использован Redis, инвалидация кеша поддерживается через опережающее кеширование, все запросы на чтение с бэкенда идут в кеш, раз в указанный период происходит обновление кеша и данные синхронизируются с базой. В кеше хранятся последние 1000 постов.

Схема базы данных:
```sql
CREATE TABLE IF NOT EXISTS friends (
    user_id UUID REFERENCES users(id)  NOT NULL,
    friend_id UUID REFERENCES users(id) NOT NULL,
    PRIMARY KEY(user_id, friend_id) 
);

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_user_id UUID NOT NULL,
    text VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

```

В кеше помимо списка постов, хранится маппинг пользователь-друг для корректной фильтрации выдачи.

Следующие докер-контейнеры были созданы для сборки стенда:
1. База данных(лидер + 1 реплика)
2. Сервер бэкенда
3. Redis-кеш.

В рамках данной домашней работы также были реализованы следующие эндпоинты:
1. `/friend/add, /friend/delete`
2. `/post/create, /post/update, /post/delete, /post/get`
3. `/post/feed`

Генерация постов происходит при помощи Python-библиотеки `faker`.

Дополнительно был создан скрипт инициализации системы при помощи REST API:
1. Создание пользователей из ранее сгенеренных данных
2. Создание пар друзей по указанному правилу(см. файл `generate/initialize_system.py`)
3. Создание постов пользователями
4. Получение списка постов(по умолчанию `offset=0, limit=2`) 

Для запуска стенда следует использовать следующие команды:
`make docker-reset && make docker-init && make docker-cache && make docker-backend && make docker-run && make init-system`

Сервер доступен на `http://localhost:8083/`