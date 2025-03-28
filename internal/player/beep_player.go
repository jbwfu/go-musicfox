package player

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/iox"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/storagex"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

const (
	sampleRate       = beep.SampleRate(44100)
	resampleQuiality = 4
)

type beepPlayer struct {
	l sync.Mutex

	curMusic URLMusic
	timer    *timex.Timer

	cacheReader     *os.File
	cacheWriter     *os.File
	cacheDownloaded bool

	curStreamer beep.StreamSeekCloser
	curFormat   beep.Format

	state      types.State
	ctrl       *beep.Ctrl
	volume     *effects.Volume
	timeChan   chan time.Duration
	stateChan  chan types.State
	musicChan  chan URLMusic
	httpClient *http.Client

	close chan struct{}
}

func NewBeepPlayer() *beepPlayer {
	p := &beepPlayer{
		state: types.Stopped,

		timeChan:  make(chan time.Duration),
		stateChan: make(chan types.State),
		musicChan: make(chan URLMusic),
		ctrl: &beep.Ctrl{
			Paused: false,
		},
		volume: &effects.Volume{
			Base:   2,
			Silent: false,
		},
		httpClient: &http.Client{},
		close:      make(chan struct{}),
	}

	errorx.WaitGoStart(p.listen)

	return p
}

// listen 开始监听
func (p *beepPlayer) listen() {
	var (
		done       = make(chan struct{})
		err        error
		ctx        context.Context
		cancel     context.CancelFunc
		doneHandle = func() {
			select {
			case done <- struct{}{}:
			case <-p.close:
			}
		}
		waitBytes = func() int {
			if p.curMusic.Type != Flac {
				return 512
			}
			return 512 * 4
		}
	)

	if err = speaker.Init(sampleRate, sampleRate.N(time.Millisecond*200)); err != nil {
		panic(err)
	}

	for {
		select {
		case <-p.close:
			if cancel != nil {
				cancel()
			}
			return
		case <-done:
			p.Stop()
		case p.curMusic = <-p.musicChan:
			p.l.Lock()
			p.pausedNoLock()
			// 清理上一轮
			if cancel != nil {
				cancel()
			}
			p.reset()

			ctx, cancel = context.WithCancel(context.Background())

			if strings.HasPrefix(p.curMusic.URL, "file://") {
				p.cacheDownloaded = true
			} else {
				p.curMusic.URL, _, _ = storagex.GetCacheURL(p.curMusic.Id)
			}
			cacheFile := strings.TrimPrefix(p.curMusic.URL, "file://")

			// FIXME: 边听边存不可用
			// 由于下载使用了临时文件，此处实际仍然是等待完整下载
			if p.awaitCache(ctx, cacheFile, true) {
				if p.cacheReader, err = os.OpenFile(cacheFile, os.O_RDONLY, 0666); err != nil {
					panic(err)
				}

			} else {
				p.stopNoLock()
				goto nextLoop
			}

			if !p.cacheDownloaded {
				// 边下载边播放
				go func(ctx context.Context) {
					defer func() {
						if errorx.Recover(true) {
							p.Stop()
						}
					}()

					if !p.awaitCache(ctx, cacheFile, false) {
						return
					}

					if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
						return
					}

					p.l.Lock()
					defer p.l.Unlock()

					// 需再开一次文件，保证其指针变化，否则将概率导致 p.ctrl.Streamer = beep.Seq(……) 直接停止播放
					if !p.loadOrReload(doneHandle) {
						return
					}
					p.cacheDownloaded = true
				}(ctx)

				if err = iox.WaitForNBytes(p.cacheReader, waitBytes(), time.Millisecond*100, 50); err != nil {
					slog.Error("WaitForNBytes err", slogx.Error(err))
					p.stopNoLock()
					goto nextLoop
				}
			}

			if !p.loadOrReload(doneHandle) {
				goto nextLoop
			}

			slog.Info("current song sample rate", slog.Int("sample_rate", int(p.curFormat.SampleRate)))

			p.volume.Streamer = p.ctrl
			speaker.Play(p.volume)

			// 计时器
			p.timer = timex.NewTimer(timex.Options{
				Duration:       8760 * time.Hour,
				TickerInternal: 200 * time.Millisecond,
				OnRun:          func(started bool) {},
				OnPause:        func() {},
				OnDone:         func(stopped bool) {},
				OnTick: func() {
					select {
					case p.timeChan <- p.timer.Passed():
					default:
					}
				},
			})
			p.resumeNoLock()

		nextLoop:
			p.l.Unlock()
		}
	}
}

