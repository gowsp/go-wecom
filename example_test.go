package wecom_test

import (
	"context"
	"log"
	"time"

	"github.com/gowsp/go-wecom"
	"github.com/gowsp/go-wecom/aibot/longconn"
	"github.com/gowsp/go-wecom/callback"
	"github.com/gowsp/go-wecom/client"
	"github.com/gowsp/go-wecom/mail"
	"github.com/gowsp/go-wecom/media"
	"github.com/gowsp/go-wecom/webhook"
	"github.com/gowsp/go-wecom/wedoc"
)

// exampleContext keeps the examples short; a real program passes its own
// context and cancellation.
func exampleContext() context.Context { return context.Background() }

// ExampleNew wires every CGI client over one shared token cache.
func ExampleNew() {
	api := wecom.New("corp-id", "corp-secret", 1000002)

	ctx := exampleContext()
	if err := api.Message.SendText(ctx, "zhangsan", "hello"); err != nil {
		log.Print(err)
	}
	if err := api.KF.SendText(ctx, "wkAJ2GCAAASSm4", "wmxxxxxxxx", "您好"); err != nil {
		log.Print(err)
	}
}

// ExampleNew_customEndpoint shows how to point the transport at a test server.
func ExampleNew_customEndpoint() {
	api := wecom.New("corp-id", "corp-secret", 0,
		client.WithBaseURL("http://127.0.0.1:8080"),
	)
	_ = api.WeDoc
}

// ExampleMedia_Upload uploads a temporary material and reuses its media id.
func ExampleMedia_Upload() {
	api := wecom.New("corp-id", "corp-secret", 1000002)

	item, err := api.Media.Upload(exampleContext(), media.TypeFile, "report.pdf", []byte("%PDF-1.4"))
	if err != nil {
		log.Print(err)
		return
	}
	data, err := api.Media.Download(exampleContext(), item.MediaID)
	if err != nil {
		log.Print(err)
		return
	}
	_ = data
}

// ExampleMail_Send sends a plain mail to one employee and one address.
func ExampleMail_Send() {
	api := wecom.New("corp-id", "corp-secret", 1000002)

	err := api.Mail.Send(exampleContext(), mail.SendRequest{
		To:      mail.Address{UserIDs: []string{"zhangsan"}, Emails: []string{"someone@example.com"}},
		Subject: "周报",
		Content: "见附件",
	})
	if err != nil {
		log.Print(err)
	}
}

// ExampleWeDoc_AddRecords appends rows to a smart-sheet sub-table.
func ExampleWeDoc_AddRecords() {
	api := wecom.New("corp-id", "corp-secret", 1000002)

	_, err := api.WeDoc.AddRecords(exampleContext(), "DOCID", "SHEETID", wedoc.Record{
		Values: map[string]any{"FIELDID": wedoc.TextCell("hello")},
	})
	if err != nil {
		log.Print(err)
	}
}

// ExampleWebhook_Send pushes one text message to a group robot.
func ExampleWebhook_Send() {
	hook := webhook.New("693a91f6-7xxx")

	err := hook.Send(exampleContext(), webhook.Message{
		MsgType: "text",
		Text:    &webhook.Text{Content: "构建完成", MentionedList: []string{"@all"}},
	})
	if err != nil {
		log.Print(err)
	}
}

// ExampleCallbackCrypto_VerifyURL proves callback URL ownership.
func ExampleCallbackCrypto_VerifyURL() {
	crypto, err := callback.NewCallbackCrypto("token", "43-char-encoding-aes-key", "corp-id")
	if err != nil {
		log.Print(err)
		return
	}

	echo, err := crypto.VerifyURL("signature", "timestamp", "nonce", "echostr")
	if err != nil {
		log.Print(err)
		return
	}
	_, _ = echo, crypto.CorpID()
}

// ExampleBot_Start keeps a bot long connection alive until the context ends.
func ExampleBot_Start() {
	ctx, cancel := context.WithCancel(exampleContext())
	defer cancel()

	conn := longconn.New("bot-id", "bot-secret",
		longconn.WithHandler(func(ctx context.Context, event longconn.Event, reply *longconn.Responder) error {
			if event.Command == longconn.MessageCallback {
				return reply.ReplyStream(ctx, "已收到")
			}
			return nil
		}),
		longconn.WithPingInterval(30*time.Second),
	)
	if err := conn.Start(ctx); err != nil {
		log.Print(err)
	}
}
