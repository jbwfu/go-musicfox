package structs

type SongSource int

const (
	SourceUnknown        SongSource = 0 // 零值表示未知或未指定
	SourceUser           SongSource = 1
	SourceArtist         SongSource = 2
	SourcePlayList       SongSource = 13 // 歌单
	SourceDjRadio        SongSource = 17 // 电台
	SourceSong           SongSource = 18 // 单曲
	SourceAlbum          SongSource = 19
	SourceMV             SongSource = 21
	SourceDailyRecommend SongSource = 24 // 每日推荐
	SourceToplist        SongSource = 31 // 排行榜
	SourceSearch         SongSource = 32
	SourceSearchLegacy   SongSource = 33 // 同样是 search
	SourceEvent          SongSource = 34 // 动态
	SourceMsg            SongSource = 35 // 消息
	SourceUserLegacy     SongSource = 50 // 同样是 user
)

var songSourceToString = map[SongSource]string{
	SourceUser:           "user",
	SourceArtist:         "artist",
	SourcePlayList:       "list",
	SourceDjRadio:        "dj",
	SourceSong:           "song",
	SourceAlbum:          "album",
	SourceMV:             "mv",
	SourceDailyRecommend: "dailySongRecommend",
	SourceToplist:        "toplist",
	SourceSearch:         "search",
	SourceSearchLegacy:   "search",
	SourceEvent:          "event",
	SourceMsg:            "msg",
	SourceUserLegacy:     "user",
}

func (s SongSource) String() string {
	if str, ok := songSourceToString[s]; ok {
		return str
	}
	return ""
}

type SourceInfo struct {
	Type     SongSource `json:"type"`               // 来源类型，用于上报
	Id       int64      `json:"id"`                 // 来源ID
	Playlist *Playlist  `json:"playlist,omitempty"` // 来自哪个歌单
	DjRadio  *DjRadio   `json:"djRadio,omitempty"`  // 来自哪个电台
}

func NewSourceInfo(sourceEntity any) *SourceInfo {
	if sourceEntity == nil {
		return nil
	}

	switch v := sourceEntity.(type) {
	case Playlist:
		return &SourceInfo{
			Type:     SourcePlayList,
			Id:       v.Id,
			Playlist: &v,
		}
	case DjRadio:
		return &SourceInfo{
			Type:    SourceDjRadio,
			Id:      v.Id,
			DjRadio: &v,
		}
	default:
		return nil
	}
}
