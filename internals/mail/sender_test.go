package mail

// func TestSendEmailWithGmail(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip()
// 	}

// 	config, err := utils.LoadConfig("../..")
// 	require.NoError(t, err)

// 	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

// 	subject := "A test email"
// 	content := `
// 	<h1>Hello world</h1>
// 	<p>This is a test message from <a href="https://github.com/milkywoosh">Lukman github</a></p>
// 	`
// 	to := []string{"destination.email@gmail.com"}
// 	attachFiles := []string{"../../note.txt", "../../readme.md"}

// 	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
// 	require.NoError(t, err)
// }
