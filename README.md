# Office-ups API

Веб-сервис для управления ИБП. Go + PostgreSQL + GORM + MinIO.

## Таблицы БД

### office_users
| Поле | Тип | Описание |
|---|---|---|
| id | SERIAL PK | ID пользователя |
| user_login | VARCHAR(50) | Логин (уникальный) |
| password_hash | VARCHAR(100) | Хэш пароля |
| is_moderator | BOOLEAN | Признак модератора |

### uninterruptible_power_supplies
| Поле | Тип | Описание |
|---|---|---|
| id | SERIAL PK | ID услуги |
| ups_name | VARCHAR(100) | Название ИБП |
| description | TEXT | Описание |
| status | VARCHAR(20) | черновик / опубликован / удален |
| image_url | VARCHAR(255) | URL картинки в MinIO |
| video_url | VARCHAR(255) | URL видео в MinIO |
| capacity | NUMERIC(10,2) | Ёмкость (А·ч) |
| max_current | NUMERIC(10,2) | Максимальный ток (А) |
| created_at | TIMESTAMP | Дата создания |
| updated_at | TIMESTAMP | Дата обновления |
| creator_id | INT FK | ID создателя |

### uninterruptible_power_supply_likes
| Поле | Тип | Описание |
|---|---|---|
| id | SERIAL PK | ID лайка |
| user_id | INT FK | ID пользователя |
| ups_id | INT FK | ID услуги |

## HTTP-методы

### Домен услуги

| Метод | URL | Код | Описание |
|---|---|---|---|
| GET | /api/ups | 200 | Список с фильтром (min_capacity, max_capacity) |
| GET | /api/ups/reels | 200/404 | Первая запись ленты |
| GET | /api/ups/reels?id=&next=true | 200/404 | Следующая запись |
| GET | /api/ups/draft | 200/404 | Черновик пользователя |
| POST | /api/ups | 201/409 | Создать + загрузить файлы |
| PUT | /api/ups/:id/publish | 200/400 | Публикация |
| DELETE | /api/ups/:id | 204 | Мягкое удаление |
| POST | /api/ups/:id/like | 204/400 | Лайк (value: 0/1) |

### Домен пользователя

| Метод | URL | Код | Описание |
|---|---|---|---|
| POST | /api/auth/register | 201/409 | Регистрация |
| POST | /api/auth/login | 200 | Вход (заглушка) |
| POST | /api/auth/logout | 200 | Выход (заглушка) |

## Запуск

1. Запустить Docker Desktop
2. `docker compose up -d`
3. `docker exec ups_minio mc alias set myminio http://localhost:9000 minioadmin minioadmin`
4. `docker exec ups_minio mc anonymous set public myminio/ups-media`
5. `go run cmd/migrate/main.go`
6. `go run cmd/ups_service/main.go`