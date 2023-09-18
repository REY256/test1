CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    surname VARCHAR(256) NOT NULL,
    patronymic VARCHAR(256),
    age SMALLINT NOT NULL,
    gender VARCHAR(6) NOT NULL
);

CREATE TABLE countries (
    user_id BIGINT NOT NULL REFERENCES users(id),
    country_id VARCHAR(10) NOT NULL,
    probability REAL NOT NULL,
    UNIQUE(user_id, country_id)
);

