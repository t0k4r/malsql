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
	sql = appendOk(sql, qb.Insert("anime_types").Str("name_of", anime.filInf.typeOf).Sql())
	sql = appendOk(sql, qb.Insert("seasons").Str("name_of", anime.filInf.season).Str("value", anime.filInf.seasonDate).Sql())
	sql = append(sql, qb.Insert("animes").Str("description", anime.Description).
		Str("title", anime.Title).Str("title_en", anime.filInf.titleEn).Str("title_jp", anime.filInf.titleJp).
		Str("mal_url", anime.MalUrl).Str("cover_url", anime.ImgUrl).
		SubQ("type_id", `SELECT t.id FROM anime_types t WHERE t.name_of = '%v'`, anime.filInf.typeOf).
		SubQ("season_id", `SELECT s.id FROM seasons s WHERE s.name_of = '%v'`, anime.filInf.season).
		Str("aired_from", getOrEmpty(anime.filInf.aired, 0)).Str("aired_to", getOrEmpty(anime.filInf.aired, 1)).Sql())
	for i, episode := range anime.Episodes {
		q := qb.Insert("episodes").
			Str("title", episode.Title).Str("stream_url", episode.Url).Int("index_of", i).
			SubQ("anime_id", `SELECT a.id FROM animes a WHERE a.mal_url = '%v'`, anime.MalUrl)
		sql = appendOk(sql, q.Sql())
	}
	for _, info := range anime.Information {
		sql = append(sql, qb.Insert("info_types").Str("name_of", info.Key).Sql())
		sql = append(sql, qb.Insert("infos").Str("value", info.Value).
			SubQ("type_id", "SELECT t.id FROM info_types t WHERE t.name_of = '%v'", info.Key).Sql())
		sql = append(sql, qb.Insert("anime_infos").
			SubQ("info_id", "SELECT id FROM infos WHERE value = '%v'", info.Value).
			SubQ("anime_id", `SELECT id FROM animes WHERE mal_url = '%v'`, anime.MalUrl).Sql())
	}
	return sql

}

func relationsSql(relation Relation) []string {
	var sql []string
	sql = appendOk(sql, qb.Insert("relations").Str("name_of", relation.TypeOf).Sql())
	q := qb.Insert("anime_relations")
	q.SubQ("anime_id", `SELECT id FROM animes WHERE mal_url = '%v'`, relation.Root)
	q.SubQ("related_id", `SELECT id FROM animes WHERE mal_url = '%v'`, relation.Related)
	q.SubQ("relation_id", `SELECT id FROM relations WHERE name_of = '%v'`, relation.TypeOf)
	sql = appendOk(sql, q.Sql())
	return sql
}