// onlystart 为 true 时检测缓存文件的开始状态， false 时检测完成状态
func (p *beepPlayer) awaitCache(ctx context.Context, file string, onlystart bool) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if storagex.IsCaching() {
				if !onlystart {
					continue // 等待缓存操作结束
				}
				if _, err := os.Stat(file); err == nil {
					return true // 缓存文件已存在，继续后续操作
				}
				continue
			}
			if _, err := os.Stat(file); err == nil {
				return true
			}
			return false
		}
	}
}

// 加载或重载音乐文件至 beep
func (p *beepPlayer) loadOrReload(callback func()) bool {
	var (
		reader *os.File
		pos    int
		err    error
	)

	if p.curStreamer == nil {
		reader = p.cacheReader
		pos = 0

	} else {
		// 除了MP3格式，其他格式无需重载
		if p.curMusic.Type == Mp3 && configs.ConfigRegistry.Player.BeepMp3Decoder != types.BeepMiniMp3Decoder {
			return true
		}
		reader, _ = os.OpenFile(strings.TrimPrefix(p.curMusic.URL, "file://"), os.O_RDONLY, 0666)

		lastStreamer := p.curStreamer
		defer func() { _ = lastStreamer.Close() }()
		pos = lastStreamer.Position()
	}

	if p.curStreamer, p.curFormat, err = DecodeSong(p.curMusic.Type, reader); err != nil {
		slog.Error(fmt.Sprintf("解码错误：%v", err))
		p.stopNoLock()
		return false
	}
	if pos >= p.curStreamer.Len() && p.curStreamer.Len() != 0 {
		pos = p.curStreamer.Len() - 1
	}
	_ = p.curStreamer.Seek(pos)
	p.ctrl.Streamer = beep.Seq(p.resampleStreamer(p.curFormat.SampleRate), beep.Callback(callback))

	return true
}

// Play 播放音乐
func (p *beepPlayer) Play(music URLMusic) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case p.musicChan <- music:
	case <-timer.C:
	}
}

func (p *beepPlayer) CurMusic() URLMusic {
	return p.curMusic
}

func (p *beepPlayer) setState(state types.State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

// State 当前状态
func (p *beepPlayer) State() types.State {
	return p.state
}

// StateChan 状态发生变更
func (p *beepPlayer) StateChan() <-chan types.State {
	return p.stateChan
}

func (p *beepPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.Passed()
}

// TimeChan 获取定时器
func (p *beepPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *beepPlayer) Seek(duration time.Duration) {
	if duration < 0 {
		return
	}
	// FIXME: 暂时仅对MP3格式提供跳转功能
	// FLAC格式(其他未测)跳转会占用大量CPU资源，比特率越高占用越高
	// 导致Seek方法卡住20-40秒的时间，之后方可随意跳转
	// minimp3未实现Seek
	if p.curStreamer == nil || p.curMusic.Type != Mp3 || configs.ConfigRegistry.Player.BeepMp3Decoder == types.BeepMiniMp3Decoder {
		return
	}
	if p.state == types.Playing || p.state == types.Paused {
		speaker.Lock()
		newPos := sampleRate.N(duration)

		if newPos < 0 {
			newPos = 0
		}
		if newPos >= p.curStreamer.Len() {
			newPos = p.curStreamer.Len() - 1
		}
		if p.curStreamer != nil {
			err := p.curStreamer.Seek(newPos)
			if err != nil {
				slog.Error("seek error", slogx.Error(err))
			}
		}
		if p.timer != nil {
			p.timer.SetPassed(duration)
		}
		speaker.Unlock()
	}
}

// UpVolume 调大音量
func (p *beepPlayer) UpVolume() {
	if p.volume.Volume >= 0 {
		return
	}
	p.l.Lock()
	defer p.l.Unlock()

	p.volume.Silent = false
	p.volume.Volume += 0.25
}

// DownVolume 调小音量
func (p *beepPlayer) DownVolume() {
	if p.volume.Volume <= -5 {
		return
	}

	p.l.Lock()
	defer p.l.Unlock()

	p.volume.Volume -= 0.25
	if p.volume.Volume <= -5 {
		p.volume.Silent = true
	}
}

func (p *beepPlayer) Volume() int {
	return int((p.volume.Volume + 5) * 100 / 5) // 转为0~100存储
}

func (p *beepPlayer) SetVolume(volume int) {
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}

	p.l.Lock()
	defer p.l.Unlock()
	p.volume.Volume = float64(volume)*5/100 - 5
}

