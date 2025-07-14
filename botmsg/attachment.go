package botmsg

// Attachment to a bot message
type Attachment interface {
	AttachmentType() AttachmentType
}

// AttachmentType to a bot message
type AttachmentType int

//goland:noinspection GoUnusedConst
const (
	// AttachmentTypeNone says there is no attachment
	AttachmentTypeNone AttachmentType = iota

	// AttachmentTypeAudio is for audio attachments
	AttachmentTypeAudio

	// AttachmentTypeFile is for file attachments
	AttachmentTypeFile

	// AttachmentTypeImage is for image attachments
	AttachmentTypeImage

	// AttachmentTypeVideo is for video attachments
	AttachmentTypeVideo
)
