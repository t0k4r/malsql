CREATE TABLE IF NOT EXISTS public.alt_title_types (
	id serial NOT NULL,
	type_of varchar NOT NULL,
	CONSTRAINT alt_title_types_pk PRIMARY KEY (id),
	CONSTRAINT alt_title_types_un UNIQUE (type_of)
);

CREATE TABLE IF NOT EXISTS public.anime_types (
	id serial NOT NULL,
	type_of varchar NOT NULL,
	CONSTRAINT anime_types_pk PRIMARY KEY (id),
	CONSTRAINT anime_types_un UNIQUE (type_of)
);

CREATE TABLE IF NOT EXISTS public.info_types (
	id serial NOT NULL,
	type_of varchar NOT NULL,
	CONSTRAINT info_types_pk PRIMARY KEY (id),
	CONSTRAINT info_types_un UNIQUE (type_of)
);

CREATE TABLE IF NOT EXISTS public.relation_types (
	id serial NOT NULL,
	type_of varchar NOT NULL,
	CONSTRAINT relation_types_pk PRIMARY KEY (id),
	CONSTRAINT relation_types_un UNIQUE (type_of)
);

CREATE TABLE IF NOT EXISTS public.seasons (
	id serial NOT NULL,
	season varchar NOT NULL,
	value date NULL,
	CONSTRAINT season_pk PRIMARY KEY (id),
	CONSTRAINT season_un UNIQUE (season)
);

CREATE TABLE IF NOT EXISTS public.stream_sources (
	id serial NOT NULL,
	stream_source varchar NOT NULL,
	CONSTRAINT stream_sources_pk PRIMARY KEY (id),
	CONSTRAINT stream_sources_un UNIQUE (stream_source)
);

CREATE TABLE IF NOT EXISTS public.animes (
	id serial NOT NULL,
	title varchar NOT NULL,
	description varchar NULL,
	mal_url varchar NOT NULL,
	cover varchar NULL,
	type_id int NULL,
	season_id int NULL,
    aired_from date NULL,
	aired_to date NULL,
	CONSTRAINT animes_pk PRIMARY KEY (id),
	CONSTRAINT animes_un UNIQUE (mal_url),
	CONSTRAINT animes_fk FOREIGN KEY (type_id) REFERENCES public.anime_types(id),
	CONSTRAINT animes_fkk FOREIGN KEY (season_id) REFERENCES public.seasons(id)
);

CREATE TABLE IF NOT EXISTS public.episodes (
	id serial NOT NULL,
	anime_id int NOT NULL,
	title varchar NULL,
	index_of int NOT NULL,
	alt_title varchar NULL,
	CONSTRAINT episodes_pk PRIMARY KEY (id),
	CONSTRAINT episodes_un UNIQUE (anime_id, index_of),
	CONSTRAINT episodes_fk FOREIGN KEY (anime_id) REFERENCES public.animes(id)
);

CREATE TABLE IF NOT EXISTS public.infos (
	id serial NOT NULL,
	info varchar NOT NULL,
	type_id int NOT NULL,
	CONSTRAINT infos_pk PRIMARY KEY (id),
	CONSTRAINT infos_un UNIQUE (info),
	CONSTRAINT infos_fk FOREIGN KEY (type_id) REFERENCES public.info_types(id)
);

CREATE TABLE IF NOT EXISTS public.relations (
	root_anime_id int NOT NULL,
	related_anime_id int NOT NULL,
	type_id int NOT NULL,
	CONSTRAINT relations_un UNIQUE (root_anime_id, related_anime_id, type_id),
	CONSTRAINT relations_fk FOREIGN KEY (root_anime_id) REFERENCES public.animes(id),
	CONSTRAINT relations_fk_1 FOREIGN KEY (type_id) REFERENCES public.relation_types(id),
	CONSTRAINT relations_fkk FOREIGN KEY (related_anime_id) REFERENCES public.animes(id)
);

CREATE TABLE IF NOT EXISTS public.alt_titles (
	anime_id int NOT NULL,
	alt_title_type_id int NOT NULL,
	alt_title varchar NOT NULL,
	CONSTRAINT alt_titles_un UNIQUE (alt_title, anime_id, alt_title_type_id),
	CONSTRAINT alt_titles_fk FOREIGN KEY (anime_id) REFERENCES public.animes(id),
	CONSTRAINT alt_titles_fk_1 FOREIGN KEY (alt_title_type_id) REFERENCES public.alt_title_types(id)
);

CREATE TABLE IF NOT EXISTS public.anime_infos (
	anime_id int NOT NULL,
	info_id int NOT NULL,
	CONSTRAINT anime_infos_un UNIQUE (anime_id, info_id),
	CONSTRAINT anime_infos_fk FOREIGN KEY (anime_id) REFERENCES public.animes(id),
	CONSTRAINT anime_infos_fk_1 FOREIGN KEY (info_id) REFERENCES public.infos(id)
);

CREATE TABLE IF NOT EXISTS public.episode_streams (
	episode_id int NOT NULL,
	stream varchar NOT NULL,
	source_id int NOT NULL,
	CONSTRAINT episode_streams_un UNIQUE (episode_id, source_id, stream),
	CONSTRAINT episode_streams_fk FOREIGN KEY (source_id) REFERENCES public.stream_sources(id),
	CONSTRAINT episode_streams_fk_1 FOREIGN KEY (episode_id) REFERENCES public.episodes(id)
);