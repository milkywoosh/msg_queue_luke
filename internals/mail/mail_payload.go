package mail

import "io"

// upload dari file multipart
type AttachFileDetail struct {
	File        io.Reader `json:"file,omitempty"`
	FileName    string
	ContentType string
}

func NewAttachFileDetail(
	File io.Reader,
	FileName string,
	ContentType string,
) AttachFileDetail {
	return AttachFileDetail{
		File,
		FileName,
		ContentType,
	}
}

type EmailPayload struct {
	Subject    string   `json:"subject"`
	Content    string   `json:"content"`
	To         []string `json:"to"`
	Cc         []string `json:"cc,omitempty"`
	Bcc        []string `json:"bcc,omitempty"`
	Attachment []AttachFileDetail
}

func NewEmailPayload(
	Subject string,
	Content string,
	To []string,
	Cc []string,
	Bcc []string,
	Attachment []AttachFileDetail,

) EmailPayload {
	return EmailPayload{
		Subject,
		Content,
		To,
		Cc,
		Bcc,
		Attachment,
	}
}
