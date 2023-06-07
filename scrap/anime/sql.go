package anime

import (
	"MalSql/scrap/anime/qb"
	_ "embed"
)

func appendOk(arr []string, elem ...string) []string {
	for _, e := range elem {
		if e != "" {
			arr = append(arr, e)
		}
	}
	return arr
}

//go:embed schema.sql
var Schema string

func animeSql(anime Anime) []string {
	var sql []string
	sql = appendOk(sql, qb.Insert("anime_types").Str("type_name", anime.filInf.typeOf).Sql())
	sql = appendOk(sql, qb.Insert("anime_seasons").Str("season_name", anime.filInf.season).Sql())
	q := qb.Insert("animes").Str("anime_description", anime.Description)
	q.Str("anime_title", anime.Title).Str("anime_title_en", anime.filInf.titleEn).Str("anime_title_jp", anime.filInf.titleJp)
	q.Str("anime_mal_url", anime.MalUrl).Str("anime_img_url", anime.ImgUrl)
	q.SubQ("type_id", `SELECT type_id FROM anime_types WHERE type_name = '%v'`, anime.filInf.typeOf)
	q.SubQ("season_id", `SELECT season_id FROM anime_seasons WHERE season_name = '%v'`, anime.filInf.season)
	q.Str("anime_start", getOrEmpty(anime.filInf.aired, 0)).Str("anime_end", getOrEmpty(anime.filInf.aired, 1))
	sql = appendOk(sql, q.Sql())
	for _, genre := range anime.filInf.genres {
		sql = appendOk(sql, qb.Insert("genres").Str("genre_name", genre).Sql())
		q := qb.Insert("anime_genres")
		q.SubQ("anime_id", `SELECT anime_id FROM animes WHERE anime_mal_url = '%v'`, anime.MalUrl)
		q.SubQ("genre_id", `SELECT genre_id FROM genres WHERE genre_name = '%v'`, genre)
		sql = appendOk(sql, q.Sql())
	}
	for _, theme := range anime.filInf.themes {
		sql = appendOk(sql, qb.Insert("themes").Str("theme_name", theme).Sql())
		q := qb.Insert("anime_themes")
		q.SubQ("anime_id", `SELECT anime_id FROM animes WHERE anime_mal_url = '%v'`, anime.MalUrl)
		q.SubQ("theme_id", `SELECT theme_id FROM themes WHERE theme_name = '%v'`, theme)
		sql = appendOk(sql, q.Sql())
	}
	for _, studio := range anime.filInf.studios {
		sql = appendOk(sql, qb.Insert("studios").Str("studio_name", studio).Sql())
		q := qb.Insert("anime_studios")
		q.SubQ("anime_id", `SELECT anime_id FROM animes WHERE anime_mal_url = '%v'`, anime.MalUrl)
		q.SubQ("studio_id", `SELECT studio_id FROM studios WHERE studio_name = '%v'`, studio)
		sql = appendOk(sql, q.Sql())
	}
	for i, episode := range anime.Episodes {
		q := qb.Insert("anime_episodes")
		q.Str("episode_title", episode.Title).Str("episode_stream_url", episode.Url)
		q.Int("episode_index", i)
		q.SubQ("anime_id", `SELECT anime_id FROM animes WHERE anime_mal_url = '%v'`, anime.MalUrl)
		sql = appendOk(sql, q.Sql())
	}
	return sql

}

func relationsSql(relation Relation) []string {
	var sql []string
	sql = appendOk(sql, qb.Insert("relations").Str("relation_name", relation.TypeOf).Sql())
	q := qb.Insert("anime_relations")
	q.SubQ("anime_id", `SELECT anime_id FROM animes WHERE anime_mal_url = '%v'`, relation.Root)
	q.SubQ("related_anime_id", `SELECT anime_id FROM animes WHERE anime_mal_url = '%v'`, relation.Related)
	q.SubQ("relation_id", `SELECT relation_id FROM relations WHERE relation_name = '%v'`, relation.TypeOf)
	sql = appendOk(sql, q.Sql())
	return sql
}
