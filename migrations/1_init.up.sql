CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(200) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS posts (
    id serial primary key,
    title varchar(200) NOT NULL,
    payload TEXT NOT NULL,
    authorID int not null references users(id) on delete cascade,
    isCommentsAllowed boolean default true,
    createdAt timestamp with time zone default now()
);

CREATE TABLE IF NOT EXISTS comments (
    id serial primary key,
    payload TEXT not null,
    postID int not null references posts(id) on delete cascade,
    authorID int not null references users(id) on delete cascade,
    replyTo int references comments(id) on delete cascade,
    createdAt timestamp with time zone default now()
);

INSERT INTO users (username) VALUES ('Alice');
INSERT INTO users (username) VALUES ('Quizert');
INSERT INTO users (username) VALUES ('Alen');