CREATE TABLE IF NOT EXISTS public.seasons (
	id serial NOT NULL,
	name_of varchar NOT NULL,
	value date NULL,
	CONSTRAINT seasons_pk PRIMARY KEY (id),
	CONSTRAINT seasons_un UNIQUE (name_of)
);
CREATE TABLE IF NOT EXISTS public.anime_types (
	id serial NOT NULL,
	name_of varchar NOT NULL,
	CONSTRAINT anime_types_pk PRIMARY KEY (id),
	CONSTRAINT anime_types_un UNIQUE (name_of)
);
CREATE TABLE IF NOT EXISTS public.animes (
	id serial NOT NULL,
	title varchar NOT NULL,
	description varchar NULL,
	mal_url varchar NOT NULL,
	cover_url varchar NULL,
	aired_from date NULL,
	aired_to date NULL,
	season_id int NULL,
	type_id int NULL,
	CONSTRAINT animes_pk PRIMARY KEY (id),
	CONSTRAINT animes_un UNIQUE (mal_url),
	CONSTRAINT animes_fk FOREIGN KEY (season_id) REFERENCES public.seasons(id),
	CONSTRAINT animes_fk_1 FOREIGN KEY (type_id) REFERENCES public.anime_types(id)
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
CREATE TABLE IF NOT EXISTS public.info_types (
	id serial NOT NULL,
	name_of varchar NOT NULL,
	CONSTRAINT info_types_pk PRIMARY KEY (id),
	CONSTRAINT info_types_un UNIQUE (name_of)
);
CREATE TABLE IF NOT EXISTS public.infos (
	id serial NOT NULL,
	value varchar NOT NULL,
	type_id int NOT NULL,
	CONSTRAINT infos_pk PRIMARY KEY (id),
	CONSTRAINT infos_un UNIQUE (value),
	CONSTRAINT infos_fk FOREIGN KEY (type_id) REFERENCES public.info_types(id)
);
CREATE TABLE IF NOT EXISTS public.anime_infos (
	anime_id int NOT NULL,
	info_id int NOT NULL,
	CONSTRAINT anime_infos_un UNIQUE (anime_id, info_id),
	CONSTRAINT anime_infos_fk FOREIGN KEY (anime_id) REFERENCES public.animes(id),
	CONSTRAINT anime_infos_fk_1 FOREIGN KEY (info_id) REFERENCES public.infos(id)
);
CREATE TABLE IF NOT EXISTS public.relations (
	id serial NOT NULL,
	name_of varchar NOT NULL,
	CONSTRAINT relations_pk PRIMARY KEY (id),
	CONSTRAINT relations_un UNIQUE (name_of)
);
CREATE TABLE IF NOT EXISTS public.anime_relations (
	anime_id int NOT NULL,
	related_id int NOT NULL,
	relation_id int NULL,
	CONSTRAINT anime_relations_un UNIQUE (anime_id,related_id,relation_id),
	CONSTRAINT anime_relations_fk FOREIGN KEY (anime_id) REFERENCES public.animes(id),
	CONSTRAINT anime_relations_fk_1 FOREIGN KEY (related_id) REFERENCES public.animes(id),
	CONSTRAINT anime_relations_fk_2 FOREIGN KEY (relation_id) REFERENCES public.relations(id)
);
