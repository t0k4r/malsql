CREATE TABLE IF NOT EXISTS public.anime_types (
	type_id serial4 NOT NULL,
	type_name varchar NULL,
	CONSTRAINT anime_type_pk PRIMARY KEY (type_id),
	CONSTRAINT anime_type_un UNIQUE (type_name)
);
CREATE TABLE IF NOT EXISTS public.anime_seasons (
	season_id serial4 NOT NULL,
	season_name varchar NULL,
	CONSTRAINT anime_season_pk PRIMARY KEY (season_id),
	CONSTRAINT anime_season_un UNIQUE (season_name)
);
CREATE TABLE IF NOT EXISTS public.animes (
	anime_id serial4 NOT NULL,
	anime_title varchar NOT NULL,
	season_id int4 NULL,
	type_id int4 NULL,
	anime_title_en varchar NULL,
	anime_title_jp varchar NULL,
	anime_start date NULL,
	anime_end date NULL,
	anime_description varchar NULL,
	anime_mal_url varchar NOT NULL,
	anime_img_url varchar NULL,
	CONSTRAINT animes_pk PRIMARY KEY (anime_id),
	CONSTRAINT animes_un UNIQUE (anime_mal_url),
	CONSTRAINT animes_fk FOREIGN KEY (season_id) REFERENCES public.anime_seasons(season_id),
	CONSTRAINT animes_fk_1 FOREIGN KEY (type_id) REFERENCES public.anime_types(type_id)
);
CREATE TABLE IF NOT EXISTS public.genres (
	genre_id serial4 NOT NULL,
	genre_name varchar NULL,
	CONSTRAINT genres_pk PRIMARY KEY (genre_id),
	CONSTRAINT genres_un UNIQUE (genre_name)
);
CREATE TABLE IF NOT EXISTS public.studios (
	studio_id serial4 NOT NULL,
	studio_name varchar NULL,
	CONSTRAINT studios_pk PRIMARY KEY (studio_id),
	CONSTRAINT studios_un UNIQUE (studio_name)
);
CREATE TABLE IF NOT EXISTS public.themes (
	theme_id serial4 NOT NULL,
	theme_name varchar NULL,
	CONSTRAINT themes_pk PRIMARY KEY (theme_id),
	CONSTRAINT themes_un UNIQUE (theme_name)
);
CREATE TABLE IF NOT EXISTS public.anime_genres (
	anime_id int4 NOT NULL,
	genre_id int4 NOT NULL,
	CONSTRAINT anime_genres_fk FOREIGN KEY (anime_id) REFERENCES public.animes(anime_id),
	CONSTRAINT anime_genres_fk_1 FOREIGN KEY (genre_id) REFERENCES public.genres(genre_id),
	CONSTRAINT anime_genres_un UNIQUE (anime_id, genre_id)
);
CREATE TABLE IF NOT EXISTS public.anime_studios (
	anime_id int4 NOT NULL,
	studio_id int4 NOT NULL,
	CONSTRAINT anime_studios_fk FOREIGN KEY (anime_id) REFERENCES public.animes(anime_id),
	CONSTRAINT anime_studios_fk_1 FOREIGN KEY (studio_id) REFERENCES public.studios(studio_id),
	CONSTRAINT anime_studios_un UNIQUE (anime_id, studio_id)
);
CREATE TABLE IF NOT EXISTS public.anime_themes (
	anime_id int4 NOT NULL,
	theme_id int4 NOT NULL,
	CONSTRAINT anime_themes_fk FOREIGN KEY (anime_id) REFERENCES public.animes(anime_id),
	CONSTRAINT anime_themes_fk_1 FOREIGN KEY (theme_id) REFERENCES public.themes(theme_id),
	CONSTRAINT anime_themes_un UNIQUE (anime_id, theme_id)
);
CREATE TABLE IF NOT EXISTS public.anime_episodes (
	episode_id serial NOT NULL,
	episode_title varchar NULL,
	episode_stream_url varchar NULL,
	episode_index int NOT NULL,
	anime_id int NOT NULL,
	CONSTRAINT anime_episodes_pk PRIMARY KEY (episode_id),
	CONSTRAINT anime_episodes_un UNIQUE (anime_id,episode_index),
	CONSTRAINT anime_episodes_fk FOREIGN KEY (anime_id) REFERENCES public.animes(anime_id)
);
CREATE TABLE IF NOT EXISTS public.relations (
	relation_id serial NOT NULL,
	relation_name varchar NULL,
	CONSTRAINT relations_pk PRIMARY KEY (relation_id),
	CONSTRAINT relations_un UNIQUE (relation_name)
);
CREATE TABLE IF NOT EXISTS public.anime_relations (
	anime_id int NOT NULL,
	related_anime_id int NOT NULL,
	relation_id int NOT NULL,
	CONSTRAINT anime_relations_fk FOREIGN KEY (anime_id) REFERENCES public.animes(anime_id),
	CONSTRAINT anime_relations_fk_1 FOREIGN KEY (related_anime_id) REFERENCES public.animes(anime_id),
	CONSTRAINT anime_relations_fk_2 FOREIGN KEY (relation_id) REFERENCES public.relations(relation_id)
);
