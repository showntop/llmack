package speech

// TranscriptionPayload ...
type TranscriptionPayload struct {
	Text  string
	Final bool
}

// TurnPayload ...
type TurnPayload struct {
	TurnID        string
	Transcription *TranscriptionPayload
}

// AgentPayload ...
type AgentPayload struct {
	TurnID        string
	Transcription TranscriptionPayload
}

// ResponsePayload ...
type ResponsePayload struct {
	Text  string
	Chunk []byte
	First bool
	Final bool
}

// TTSResultPayload ...
type TTSResultPayload struct {
	Text   string
	Stream chan []byte
}