func (p *beepPlayer) pausedNoLock() {
	if p.state != types.Playing {
		return
	}
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(types.Paused)
}

// Pause 暂停播放
func (p *beepPlayer) Pause() {
	p.l.Lock()
	defer p.l.Unlock()
	p.pausedNoLock()
}

func (p *beepPlayer) resumeNoLock() {
	if p.state == types.Playing {
		return
	}
	p.ctrl.Paused = false
	go p.timer.Run()
	p.setState(types.Playing)
}

// Resume 继续播放
func (p *beepPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	p.resumeNoLock()
}

func (p *beepPlayer) stopNoLock() {
	if p.state == types.Stopped {
		return
	}
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(types.Stopped)
}

// Stop 停止
func (p *beepPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	p.stopNoLock()
}

// Toggle 切换状态
func (p *beepPlayer) Toggle() {
	switch p.State() {
	case types.Paused, types.Stopped:
		p.Resume()
	case types.Playing:
		p.Pause()
	default:
		p.Resume()
	}
}

// Close 关闭
func (p *beepPlayer) Close() {
	p.l.Lock()
	defer p.l.Unlock()

	if p.timer != nil {
		p.timer.Stop()
	}
	close(p.close)
	speaker.Clear()
	speaker.Close()
}

func (p *beepPlayer) reset() {
	// 关闭旧计时器
	if p.timer != nil {
		p.timer.SetPassed(0)
		p.timer.Stop()
	}
	if p.cacheReader != nil {
		_ = p.cacheReader.Close()
	}
	if p.cacheWriter != nil {
		_ = p.cacheWriter.Close()
	}
	if p.curStreamer != nil {
		_ = p.curStreamer.Close()
		p.curStreamer = nil
	}
	p.cacheDownloaded = false
	speaker.Clear()
}

func (p *beepPlayer) streamer(samples [][2]float64) (n int, ok bool) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("streamer panic", slogx.Error(err))
			p.Stop()
		}
	}()
	pos := p.curStreamer.Position()
	n, ok = p.curStreamer.Stream(samples)
	err := p.curStreamer.Err()
	if err == nil && (ok || p.cacheDownloaded) {
		return
	}
	p.pausedNoLock()

	retry := 4
	for !ok && retry > 0 {
		if p.curMusic.Type == Flac {
			if err = p.curStreamer.Seek(pos); err != nil {
				return
			}
		}
		errorx.ResetError(p.curStreamer)

		select {
		case <-time.After(time.Second * 5):
			n, ok = p.curStreamer.Stream(samples)
		case <-p.close:
			return
		}
		retry--
	}
	p.resumeNoLock()
	return
}

func (p *beepPlayer) resampleStreamer(old beep.SampleRate) beep.Streamer {
	if old == sampleRate {
		return beep.StreamerFunc(p.streamer)
	}
	return beep.Resample(resampleQuiality, old, sampleRate, beep.StreamerFunc(p.streamer))
}
