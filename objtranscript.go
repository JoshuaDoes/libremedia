package main

type ObjectTranscript struct {
	RightToLeft bool `json:"rightToLeft,omitempty"`
	Provider string `json:"provider,omitempty"`
	ProviderLyricsID string `json:"providerLyricsId,omitempty"`
	ProviderTrackID string `json:"providerTrackId,omitempty"` //SyncLyricsURI
	TimeSynced bool `json:"timeSynced,omitempty"`
	Lines []*ObjectTranscriptLine `json:"lines,omitempty"`
}

type ObjectTranscriptLine struct {
	StartTimeMs int `json:"startTimeMs,omitempty"`
	Text string `json:"text,omitempty"`
}