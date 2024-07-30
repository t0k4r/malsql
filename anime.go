package malsql

import (
	"fmt"
	"malsql/mal"
	"strings"
)

type Anime struct{ mal.Anime }

func (a *Anime) Sql() ([]string, []string) {
	var animeSql []string
	var relationsSql []string
	typeName, typOk := a.Type()
	if typOk {
		animeSql = append(animeSql,
			fmt.Sprintf("insert into anime_types (name) values ('%v') on conflict do nothing", typeName))
	}
	seasonName, seasonDate, seasonOk := a.Season()
	if seasonOk {
		animeSql = append(animeSql,
			fmt.Sprintf("insert into seasons (name,value) values ('%v','%v') on conflict do nothing",
				seasonName, seasonDate.Format("2006-01-02")))
	}
	animeSql = append(animeSql,
		fmt.Sprintf("insert into animes (title,mal_url,img_url,description,type_id,season_id) values ('%v','%v','%v','%v',%v,%v) on conflict do nothing",
			a.Title, a.MalUrl, a.ImgUrl, strings.ReplaceAll(a.Description, "'", "''"),
			func() string {
				if typOk {
					return fmt.Sprintf("(select id from types where name='%v')", typeName)
				}
				return "null"
			}(),
			func() string {
				if seasonOk {
					return fmt.Sprintf("(select id from seasons where name='%v')", seasonName)
				}
				return "null"
			}()))
	animeSubQ := fmt.Sprintf("select id from animes where mal_url='%v'", a.MalUrl)
	for altTitleType, altTitle := range a.TitleAlt() {
		animeSql = append(animeSql, fmt.Sprintf("insert into alt_title_types (name) values ('%v') on conflict do nothing", altTitleType))
		animeSql = append(animeSql,
			fmt.Sprintf("insert into anime_alt_titles values (anime_id,type_id,alt_title) ((%v), (select id from alt_title_types where name='%v'),'%v')",
				animeSubQ, altTitleType, altTitle))
	}
	for infoType, infos := range a.Infos() {
		animeSql = append(animeSql,
			fmt.Sprintf("insert info_types (name) values ('%v') on conflict do nothing", infoType))
		for _, info := range infos {
			animeSql = append(animeSql,
				fmt.Sprintf("insert into anime_infos (anime_id,type_id,info) values ((%v),(select id from info_types where value='%v'),'%v') on conflict do nothing",
					animeSubQ, infoType, info))
		}
	}
	for i, episode := range a.Episodes {
		animeSql = append(animeSql,
			fmt.Sprintf("insert into anime_episodes (anime_id,index,title,alt_title,mal_url) values ((%v),%v,'%v','%v','%v') on conflict do nothing",
				animeSubQ, i, episode.Title, episode.TitleAlt, episode.Url))
	}

	for relationType, relations := range a.Related {
		relationsSql = append(relationsSql,
			"insert into relation_types (name) value ('"+relationType+"') on conflict do nothing")
		for _, relation := range relations {
			relationsSql = append(relationsSql,
				fmt.Sprintf("insert into anime_relations (anime_id, related_anime_id, relation_id) values (%v,%v,%v) on conflict do nothing",
					"(select id from animes where name='"+a.MalUrl+"')",
					"(select if from anime_relations where name='"+relation+"')",
					"(select if from animes where name='"+relationType+"')"))
		}
	}
	return animeSql, relationsSql
}
