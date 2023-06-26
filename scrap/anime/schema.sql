CREATE TABLE IF NOT EXISTS public.seasons (
	id serial NOT NULL,
	name_of varchar NOT NULL,
	value date NULL,
	CONSTRAINT seasons_pk PRIMARY KEY (id),
	CONSTRAINT seasons_un UNIQUE (name_of)
);
CREATE TABLE IF NOT EXISTS public.types_of (
	id serial NOT NULL,
	name_of varchar NOT NULL,
	CONSTRAINT types_of_pk PRIMARY KEY (id),
	CONSTRAINT types_of_un UNIQUE (name_of)
);
CREATE TABLE IF NOT EXISTS public.animes (
	id serial NOT NULL,
	title varchar NOT NULL,
	title_en varchar NULL,
	title_jp varchar NULL,
	description varchar NULL,
	mal_url varchar NOT NULL,
	cover_url varchar NULL,
	aired_from date NULL,
	aired_to date NULL,
	season_id int NULL,
	type_of_id int NULL,
	CONSTRAINT newtable_pk PRIMARY KEY (id),
	CONSTRAINT newtable_un UNIQUE (mal_url),
	CONSTRAINT animes_fk FOREIGN KEY (season_id) REFERENCES public.seasons(id),
	CONSTRAINT animes_fk_1 FOREIGN KEY (type_of_id) REFERENCES public.types_of(id)
);
CREATE TABLE IF NOT EXISTS public.info_types (
	id serial NOT NULL,
	name_of varchar NOT NULL,
	CONSTRAINT infos_pk PRIMARY KEY (id),
	CONSTRAINT infos_un UNIQUE (name_of)
);
CREATE TABLE IF NOT EXISTS public.anime_infos (
	anime_id int NOT NULL,
	info_type_id int NOT NULL,
	value varchar NOT NULL,
	CONSTRAINT anime_infos_un UNIQUE (anime_id, info_type_id, value),
	CONSTRAINT anime_infos_fk FOREIGN KEY (anime_id) REFERENCES public.animes(id),
	CONSTRAINT anime_infos_fk_1 FOREIGN KEY (info_type_id) REFERENCES public.info_types(id)
);
CREATE TABLE IF NOT EXISTS public.episodes (
	id serial NOT NULL,
	title varchar NULL,
	stream_url varchar NULL,
	index_of varchar NULL,
	anime_id int NOT NULL,
	CONSTRAINT episodes_pk PRIMARY KEY (id),
	CONSTRAINT episodes_un UNIQUE (anime_id,index_of),
	CONSTRAINT episodes_fk FOREIGN KEY (anime_id) REFERENCES public.animes(id)
);