CREATE TABLE IF NOT EXISTS alt_title_types (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	type_of  TEXT NOT NULL,
	CONSTRAINT alt_title_types_un UNIQUE (type_of)
);

CREATE TABLE IF NOT EXISTS anime_types (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	type_of  TEXT NOT NULL,
	CONSTRAINT anime_types_un UNIQUE (type_of)
);

CREATE TABLE IF NOT EXISTS info_types (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	type_of  TEXT NOT NULL,
	CONSTRAINT info_types_un UNIQUE (type_of)
);

CREATE TABLE IF NOT EXISTS relation_types (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	type_of  TEXT NOT NULL,
	CONSTRAINT relation_types_un UNIQUE (type_of)
);

CREATE TABLE IF NOT EXISTS seasons (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	season  TEXT NOT NULL,
	value date NULL,
	CONSTRAINT season_un UNIQUE (season)
);

CREATE TABLE IF NOT EXISTS stream_sources (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	stream_source  TEXT NOT NULL,
	CONSTRAINT stream_sources_un UNIQUE (stream_source)
);

CREATE TABLE IF NOT EXISTS animes (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	title  TEXT NOT NULL,
	description  TEXT NULL,
	mal_url  TEXT NOT NULL,
	cover  TEXT NULL,
	type_id int NULL,
	season_id int NULL,
    aired_from date NULL,
	aired_to date NULL,
	CONSTRAINT animes_un UNIQUE (mal_url),
	CONSTRAINT animes_fk FOREIGN KEY (type_id) REFERENCES anime_types(id),
	CONSTRAINT animes_fkk FOREIGN KEY (season_id) REFERENCES seasons(id)
);

CREATE TABLE IF NOT EXISTS episodes (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	anime_id int NOT NULL,
	title  TEXT NULL,
	index_of int NOT NULL,
	alt_title  TEXT NULL,
	CONSTRAINT episodes_un UNIQUE (anime_id, index_of),
	CONSTRAINT episodes_fk FOREIGN KEY (anime_id) REFERENCES animes(id)
);

CREATE TABLE IF NOT EXISTS infos (
	id  INTEGER PRIMARY KEY AUTOINCREMENT,
	info  TEXT NOT NULL,
	type_id int NOT NULL,
	CONSTRAINT infos_un UNIQUE (info),
	CONSTRAINT infos_fk FOREIGN KEY (type_id) REFERENCES info_types(id)
);

CREATE TABLE IF NOT EXISTS relations (
	root_anime_id int NOT NULL,
	related_anime_id int NOT NULL,
	type_id int NOT NULL,
	CONSTRAINT relations_un UNIQUE (root_anime_id, related_anime_id, type_id),
	CONSTRAINT relations_fk FOREIGN KEY (root_anime_id) REFERENCES animes(id),
	CONSTRAINT relations_fk_1 FOREIGN KEY (type_id) REFERENCES relation_types(id),
	CONSTRAINT relations_fkk FOREIGN KEY (related_anime_id) REFERENCES animes(id)
);

CREATE TABLE IF NOT EXISTS alt_titles (
	anime_id int NOT NULL,
	alt_title_type_id int NOT NULL,
	alt_title  TEXT NOT NULL,
	CONSTRAINT alt_titles_un UNIQUE (alt_title, anime_id, alt_title_type_id),
	CONSTRAINT alt_titles_fk FOREIGN KEY (anime_id) REFERENCES animes(id),
	CONSTRAINT alt_titles_fk_1 FOREIGN KEY (alt_title_type_id) REFERENCES alt_title_types(id)
);

CREATE TABLE IF NOT EXISTS anime_infos (
	anime_id int NOT NULL,
	info_id int NOT NULL,
	CONSTRAINT anime_infos_un UNIQUE (anime_id, info_id),
	CONSTRAINT anime_infos_fk FOREIGN KEY (anime_id) REFERENCES animes(id),
	CONSTRAINT anime_infos_fk_1 FOREIGN KEY (info_id) REFERENCES infos(id)
);

CREATE TABLE IF NOT EXISTS episode_streams (
	episode_id int NOT NULL,
	stream  TEXT NOT NULL,
	source_id int NOT NULL,
	CONSTRAINT episode_streams_un UNIQUE (episode_id, source_id, stream),
	CONSTRAINT episode_streams_un_1 UNIQUE (stream),
	CONSTRAINT episode_streams_fk FOREIGN KEY (source_id) REFERENCES stream_sources(id),
	CONSTRAINT episode_streams_fk_1 FOREIGN KEY (episode_id) REFERENCES episodes(id)
);
