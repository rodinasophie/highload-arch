# Шардирование  
Данная домашняя работа была направлена на изучение шардирования при помощи технологии  `Citus` в `PostgreSQL`. В рамках выполненной работы были проделаны следующие шаги:
1. Реализована подсистема диалогов для социальной сети
2. Поддержаны следующие `REST API` вызовы: 
    1. `/dialog/:id/send, /dialog/:id/list`
3. Развернут кластер `Citus`, состоящий из одного координатора и одного воркера.
4. Поддержано распределенное хранение всех таблиц системы по следующией логике:
    1. Все таблицы, специфичные для пользователя, шардируются по идентификатору пользователя и используется технология `colocate` для упорядочивания хранения данных одного и того же пользователя в рамках одного воркера.
    3. Подсистема диалогов шардирована по ключу `dialog_id`, состоящему из пары идентификаторов пользователей-участников диалога. Данный подход позволяет хранить все сообщения переписки в рамках одного воркера, что ускоряет подгрузку переписки. Переписки с разными пользователями могут храниться на разных воркерах.
5. Поддержан вариант решардирования по следующему принципу:
    1. Добавляем новый воркер в систему
    2. Включаем логическую репликацию на координаторе и существующих воркерах, перезагружаем существующие узлы: `alter system set wal_level = logical; run_command_on_workers('alter system set wal_level = logical');`
    3. Регистрируем нового воркера на хосте координатора: `citus_add_node('172.16.238.100', 5432);`
    4. Включаем перебалансировку шардов: `citus_rebalance_start(); citus_rebalance_wait();`

Схема базы данных:
```sql

SELECT create_distributed_table('users', 'id');
SELECT create_distributed_table('user_credentials', 'id', colocate_with => 'users');
SELECT create_distributed_table('user_tokens', 'id', colocate_with => 'users');
SELECT create_distributed_table('friends', 'id', colocate_with => 'users');
SELECT create_distributed_table('posts', 'author_user_id', colocate_with => 'users');

SELECT create_distributed_table('dialogs', 'dialog_id');

```

Для запуска стенда следует использовать следующие команды:
`make docker-reset && make docker-citus && make docker-cache && make docker-backend && make docker-run && make init-system`

Для перебалансировки следует использовать команду:
`make docker-rebalance`

`Docker-compose` файл для поддержки `Citus` - `compose.yaml`.

Сервер доступен на `http://localhost:8083/`