package main

import (
	"github.com/go-music-players/mpris"
	"github.com/issadarkthing/gomu/player"
	"github.com/tramhao/id3v2"
)

// MPRISAdapter adapts gomu's Player to the mpris.Player interface
type MPRISAdapter struct {
	gomu *Gomu
}

// NewMPRISAdapter creates a new MPRIS adapter for gomu
func NewMPRISAdapter(g *Gomu) *MPRISAdapter {
	return &MPRISAdapter{
		gomu: g,
	}
}

// Play starts or resumes playback
func (a *MPRISAdapter) Play() error {
	if a.gomu.player.IsPaused() {
		a.gomu.player.Play()
	}
	return nil
}

// Pause pauses playback
func (a *MPRISAdapter) Pause() error {
	if !a.gomu.player.IsPaused() && a.gomu.player.IsRunning() {
		a.gomu.player.Pause()
	}
	return nil
}

// Stop stops playback
func (a *MPRISAdapter) Stop() error {
	// gomu doesn't have explicit stop, use skip to clear current track
	if a.gomu.player.IsRunning() {
		a.gomu.player.Skip()
	}
	return nil
}

// Next plays next track in queue
func (a *MPRISAdapter) Next() error {
	if a.gomu.player.IsRunning() {
		a.gomu.player.Skip()
	}
	return nil
}

// Previous plays previous track (not implemented in gomu)
func (a *MPRISAdapter) Previous() error {
	// gomu doesn't support previous
	return nil
}

// GetPlaybackStatus returns current playback status
func (a *MPRISAdapter) GetPlaybackStatus() (mpris.PlaybackStatus, error) {
	if a.gomu.player.IsRunning() {
		if a.gomu.player.IsPaused() {
			return mpris.StatusPaused, nil
		}
		return mpris.StatusPlaying, nil
	}
	return mpris.StatusStopped, nil
}

// GetMetadata returns track metadata from ID3v2 tags
func (a *MPRISAdapter) GetMetadata() (*mpris.Metadata, error) {
	currentSong := a.gomu.player.GetCurrentSong()
	if currentSong == nil {
		return nil, nil
	}

	audioFile, ok := currentSong.(*player.AudioFile)
	if !ok {
		return nil, nil
	}

	metadata := &mpris.Metadata{
		TrackID: audioFile.Path(),
		Title:   audioFile.Name(),
	}

	// Try to read ID3v2 tags for richer metadata
	tag, err := id3v2.Open(audioFile.Path(), id3v2.Options{Parse: true})
	if err == nil {
		defer tag.Close()

		// Use ID3 title if available
		if title := tag.Title(); title != "" {
			metadata.Title = title
		}

		// Use ID3 artist if available
		if artist := tag.Artist(); artist != "" {
			metadata.Artist = []string{artist}
		}

		// Use ID3 album if available
		if album := tag.Album(); album != "" {
			metadata.Album = album
		}

		// Get album art from ID3 tags
		pictures := tag.GetFrames(tag.CommonID("Attached picture"))
		if len(pictures) > 0 {
			if pic, ok := pictures[0].(id3v2.PictureFrame); ok {
				// Album art is available in pic.Picture
				// The mpris library will need to handle artwork URLs
				// For now, we'll leave ArtURL empty as it requires serving the image
				_ = pic
			}
		}
	}

	// Set track length if available
	if length := audioFile.Len(); length > 0 {
		metadata.Duration = length
	}

	return metadata, nil
}

// CanPlay returns true if can play
func (a *MPRISAdapter) CanPlay() bool {
	return true
}

// CanPause returns true if can pause
func (a *MPRISAdapter) CanPause() bool {
	return true
}

// CanGoNext returns true if can go to next track
func (a *MPRISAdapter) CanGoNext() bool {
	return true // gomu supports skip/next
}

// CanGoPrevious returns true if can go to previous track
func (a *MPRISAdapter) CanGoPrevious() bool {
	return false // gomu doesn't support previous
}

// CanSeek returns true if can seek
func (a *MPRISAdapter) CanSeek() bool {
	return true // gomu has Seek() method
}

// CanControl returns true if player can be controlled
func (a *MPRISAdapter) CanControl() bool {
	return true
}
