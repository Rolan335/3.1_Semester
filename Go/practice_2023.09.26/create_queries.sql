--docker exec -it 4892b79409ccef90033f1b3fa858ebad2a1c91b819640c4a46ec22584ab6357c psql -U postgres

CREATE TABLE MONITORS(
	ID_MONITORS BIGSERIAL NOT NULL CONSTRAINT PK_MONITORS PRIMARY KEY,
	DISPLAY_ID BIGINT NOT NULL REFERENCES DISPLAYS (ID_DISPLAYS),
	GSYNC_PREMIUM BOOLEAN NOT NULL,
	CURVED BOOLEAN NOT NULL
);

CREATE TABLE DISPLAYS(
	ID_DISPLAYS BIGSERIAL NOT NULL CONSTRAINT PK_DISPLAYS PRIMARY KEY,
	DIAGONAL REAL NOT NULL,
	RESOLUTION TEXT NOT NULL,
	TYPE TEXT NOT NULL,
	GSYNC BOOLEAN NOT NULL
);

CREATE TABLE USERS(
    ID_USER BIGSERIAL NOT NULL CONSTRAINT PK_USERS PRIMARY KEY,
    USERNAME TEXT NOT NULL,
    PASSWORD TEXT NOT NULL, --Хранится хэш
    EMAIL TEXT NOT NULL,
    IS_ADMIN BOOLEAN NULL
);