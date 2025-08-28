package structs

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type Artist struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func NewArtist(json []byte) (Artist, error) {
	var artist Artist

	if len(json) == 0 {
		return artist, errors.New("json is empty")
	}

	arId, err := jsonparser.GetInt(json, "id")
	if err != nil {
		return artist, err
	}
	artist.Id = arId

	if arName, err := jsonparser.GetString(json, "name"); err == nil {
		artist.Name = arName
	}

	return artist, nil
}

// NewArtists 从 josn 数组获取所有艺术家
func NewArtists(json []byte, keys ...string) ([]Artist, error) {
	var artists []Artist
	if len(json) == 0 {
		return artists, errors.New("json is empty")
	}
	_, err := jsonparser.ArrayEach(json, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		artist, err := NewArtist(value)

		if err == nil {
			artists = append(artists, artist)
		}
	}, keys...)
	return artists, err
}
