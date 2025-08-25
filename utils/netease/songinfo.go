package netease

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	"github.com/go-musicfox/go-musicfox/utils/mathx"
	"github.com/go-musicfox/netease-music/service"
)

// PlayableInfo 歌曲的 URL 等信息
type PlayableInfo struct {
	URL       string
	MusicType string
	Size      int64
	Quality   service.SongQualityLevel
}

var brMap = map[service.SongQualityLevel]string{
	service.Standard: "320000",
	service.Higher:   "320000",
	service.Exhigh:   "320000",
	service.Lossless: "999000",
	service.Hires:    "999000",
}

// FetchPlayableInfo 从网易云API获取一首歌的可播放信息
func FetchPlayableInfo(songID int64, quality service.SongQualityLevel) (PlayableInfo, error) {
	urlService := service.SongUrlV1Service{
		ID:      strconv.FormatInt(songID, 10),
		Level:   quality,
		SkipUNM: true,
	}
	code, response := urlService.SongUrl()
	if code != 200 {
		return PlayableInfo{}, errors.New(string(response))
	}

	filex.WriteFile(response, "/tmp/xxx/song1.json")
	var (
		err1, err2    error
		freeTrialInfo jsonparser.ValueType
	)
	url, err1 := jsonparser.GetString(response, "data", "[0]", "url")
	_, freeTrialInfo, _, err2 = jsonparser.Get(response, "data", "[0]", "freeTrialInfo")
	if err1 != nil || err2 != nil || url == "" || (freeTrialInfo != jsonparser.NotExist && freeTrialInfo != jsonparser.Null) {
		br, ok := brMap[quality]
		if !ok {
			br = "320000"
		}
		s := service.SongUrlService{
			ID: strconv.FormatInt(songID, 10),
			Br: br,
		}
		code, response = s.SongUrl()
		if code != 200 {
			return PlayableInfo{}, errors.New(string(response))
		}
	}

	url, _ = jsonparser.GetString(response, "data", "[0]", "url")

	size, _ := jsonparser.GetInt(response, "data", "[0]", "size")
	if size > 0 {
		slog.Info("music size", "size", mathx.FormatBytes(size))
	}

	musicType, _ := jsonparser.GetString(response, "data", "[0]", "type")
	if musicType = strings.ToLower(musicType); musicType == "" {
		musicType = "mp3"
	}

	ok, err := ShouldInterceptURL(url, bannedLinkFeatures)

	if err != nil {
		return PlayableInfo{}, err
	}
	if ok {
		return PlayableInfo{}, errors.New(fmt.Sprintf("检测的无效的播放链接: %v", url))
	}

	return PlayableInfo{
		URL:       url,
		MusicType: musicType,
		Size:      size,
		Quality:   quality,
	}, nil
}

var bannedLinkFeatures = []string{
	"/resource/n2/73/84/3759149332.mp3",
}

// ShouldInterceptURL 检查给定的 URL 是否匹配任何一个不想要的特征。
// urlToCheck: 需要被检测的链接字符串。
// unwantedFeatures: 一个包含所有不想要的 URL 路径后缀的列表。
//
// 返回值:
//   - bool: 如果链接应该被拦截，则为 true。
//   - error: 如果 urlToCheck 不是一个有效的 URL，则返回错误。
func ShouldInterceptURL(urlToCheck string, unwantedFeatures []string) (bool, error) {
	parsedURL, err := url.Parse(urlToCheck)
	if err != nil {
		return false, fmt.Errorf("无法解析 URL: %w", err)
	}

	for _, feature := range unwantedFeatures {
			if strings.HasSuffix(parsedURL.Path, feature) {
					slog.Warn("无效的酷我链接，已跳过播放")
			return true, nil
		}
	}

	return false, nil
}
