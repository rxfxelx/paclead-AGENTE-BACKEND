package flow

import (
    "context"

    "pac-lead-agent/internal/clients"
)

func SendAssistantReplyAudio(ctx context.Context, ai *clients.OpenAI, whats *clients.Whats, threadID, number string) error {
    reply, err := ai.LastMessageText(ctx, threadID)
    if err != nil || reply == "" { return err }
    b64, err := ai.TextToSpeech(ctx, reply)
    if err != nil { return err }
    return whats.SendAudioBase64(ctx, number, b64)
}
